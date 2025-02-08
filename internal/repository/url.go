package repository

import (
	"context"
	"ozon_task/domain"
)

//go:generate go run github.com/vektra/mockery/v2@v2.52.1 --name=URL --filename=url_repository_mock.go
type URL interface {
	PutShortenedURL(ctx context.Context, original domain.URL, shortened *domain.URL) error
	GetOriginalURLByShortened(ctx context.Context, shortened domain.URL) (domain.URL, error)
	GetShortenedURLByOriginal(ctx context.Context, original domain.URL) (domain.URL, error)
}
