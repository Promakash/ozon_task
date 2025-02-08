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

func NewURLHandler(logger *slog.Logger, service usecases.URL) *URLHandler {
	return &URLHandler{
		logger:  logger,
		service: service,
	}
}

func (h *URLHandler) WithURLHandlers() handlers.RouterOption {
	return func(r chi.Router) {
		handlers.AddHandler(r.Post, "/urls", h.postShortURL)
		handlers.AddHandler(r.Get, "/urls/{shortened}", h.getOriginalURL)
	}
}

func (h *URLHandler) postShortURL(r *http.Request) resp.Response {
	const op = "URLHandler.postShortURL"
	log := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, err := types.CreatePostShorURLRequest(r)
	if err != nil {
		log.Error("error while processing request", pkglog.Err(err))
		return h.handleError(err, nil)
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.responseTimeout)
	defer cancel()

	shortened, err := h.service.ShortenURL(ctx, req.OriginalURL)
	if err != nil {
		log.Error("failed to generate url", pkglog.Err(err))
	}
	return h.handleError(err, shortened)
}

func (h *URLHandler) getOriginalURL(r *http.Request) resp.Response {
	const op = "URLHandler.getOriginalURL"
	log := h.logger.With(
		slog.String("op", op),
		slog.String("request_id", middleware.GetReqID(r.Context())),
	)

	req, err := types.CreateGetOriginalURLRequest(r)
	if err != nil {
		log.Error("error while processing request", pkglog.Err(err))
		return h.handleError(err, nil)
	}

	ctx, cancel := context.WithTimeout(r.Context(), h.responseTimeout)
	defer cancel()

	original, err := h.service.ResolveURL(ctx, req.ShortenedURL)
	if err != nil {
		log.Error("failed to get original url", pkglog.Err(err))
	}
	return h.handleError(err, original)
}

func (h *URLHandler) handleError(err error, r any) resp.Response {
	switch {
	case err == nil:
		return resp.OK(r)
	case errors.Is(err, domain.ErrInvalidShortened) || errors.Is(err, domain.ErrInvalidOriginal):
		return resp.BadRequest(err)
	case errors.Is(err, domain.ErrShortenedNotFound):
		return resp.NotFound(err)
	case errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled):
		return resp.RequestTimeout(err)
	default:
		return resp.Unknown(err)
	}
}
