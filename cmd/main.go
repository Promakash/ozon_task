package main

import (
	"context"
	"errors"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"golang.org/x/sync/errgroup"
	"log/slog"
	_ "ozon_task/docs"
	grpcapp "ozon_task/internal/app/grpc"
	httpapp "ozon_task/internal/app/http"
	"ozon_task/internal/config"
	"ozon_task/internal/repository"
	"ozon_task/internal/repository/inmem"
	"ozon_task/internal/repository/postgres"
	"ozon_task/internal/usecases/service"
	pkgconfig "ozon_task/pkg/config"
	"ozon_task/pkg/infra"
	pkgredis "ozon_task/pkg/infra/cache/redis"
	"ozon_task/pkg/infra/cache/stub"
	pkginmem "ozon_task/pkg/infra/kv/inmem"
	pkglog "ozon_task/pkg/log"
	"ozon_task/pkg/shutdown"
	"runtime"
	"time"
)

//	@title			URL Shortener API
//	@version		1.0
//	@description	API for URL Shortener service
//	@termsOfService	http://swagger.io/terms/

//	@host		localhost:8080
//	@BasePath	/api/v1/

const (
	configEnvVar = "SHORTENER_CONFIG"
	APIPath      = "/api/v1"
)

// flags
// -inmem - use inmemory storage instead of postgresql
// -redis - use redis as cache (works only if inmem disabled and redis is live)
func main() {
	flags := config.ParseFlags()
	cfg := config.Config{}
	pkgconfig.MustLoad(configEnvVar, &cfg)

	log, file := pkglog.NewLogger(cfg.Logger)
	defer func() { _ = file.Close() }()
	slog.SetDefault(log)
	log.Info("Starting URL Shortener", slog.Any("config", cfg))

	urlRepo, dbPool, redisClient := initStorage(flags, cfg, log)

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
		return shutdownServices(grpcApp, httpApp)
	})

	err := g.Wait()

	if dbPool != nil {
		dbPool.Close()
	}

	if redisClient != nil {
		_ = redisClient.Close()
	}

	if err != nil && !errors.Is(err, shutdown.ErrOSSignal) {
		log.Error("Exit reason", slog.String("error", err.Error()))
	}
}

// initStorage inits repository depend on flags
func initStorage(flags config.AppFlags, cfg config.Config, log *slog.Logger) (repository.URL, *pgxpool.Pool, *redis.Client) {
	var (
		urlRepo     repository.URL
		dbPool      *pgxpool.Pool
		redisClient *redis.Client
	)

	if flags.UseInMemStorage {
		partitionsNumber := runtime.GOMAXPROCS(0) * 2
		kv := pkginmem.NewPartitionedKVStorage(partitionsNumber)
		urlRepo = inmem.NewURLRepository(kv)
		log.Info("Using in-memory storage")
		return urlRepo, nil, nil
	}

	var err error
	dbPool, err = infra.NewPostgresPool(cfg.PG)
	if err != nil {
		pkglog.Fatal(log, "error while setting new postgres connection: ", err)
	}

	if flags.UseRedis {
		redisClient, err = pkgredis.NewRedisClient(cfg.Redis)
		if err != nil {
			pkglog.Fatal(log, "error while setting new redis connection: ", err)
		}
		cacheService := pkgredis.NewRedisService(redisClient, log)
		urlRepo = postgres.NewURLRepository(dbPool, cacheService, cfg.Redis.TTL, cfg.Redis.WriteTimeout)
		log.Info("Using Postgres with redis cache")
	} else {
		urlRepo = postgres.NewURLRepository(dbPool, stub.NewStub(), 0, 0)
		log.Info("Using Postgres without redis")
	}

	return urlRepo, dbPool, redisClient
}

// shutdownServices gracefully shutdown apps
func shutdownServices(grpcApp *grpcapp.App, httpApp *httpapp.App) error {
	grpcApp.Stop()
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return httpApp.Stop(shutdownCtx)
}
