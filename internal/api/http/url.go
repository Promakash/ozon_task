package http

import (
	"log/slog"
	"ozon_task/internal/usecases"
)

type URLHandler struct {
	logger  *slog.Logger
	service usecases.URL
}

func NewURLHandler(logger *slog.Logger, service usecases.URL) *URLHandler {
	return &URLHandler{
		logger:  logger,
		service: service,
	}
}
