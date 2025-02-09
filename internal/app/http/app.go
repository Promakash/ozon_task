package http

import (
	"context"
	"log/slog"
	"net/http"
	apihttp "ozon_task/internal/api/http"
	"ozon_task/internal/config"
	"ozon_task/internal/usecases"
	"ozon_task/pkg/http/handlers"
)

type App struct {
	log    *slog.Logger
	server *http.Server
}

func New(
	log *slog.Logger,
	APIPath string,
	service usecases.URL,
	cfg config.HTTPConfig,
) *App {
	urlHandler := apihttp.NewURLHandler(
		log,
		service,
		cfg.OperationsTimeout,
	)

	publicHandler := handlers.NewHandler(
		APIPath,
		handlers.WithLogging(log),
		handlers.WithProfilerHandlers(),
		handlers.WithRequestID(),
		handlers.WithRecover(),
		handlers.WithSwagger(),
		handlers.WithHealthHandler(),
		handlers.WithErrHandlers(),
		urlHandler.WithURLHandlers(),
	)

	srv := &http.Server{
		Addr:         cfg.Address,
		Handler:      publicHandler,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
		IdleTimeout:  cfg.IdleTimeout,
	}

	return &App{
		log:    log,
		server: srv,
	}
}

func (a *App) Run() error {
	const op = "http.App"

	log := a.log.With(
		slog.String("op", op),
		slog.String("address", a.server.Addr),
	)

	log.Info("HTTP server starting")
	return a.server.ListenAndServe()
}

func (a *App) Stop(ctx context.Context) error {
	const op = "http.Stop"
	log := a.log.With(slog.String("op", op))

	log.Info("HTTP server shutting down", slog.String("addr", a.server.Addr))
	return a.server.Shutdown(ctx)
}
