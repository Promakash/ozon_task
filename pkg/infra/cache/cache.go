package cache

import (
	"context"
	"time"
)

type Cache interface {
	Set(ctx context.Context, key string, value interface{}, TTL time.Duration) error
	Get(ctx context.Context, key string, value interface{}) error
	Delete(ctx context.Context, keys ...string) error
}
