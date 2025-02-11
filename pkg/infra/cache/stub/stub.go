package stub

import (
	"context"
	"errors"
	"ozon_task/pkg/infra/cache"
	"time"
)

var NotImplemented = errors.New("not implemented")

type Stub struct{}

func NewStub() cache.Cache {
	return &Stub{}
}

func (s *Stub) Set(ctx context.Context, key string, value interface{}, TTL time.Duration) error {
	return NotImplemented
}

func (s *Stub) Get(ctx context.Context, key string, value interface{}) error {
	return NotImplemented
}

func (s *Stub) Delete(ctx context.Context, keys ...string) error {
	return NotImplemented
}
