package service

import (
	"context"
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

func (s *URLService) ShortenURL(original domain.URL) (domain.URL, error) {
	const ShortenedURLSize = 10

	shorted, err := s.repo.GetShortenedURLByOriginal(context.TODO(), original)
	if err == nil {
		return shorted, nil
	}

	newURL, err := pkgrandom.NewRandomString(ShortenedURLSize)
	if err != nil {
		return "", err
	}

	err = s.repo.PutShortenedURL(context.TODO(), original, &newURL)
	if err != nil {
		return "", err
	}

	return newURL, nil
}

func (s *URLService) ResolveURL(shorted domain.URL) (domain.URL, error) {
	original, err := s.repo.GetOriginalURLByShortened(context.TODO(), shorted)
	if err != nil {
		return "", err
	}

	return original, nil
}
