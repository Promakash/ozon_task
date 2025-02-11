package inmem

import (
	"context"
	"ozon_task/domain"
	"ozon_task/internal/repository"
	"ozon_task/pkg/infra/kv"
)

type URLRepository struct {
	storage kv.Storage
}

func NewURLRepository(storage kv.Storage) repository.URL {
	return &URLRepository{
		storage: storage,
	}
}

func (r *URLRepository) CreateOrGetShortenedURL(_ context.Context, original domain.URL, shortened domain.ShortURL) (domain.ShortURL, error) {
	if existingShort, ok := r.storage.Get(original); ok {
		return existingShort, nil
	}

	r.storage.Set(original, shortened)
	r.storage.Set(shortened, original)

	return shortened, nil
}

func (r *URLRepository) GetOriginalURLByShortened(_ context.Context, shortened domain.ShortURL) (domain.URL, error) {
	if original, ok := r.storage.Get(shortened); ok {
		return original, nil
	}
	return "", domain.ErrOriginalNotFound
}

func (r *URLRepository) GetShortenedURLByOriginal(_ context.Context, original domain.URL) (domain.ShortURL, error) {
	if shortened, ok := r.storage.Get(original); ok {
		return shortened, nil
	}
	return "", domain.ErrShortenedNotFound
}
