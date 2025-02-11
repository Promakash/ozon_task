package stub

import (
	"context"
	"errors"
	"ozon_task/pkg/infra/cache"
	"time"
)

var ErrNotImplemented = errors.New("not implemented")

type Stub struct{}

func NewStub() cache.Cache {
	return &Stub{}
}

func (s *Stub) Set(_ context.Context, key string, value interface{}, TTL time.Duration) error {
	return ErrNotImplemented
}

func (s *Stub) Get(_ context.Context, key string, value interface{}) error {
	return ErrNotImplemented
}

func (s *Stub) Delete(_ context.Context, keys ...string) error {
	return ErrNotImplemented
}
