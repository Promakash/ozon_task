package service

import (
	"context"
	"errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"ozon_task/domain"
	"ozon_task/internal/repository/mocks"
	"testing"
	"time"
)

func TestShortenURL_NewURL(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	originalURL := "https://finance.ozon.ru"

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).Return("", domain.ErrShortenedNotFound)
	mockRepo.On("GetOriginalURLByShortened", mock.Anything, mock.Anything).Return("", domain.ErrOriginalNotFound)
	mockRepo.On("PutShortenedURL", mock.Anything, originalURL, mock.Anything).Return(nil)

	_, err := svc.ShortenURL(ctx, originalURL)

	assert.NoError(t, err)

	mockRepo.AssertExpectations(t)
}

func TestShortenURL_ExistedURL(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	originalURL := "https://finance.ozon.ru"
	shortenedURL := "abc123"

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).Return(shortenedURL, nil)

	result, err := svc.ShortenURL(ctx, originalURL)

	assert.NoError(t, err)
	assert.Equal(t, shortenedURL, result)

	mockRepo.AssertExpectations(t)
}

func TestShortenURL_RetryOnCollision(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	originalURL := "https://finance.ozon.ru"

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).Return("", domain.ErrShortenedNotFound)
	mockRepo.On("GetOriginalURLByShortened", mock.Anything, mock.Anything).Return("https://ozon.ru", nil).Once()
	mockRepo.On("GetOriginalURLByShortened", mock.Anything, mock.Anything).Return("", domain.ErrOriginalNotFound)
	mockRepo.On("PutShortenedURL", mock.Anything, originalURL, mock.Anything).Return(nil)

	result, err := svc.ShortenURL(ctx, originalURL)

	assert.NoError(t, err)
	assert.NotEmpty(t, result)

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

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).Return("", domain.ErrOriginalNotFound)
	mockRepo.On("GetOriginalURLByShortened", mock.Anything, mock.Anything).Return("https://ozon.ru", nil)

	result, err := svc.ShortenURL(ctx, originalURL)

	assert.Error(t, err)
	assert.Empty(t, result)
}

func TestShortenURL_UnexpectedDBError(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	originalURL := "https://finance.ozon.ru"

	mockRepo.On("GetShortenedURLByOriginal", mock.Anything, originalURL).Return("", errors.New("no connection to the db"))

	result, err := svc.ShortenURL(ctx, originalURL)

	assert.Error(t, err)
	assert.Empty(t, result)

	mockRepo.AssertExpectations(t)
}

func TestResolveURL_ExistedURL(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	shortenedURL := "abc123"
	originalURL := "https://finance.ozon.ru"

	mockRepo.On("GetOriginalURLByShortened", mock.Anything, shortenedURL).Return(originalURL, nil)

	result, err := svc.ResolveURL(ctx, shortenedURL)

	assert.NoError(t, err)
	assert.Equal(t, originalURL, result)

	mockRepo.AssertExpectations(t)
}

func TestResolveURL_NotFound(t *testing.T) {
	t.Parallel()
	mockRepo := new(mocks.URL)
	svc := NewURLService(mockRepo)

	ctx := context.Background()
	shortenedURL := "abc123"

	mockRepo.On("GetOriginalURLByShortened", mock.Anything, shortenedURL).Return("", domain.ErrOriginalNotFound)

	result, err := svc.ResolveURL(ctx, shortenedURL)

	assert.Error(t, err)
	assert.Empty(t, result)
	assert.Equal(t, true, errors.Is(err, domain.ErrOriginalNotFound))

	mockRepo.AssertExpectations(t)
}
