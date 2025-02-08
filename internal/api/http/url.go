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

// @Summary		Create shortened URL
// @Description	Accepts a JSON payload containing the original URL and returns a generated shortened URL.
//
// If a shortened URL for the given original URL already exists, the existing shortened URL is returned.
//
// @Accept			json
// @Produce		json
// @Param			original_url	body		types.PostShortURLRequest	true	"JSON payload with the original URL"
// @Success		200				{object}	types.PostShortURLResponse	"Shortened URL successfully created or retrieved"
// @Failure		400				{object}	responses.ErrorResponse		"Bad request"
// @Failure		408				{object}	responses.ErrorResponse		"Request timeout (e.g. user disconnect or service timeout)"
// @Failure		500				{object}	responses.ErrorResponse		"Internal server error"
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

// @Summary		Retrieve original URL
// @Description	Given a shortened URL, returns the original URL.
// @Produce		json
// @Param			shortened	path		string							true	"Shortened URL (must be exactly 10 characters long and consist only of uppercase and lowercase English letters, digits, and underscore)"
// @Success		200			{object}	types.GetOriginalURLResponse	"Original URL successfully retrieved"
// @Failure		400			{object}	responses.ErrorResponse			"Bad request"
// @Failure		404			{object}	responses.ErrorResponse			"Shortened URL not found"
// @Failure		408			{object}	responses.ErrorResponse			"Request timeout (e.g. user disconnect or service timeout)"
// @Failure		500			{object}	responses.ErrorResponse			"Internal server error"
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
	switch {
	case err == nil:
		return resp.OK(r)
	case errors.Is(err, domain.ErrInvalidShortened) || errors.Is(err, domain.ErrInvalidOriginal):
		return resp.BadRequest(err)
	case errors.Is(err, domain.ErrShortenedNotFound) || errors.Is(err, domain.ErrOriginalNotFound):
		return resp.NotFound(err)
	case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
		return resp.RequestTimeout(err)
	default:
		return resp.Unknown(err)
	}
}
