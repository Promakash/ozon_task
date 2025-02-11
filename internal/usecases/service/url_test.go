package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/require"
	"ozon_task/domain"
	"ozon_task/internal/repository/mocks"
	"testing"
	"time"

	"github.com/stretchr/testify/mock"
)

func TestShortenURL_NewURL(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	originalURL := "https://finance.ozon.ru"

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).
		Return("", domain.ErrShortenedNotFound)
	mockRepo.On("GetOriginalURLByShortened", mock.Anything, mock.Anything).
		Return("", domain.ErrOriginalNotFound)
	mockRepo.On("CreateOrGetShortenedURL", mock.Anything, originalURL, mock.Anything).
		Return(mock.Anything, nil)

	_, err := svc.ShortenURL(ctx, originalURL)

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestShortenURL_ExistedURL(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	originalURL := "https://finance.ozon.ru"
	shortenedURL := "abc123"

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).
		Return(shortenedURL, nil)

	result, err := svc.ShortenURL(ctx, originalURL)

	require.NoError(t, err)
	require.Equal(t, shortenedURL, result)

	mockRepo.AssertExpectations(t)
}

func TestShortenURL_RetryOnCollision(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	originalURL := "https://finance.ozon.ru"

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).
		Return("", domain.ErrShortenedNotFound)
	mockRepo.On("GetOriginalURLByShortened", mock.Anything, mock.Anything).
		Return("https://ozon.ru", nil).Once()
	mockRepo.On("GetOriginalURLByShortened", mock.Anything, mock.Anything).
		Return("", domain.ErrOriginalNotFound)
	mockRepo.On("CreateOrGetShortenedURL", mock.Anything, originalURL, mock.Anything).
		Return(mock.Anything, nil)

	_, err := svc.ShortenURL(ctx, originalURL)

	require.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestShortenURL_ContextTimeout(t *testing.T) {
	t.Parallel()
	const operationTimeout = time.Second * 5

	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx, cancel := context.WithTimeout(context.Background(), operationTimeout)
	defer cancel()
	originalURL := "https://finance.ozon.ru"

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).
		Return("", domain.ErrOriginalNotFound)
	mockRepo.On("GetOriginalURLByShortened", mock.Anything, mock.Anything).
		Return("", context.DeadlineExceeded)

	result, err := svc.ShortenURL(ctx, originalURL)

	require.Error(t, err)
	require.Empty(t, result)
}

func TestShortenURL_UnexpectedDBError(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	originalURL := "https://finance.ozon.ru"

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).
		Return("", errors.New("no connection to the db"))

	result, err := svc.ShortenURL(ctx, originalURL)

	require.Error(t, err)
	require.Empty(t, result)

	mockRepo.AssertExpectations(t)
}

func TestResolveURL_ExistedURL(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	shortenedURL := "abc123"
	originalURL := "https://finance.ozon.ru"

	mockRepo.On("GetOriginalURLByShortened", mock.Anything, shortenedURL).
		Return(originalURL, nil)

	result, err := svc.ResolveURL(ctx, shortenedURL)

	require.NoError(t, err)
	require.Equal(t, originalURL, result)

	mockRepo.AssertExpectations(t)
}

func TestResolveURL_NotFound(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	shortenedURL := "abc123"

	mockRepo.On("GetOriginalURLByShortened", mock.Anything, shortenedURL).
		Return("", domain.ErrOriginalNotFound)

	result, err := svc.ResolveURL(ctx, shortenedURL)

	require.Error(t, err)
	require.Empty(t, result)
	require.ErrorIs(t, err, domain.ErrOriginalNotFound)

	mockRepo.AssertExpectations(t)
}
