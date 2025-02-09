package main

import (
	"context"
	"errors"
	"golang.org/x/sync/errgroup"
	"log/slog"
	_ "ozon_task/docs"
	grpcapp "ozon_task/internal/app/grpc"
	httpapp "ozon_task/internal/app/http"
	"ozon_task/internal/config"
	"ozon_task/internal/repository/postgres"
	"ozon_task/internal/usecases/service"
	pkgconfig "ozon_task/pkg/config"
	"ozon_task/pkg/infra"
	pkglog "ozon_task/pkg/log"
	"ozon_task/pkg/shutdown"
	"time"
)

//	@title			URL Shortener API
//	@version		1.0
//	@description	API for URL Shortener service
//	@termsOfService	http://swagger.io/terms/

//	@host		localhost:8080
//	@BasePath	/api/v1/

const (
	ConfigEnvVar = "SHORTENER_CONFIG"
	APIPath      = "/api/v1"
)

func main() {
	cfg := config.Config{}
	pkgconfig.MustLoad(ConfigEnvVar, &cfg)

	log, file := pkglog.NewLogger(cfg.Logger)
	defer func() { _ = file.Close() }()
	slog.SetDefault(log)
	log.Info("Starting URL Shortener", slog.Any("config", cfg))

	dbPool, err := infra.NewPostgresPool(cfg.PG)
	if err != nil {
		pkglog.Fatal(log, "error while setting new postgres connection: ", err)
	}
	defer dbPool.Close()
	urlRepo := postgres.NewURLRepository(dbPool)

	urlService := service.NewURLService(urlRepo)

	grpcApp := grpcapp.New(log, urlService, cfg.GRPC)
	httpApp := httpapp.New(log, APIPath, urlService, cfg.HTTPServer)

	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		return shutdown.ListenSignal(ctx, log)
	})

	g.Go(func() error {
		return httpApp.Run()
	})

	g.Go(func() error {
		return grpcApp.Run()
	})

	g.Go(func() error {
		<-ctx.Done()
		log.Info("Shutdown signal received, stopping servers")
		grpcApp.Stop()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return httpApp.Stop(shutdownCtx)
	})

	err = g.Wait()
	if err != nil && !errors.Is(err, shutdown.ErrOSSignal) {
		log.Error("Exit reason", slog.String("error", err.Error()))
	}
}
