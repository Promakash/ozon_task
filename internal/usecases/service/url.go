package service

import (
	"context"
	"errors"
	"fmt"
	"ozon_task/domain"
	"ozon_task/internal/repository"
	pkgrandom "ozon_task/pkg/random"
)

type URLService struct {
	repo repository.URL
}

func NewURLService(repo repository.URL) *URLService {
	return &URLService{
		repo: repo,
	}
}

func (s *URLService) generateShortURL(ctx context.Context) (domain.ShortURL, error) {
	for {
		select {
		case <-ctx.Done():
			return "", fmt.Errorf("generateShortURL: context cancelled: %w", ctx.Err())

		default:
			newURL, err := pkgrandom.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
			if err != nil {
				return "", fmt.Errorf("generateShortURL: failed to generate random string: %w", err)
			}

			_, err = s.repo.GetOriginalURLByShortened(ctx, newURL)
			if !errors.Is(err, domain.ErrOriginalNotFound) {
				continue
			}

			return newURL, nil
		}
	}
}

func (s *URLService) ShortenURL(ctx context.Context, original domain.URL) (domain.ShortURL, error) {

	shortened, err := s.repo.GetShortenedURLByOriginal(ctx, original)
	if err == nil {
		return shortened, nil
	} else if ok := errors.Is(err, domain.ErrShortenedNotFound); !ok {
		return "", fmt.Errorf("ShortenURL: failed to check for existing shortened URL for %q: %w", original, err)
	}

	newURL, err := s.generateShortURL(ctx)
	if err != nil {
		return "", fmt.Errorf("ShortenURL: %w", err)
	}

	err = s.repo.PutShortenedURL(ctx, original, &newURL)
	if err != nil {
		return "", fmt.Errorf("ShortenURL: failed to put new shortened URL %q for original %q: %w", newURL, original, err)
	}

	return newURL, nil
}

func (s *URLService) ResolveURL(ctx context.Context, shortened domain.ShortURL) (domain.URL, error) {
	original, err := s.repo.GetOriginalURLByShortened(ctx, shortened)
	if err != nil {
		return "", fmt.Errorf("ResolveURL: failed to resolve original URL for shortened %q: %w", shortened, err)
	}

	return original, nil
}
