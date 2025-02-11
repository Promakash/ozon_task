package usecases

import (
	"context"
	"ozon_task/domain"
)

// URL defines the interface for the service layer of the URL domain.
//
//go:generate go run github.com/vektra/mockery/v2@v2.50 --name=URL --filename=url_service_mock.go
type URL interface {
	// ShortenURL generates a shortened version of the given original URL.
	// If the URL has already been shortened, it returns the existing shortened URL.
	// Returns a shortened URL or an error.
	ShortenURL(ctx context.Context, original domain.URL) (domain.ShortURL, error)

	// ResolveURL retrieves the original URL from its shortened version.
	// Returns `domain.ErrOriginalNotFound` if the shortened URL does not exist.
	ResolveURL(ctx context.Context, shortened domain.ShortURL) (domain.URL, error)
}
