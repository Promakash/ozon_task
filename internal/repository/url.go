package repository

import (
	"context"
	"ozon_task/domain"
)

type URL interface {
	PutShortenedURL(ctx context.Context, original domain.URL, shortened *domain.URL) error
	GetOriginalURLByShortened(ctx context.Context, shortened domain.URL) (domain.URL, error)
	GetShortenedURLByOriginal(ctx context.Context, original domain.URL) (domain.URL, error)
}
