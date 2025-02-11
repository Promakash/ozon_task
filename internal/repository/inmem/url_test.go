package inmem_test

import (
	"context"
	"github.com/stretchr/testify/require"
	"testing"

	"ozon_task/domain"
	"ozon_task/internal/repository/inmem"
	pkginmem "ozon_task/pkg/infra/kv/inmem"
)

const partitionsCount = 6

func TestURLRepository_CreateOrGetShortenedURL(t *testing.T) {
	ctx := context.Background()
	storage := pkginmem.NewPartitionedKVStorage(partitionsCount)
	repo := inmem.NewURLRepository(storage)

	originalURL := "https://ozon.ru"
	shortenedURL := "abc123XYZ"

	result, err := repo.CreateOrGetShortenedURL(ctx, originalURL, shortenedURL)
	require.NoError(t, err)
	require.Equal(t, shortenedURL, result)

	result2, err := repo.CreateOrGetShortenedURL(ctx, originalURL, "differentShort")
	require.NoError(t, err)
	require.Equal(t, shortenedURL, result2)
}

func TestURLRepository_GetOriginalURLByShortened(t *testing.T) {
	ctx := context.Background()
	storage := pkginmem.NewPartitionedKVStorage(partitionsCount)
	repo := inmem.NewURLRepository(storage)

	originalURL := "https://ozon.ru"
	shortenedURL := "abc123XYZ"

	storage.Set(shortenedURL, originalURL)
	result, err := repo.GetOriginalURLByShortened(ctx, shortenedURL)
	require.NoError(t, err)
	require.Equal(t, originalURL, result)

	_, err = repo.GetOriginalURLByShortened(ctx, "nonexistent")
	require.ErrorIs(t, err, domain.ErrOriginalNotFound)
}

func TestURLRepository_GetShortenedURLByOriginal(t *testing.T) {
	ctx := context.Background()
	storage := pkginmem.NewPartitionedKVStorage(partitionsCount)
	repo := inmem.NewURLRepository(storage)

	originalURL := "https://ozon.ru"
	shortenedURL := "abc123XYZ"
	nonExistURL := "https://fintech.ozon.ru"

	storage.Set(originalURL, shortenedURL)
	result, err := repo.GetShortenedURLByOriginal(ctx, originalURL)
	require.NoError(t, err)
	require.Equal(t, shortenedURL, result)

	_, err = repo.GetShortenedURLByOriginal(ctx, nonExistURL)
	require.ErrorIs(t, err, domain.ErrShortenedNotFound)
}
