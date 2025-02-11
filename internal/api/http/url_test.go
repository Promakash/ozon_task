package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/http/httptest"
	"ozon_task/pkg/http/responses"
	"ozon_task/pkg/random"
	"testing"
	"time"

	"log/slog"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"ozon_task/domain"
	"ozon_task/internal/api/http/types"
	"ozon_task/internal/usecases/mocks"
)

var dummyLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

const responseTimeout = 5 * time.Second
const getPath = "/urls/"
const getOriginalQueryParam = "shortened"

func createJSONHandlerRequest(method, path string, payload interface{}) (*http.Request, error) {
	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req := httptest.NewRequest(method, path, bytes.NewReader(bodyBytes))
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func createGetOriginalRequest(method, path string, shortURL domain.URL) *http.Request {
	req := httptest.NewRequest(method, path, nil)
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add(getOriginalQueryParam, shortURL)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	return req
}

func TestPostShortURL_Success(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "https://vk.com"
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	assert.NoError(t, err)

	mockService.
		On("ShortenURL", mock.Anything, originalURL).
		Return(shortURL, nil)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	reqPayload := types.PostShortURLRequest{
		OriginalURL: originalURL,
	}

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, reqPayload)
	assert.NoError(t, err)

	resp := handler.postShortURL(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, shortURL, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestPostShortURL_EmptyURL(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := ""

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	reqPayload := types.PostShortURLRequest{
		OriginalURL: originalURL,
	}

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, reqPayload)
	assert.NoError(t, err)

	resp := handler.postShortURL(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

	mockService.AssertExpectations(t)
}

func TestPostShortURL_NoHTTPScheme(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "vk.com"
	changedURL := "https://vk.com"

	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	assert.NoError(t, err)

	mockService.
		On("ShortenURL", mock.Anything, changedURL).
		Return(shortURL, nil)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	reqPayload := types.PostShortURLRequest{
		OriginalURL: originalURL,
	}

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, reqPayload)
	assert.NoError(t, err)

	resp := handler.postShortURL(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, shortURL, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestPostShortURL_InsecureHTTPScheme(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "http://ozon.ru"
	changedURL := "https://ozon.ru"

	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	assert.NoError(t, err)

	mockService.
		On("ShortenURL", mock.Anything, changedURL).
		Return(shortURL, nil)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	reqPayload := types.PostShortURLRequest{
		OriginalURL: originalURL,
	}

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, reqPayload)
	assert.NoError(t, err)

	resp := handler.postShortURL(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, shortURL, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestPostShortURL_BrokenJSON(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, "{invalid json")
	assert.NoError(t, err)

	resp := handler.postShortURL(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

	mockService.AssertExpectations(t)
}

func TestPostShortURL_ContextCancel(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "https://ozon.ru"
	expectedErr := context.DeadlineExceeded
	expectedReturn := *responses.RequestTimeout(expectedErr)

	mockService.
		On("ShortenURL", mock.Anything, originalURL).
		Return("", expectedErr)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	reqPayload := types.PostShortURLRequest{
		OriginalURL: originalURL,
	}

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, reqPayload)
	assert.NoError(t, err)

	resp := handler.postShortURL(req)

	assert.Equal(t, http.StatusRequestTimeout, resp.StatusCode())
	assert.Equal(t, expectedReturn, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestPostShortURL_UnExpectedDBError(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "https://ozon.ru"
	expectedErr := errors.New("no connection to the db")
	expectedReturn := *responses.Unknown(expectedErr)

	mockService.
		On("ShortenURL", mock.Anything, originalURL).
		Return("", expectedErr)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	reqPayload := types.PostShortURLRequest{
		OriginalURL: originalURL,
	}

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, reqPayload)
	assert.NoError(t, err)

	resp := handler.postShortURL(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	assert.Equal(t, expectedReturn, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_ExistedURL(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "https://ozon.ru"
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	assert.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	mockService.
		On("ResolveURL", mock.Anything, shortURL).
		Return(originalURL, nil)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)

	assert.Equal(t, http.StatusOK, resp.StatusCode())
	assert.Equal(t, originalURL, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_NotFound(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	assert.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	mockService.
		On("ResolveURL", mock.Anything, shortURL).
		Return("", domain.ErrOriginalNotFound)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)

	assert.Equal(t, http.StatusNotFound, resp.StatusCode())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_WrongShortURLLength(t *testing.T) {
	t.Parallel()
	const urlSize = 15

	mockService := new(mocks.URL)
	shortURL, err := random.NewRandomString(urlSize, domain.AllowedSymbols)
	assert.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_WrongShortURLAlphabet(t *testing.T) {
	t.Parallel()
	const alphabet = "абвгдеёжзиклмнопрст"

	mockService := new(mocks.URL)
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, alphabet)
	assert.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_EmptyQuery(t *testing.T) {
	t.Parallel()

	mockService := new(mocks.URL)
	shortURL := ""
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := httptest.NewRequest(http.MethodGet, queryPath, nil)
	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add(getOriginalQueryParam, shortURL)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	resp := handler.getOriginalURL(req)

	assert.Equal(t, http.StatusBadRequest, resp.StatusCode())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_UnExpectedDBError(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	assert.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	mockService.
		On("ResolveURL", mock.Anything, shortURL).
		Return("", errors.New("no connection to the db"))

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)

	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode())

	mockService.AssertExpectations(t)
}
