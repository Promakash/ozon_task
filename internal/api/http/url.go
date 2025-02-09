package http

import (
	"context"
	"errors"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"log/slog"
	"net/http"
	"ozon_task/domain"
	"ozon_task/internal/api/http/types"
	"ozon_task/internal/usecases"
	pkgerr "ozon_task/pkg/error"
	"ozon_task/pkg/http/handlers"
	resp "ozon_task/pkg/http/responses"
	pkglog "ozon_task/pkg/log"
	"time"
)

type URLHandler struct {
	logger          *slog.Logger
	service         usecases.URL
	responseTimeout time.Duration
}

func NewURLHandler(logger *slog.Logger, service usecases.URL, responseTimeout time.Duration) *URLHandler {
	return &URLHandler{
		logger:          logger,
		service:         service,
		responseTimeout: responseTimeout,
	}
}

const postShortPath = "/urls"
const getOriginalPath = "/urls/{shortened}"

func (h *URLHandler) WithURLHandlers() handlers.RouterOption {
	return func(r chi.Router) {
		handlers.AddHandler(r.Post, postShortPath, h.postShortURL)
		handlers.AddHandler(r.Get, getOriginalPath, h.getOriginalURL)
	}
}

// @Summary		Create a shortened URL
// @Description Accepts a JSON payload containing the original URL and returns a generated shortened URL.
// @Description
// @Description The provided `original_url` must be a valid, publicly accessible URL.
// @Description - If the URL does not include an HTTP scheme (`http://` or `https://`), the service will automatically prepend `https://`.
// @Description - If the provided URL results in more than **10 redirects**, the response will contain the URL state at the **10th redirect**.
// @Description
// @Description If a shortened URL already exists for the given original URL, the existing shortened URL will be returned.
//
// @Accept			json
// @Produce		json
// @Param			original_url	body		types.PostShortURLRequest	true	"Original URL (must be publicly accessible; if no HTTP scheme is provided, `https://` is added automatically; URLs with more than 10 redirects return the last reachable state)."
// @Success		200				{object}	types.PostShortURLResponse	"Successfully created or retrieved an existing shortened URL"
// @Failure		400				{object}	responses.ErrorResponse		"Invalid request: the provided URL is malformed, inaccessible, or empty"
// @Failure		408				{object}	responses.ErrorResponse		"Request timeout: exceeded server execution time or client disconnected"
// @Failure		500				{object}	responses.ErrorResponse		"Internal service error"
// @Router			/urls [post]
func (h *URLHandler) postShortURL(r *http.Request) resp.Response {
	const op = "URLHandler.postShortURL"
	log := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, err := types.CreatePostShorURLRequest(r)
	if err != nil {
		log.Error("error while processing request", pkglog.Err(err))
		return h.handleResult(err, nil)
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.responseTimeout)
	defer cancel()

	shortened, err := h.service.ShortenURL(ctx, req.OriginalURL)
	if err != nil {
		log.Error("failed to generate shortened url", pkglog.Err(err))
	}

	return h.handleResult(err, shortened)
}

// @Summary		Retrieve the original URL
// @Description	Given a shortened URL, returns the corresponding original URL.
// @Description
// @Description The `shortened` URL must be exactly **10 characters long** and consist only of:
// @Description - Uppercase and lowercase English letters (`A-Z, a-z`)
// @Description - Digits (`0-9`)
// @Description - Underscore (`_`)
//
// @Produce		json
// @Param			shortened	path		string							true	"Shortened URL (must be 10 characters long and follow the defined character set)"
// @Success		200			{object}	types.GetOriginalURLResponse	"Successfully retrieved the original URL"
// @Failure		400			{object}	responses.ErrorResponse			"Invalid format: incorrect length or invalid characters in the shortened URL"
// @Failure		404			{object}	responses.ErrorResponse			"Shortened URL not found in the system"
// @Failure		408			{object}	responses.ErrorResponse			"Request timeout: exceeded server execution time or client disconnected"
// @Failure		500			{object}	responses.ErrorResponse			"Internal service error"
// @Router			/urls/{shortened} [get]
func (h *URLHandler) getOriginalURL(r *http.Request) resp.Response {
	const op = "URLHandler.getOriginalURL"
	log := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, err := types.CreateGetOriginalURLRequest(r)
	if err != nil {
		log.Error("error while processing request", pkglog.Err(err))
		return h.handleResult(err, nil)
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.responseTimeout)
	defer cancel()

	original, err := h.service.ResolveURL(ctx, req.ShortenedURL)
	if err != nil {
		log.Error("failed to get original url", pkglog.Err(err))
	}

	return h.handleResult(err, original)
}

func (h *URLHandler) handleResult(err error, r any) resp.Response {
	if err == nil {
		return resp.OK(r)
	}

	err = pkgerr.UnwrapAll(err)

	switch {
	case errors.Is(err, domain.ErrInvalidShortened),
		errors.Is(err, domain.ErrInvalidOriginal),
		errors.Is(err, domain.ErrInaccessibleOriginal):
		return resp.BadRequest(err)
	case errors.Is(err, domain.ErrShortenedNotFound),
		errors.Is(err, domain.ErrOriginalNotFound):
		return resp.NotFound(err)
	case errors.Is(err, context.DeadlineExceeded),
		errors.Is(err, context.Canceled):
		return resp.RequestTimeout(err)
	default:
		return resp.Unknown(err)
	}
}
