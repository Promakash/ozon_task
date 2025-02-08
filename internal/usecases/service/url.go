package service

import (
	"context"
	"errors"
	"log/slog"
	"ozon_task/domain"
	"ozon_task/internal/repository"
	pkgrandom "ozon_task/pkg/random"
)

type URLService struct {
	log  *slog.Logger
	repo repository.URL
}

func NewURLService(log *slog.Logger, repo repository.URL) *URLService {
	return &URLService{
		log:  log,
		repo: repo,
	}
}

func (s *URLService) generateURL(ctx context.Context) (domain.URL, error) {
	for {
		select {
		case <-ctx.Done():
			return "", ctx.Err()

		default:
			newURL, err := pkgrandom.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
			if err != nil {
				return "", err
			}

			_, err = s.repo.GetOriginalURLByShortened(ctx, newURL)
			if !errors.Is(err, domain.ErrOriginalNotFound) {
				continue
			}

			return newURL, nil
		}
	}
}

func (s *URLService) ShortenURL(ctx context.Context, original domain.URL) (domain.URL, error) {

	shortened, err := s.repo.GetShortenedURLByOriginal(ctx, original)
	if err == nil {
		return shortened, nil
	} else if ok := errors.Is(err, domain.ErrShortenedNotFound); !ok {
		return "", err
	}

	newURL, err := s.generateURL(ctx)
	if err != nil {
		return "", err
	}

	err = s.repo.PutShortenedURL(ctx, original, &newURL)
	if err != nil {
		return "", err
	}

	return newURL, nil
}

func (s *URLService) ResolveURL(ctx context.Context, shorted domain.URL) (domain.URL, error) {
	original, err := s.repo.GetOriginalURLByShortened(ctx, shorted)
	if err != nil {
		return "", err
	}

	return original, nil
}
