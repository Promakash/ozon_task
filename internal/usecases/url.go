package usecases

import (
	"context"
	"ozon_task/domain"
)

type URL interface {
	ShortenURL(ctx context.Context, original domain.URL) (domain.URL, error)
	ResolveURL(ctx context.Context, shorted domain.URL) (domain.URL, error)
}
