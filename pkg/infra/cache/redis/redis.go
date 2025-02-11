package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log/slog"
	"ozon_task/pkg/infra/cache"
	pkglog "ozon_task/pkg/log"
	"time"
)

const CacheAlwaysAlive = redis.KeepTTL

type Redis struct {
	client *redis.Client
	logger *slog.Logger
}

func NewRedisClient(cfg Config) (*redis.Client, error) {
	address := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	rdb := redis.NewClient(&redis.Options{
		Addr:         address,
		Password:     cfg.Password,
		WriteTimeout: cfg.WriteTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		DB:           0,
		Protocol:     3,
	})

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	_, err := rdb.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	return rdb, nil
}

func NewRedisService(client *redis.Client, logger *slog.Logger) cache.Cache {
	return &Redis{
		client: client,
		logger: logger,
	}
}

func (r *Redis) Set(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
	const op = "Redis.Set"
	log := r.logger.With(
		slog.String("op", op),
	)

	bytes, err := json.Marshal(value)
	if err != nil {
		log.Error("error while marshalling json", pkglog.Err(err))
		return err
	}

	err = r.client.Set(ctx, key, bytes, TTL).Err()
	if err != nil {
		log.Error("error while setting new data", pkglog.Err(err))
		return err
	}

	return nil
}

func (r *Redis) Get(ctx context.Context, key string, value interface{}) error {
	const op = "Redis.Get"
	log := r.logger.With(
		slog.String("op", op),
	)

	val, err := r.client.Get(ctx, key).Result()
	if err != nil {
		log.Error("error while getting data", pkglog.Err(err))
		return err
	}

	err = json.Unmarshal([]byte(val), value)
	if err != nil {
		log.Error("error while unmarshalling data", pkglog.Err(err))
		return err
	}

	return nil
}

func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	const op = "Redis.Delete"
	log := r.logger.With(
		slog.String("op", op),
	)

	err := r.client.Del(ctx, keys...).Err()
	if err != nil {
		log.Error("error while deleting data", pkglog.Err(err))
		return err
	}

	return nil
}

func ShutdownClient(client *redis.Client) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = client.Shutdown(ctx)
}
