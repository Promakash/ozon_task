package repository

import (
	"context"
	"ozon_task/domain"
)

// URL defines the interface for data layer of url domain
//
//go:generate go run github.com/vektra/mockery/v2@v2.52.1 --name=URL --filename=url_repository_mock.go
type URL interface {
	// CreateOrGetShortenedURL creates a new shortened URL or returns an existing one(if concurrent execution happened).
	// Takes the original URL and its shortened version.
	// Returns the shortened URL or an error.
	CreateOrGetShortenedURL(ctx context.Context, original domain.URL, shortened domain.ShortURL) (domain.ShortURL, error)

	// GetOriginalURLByShortened retrieves the original URL by its shortened version.
	// Returns `domain.ErrOriginalNotFound` if the shortened URL is not found.
	GetOriginalURLByShortened(ctx context.Context, shortened domain.ShortURL) (domain.URL, error)

	// GetShortenedURLByOriginal retrieves the shortened URL by its original version.
	// Returns `domain.ErrShortenedNotFound` if the original URL is not found.
	GetShortenedURLByOriginal(ctx context.Context, original domain.URL) (domain.ShortURL, error)
}
