package usecases

import (
	"context"
	"ozon_task/domain"
)

//go:generate go run github.com/vektra/mockery/v2@v2.50 --name=URL --filename=url_service_mock.go
type URL interface {
	ShortenURL(ctx context.Context, original domain.URL) (domain.ShortURL, error)
	ResolveURL(ctx context.Context, shortened domain.ShortURL) (domain.URL, error)
}
