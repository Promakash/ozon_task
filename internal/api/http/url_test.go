package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/stretchr/testify/require"
	"io"
	"net/http"
	"net/http/httptest"
	"ozon_task/pkg/http/responses"
	"ozon_task/pkg/random"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"

	"log/slog"

	"github.com/stretchr/testify/mock"

	"ozon_task/domain"
	"ozon_task/internal/api/http/types"
	"ozon_task/internal/usecases/mocks"
)

var dummyLogger = slog.New(slog.NewTextHandler(io.Discard, nil))

const httpPath = "http://localhost:8080/"
const responseTimeout = 5 * time.Second
const getPath = "api/v1/resolve/"
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
	fullURL := fmt.Sprintf("%s%s%s", httpPath, path, shortURL)

	req := httptest.NewRequest(method, fullURL, nil)
	chiCtx := chi.NewRouteContext()

	chiCtx.URLParams.Add(getOriginalQueryParam, shortURL)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))
	return req
}

func TestPostShortURL_Success(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "https://ozon.ru"
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	require.NoError(t, err)

	mockService.
		On("ShortenURL", mock.Anything, originalURL).
		Return(shortURL, nil)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	reqPayload := types.PostShortURLRequest{
		OriginalURL: originalURL,
	}

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, reqPayload)
	require.NoError(t, err)

	resp := handler.postShortURL(req)
	expectedResp := &types.PostShortURLResponse{ShortenedURL: shortURL}

	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Equal(t, expectedResp, resp.GetPayload())

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
	require.NoError(t, err)

	resp := handler.postShortURL(req)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode())

	mockService.AssertExpectations(t)
}

func TestPostShortURL_NoHTTPScheme(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "ozon.ru"
	changedURL := "https://ozon.ru"

	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	require.NoError(t, err)

	mockService.
		On("ShortenURL", mock.Anything, changedURL).
		Return(shortURL, nil)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	reqPayload := types.PostShortURLRequest{
		OriginalURL: originalURL,
	}

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, reqPayload)
	require.NoError(t, err)

	resp := handler.postShortURL(req)
	expectedResp := &types.PostShortURLResponse{ShortenedURL: shortURL}

	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Equal(t, expectedResp, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestPostShortURL_InsecureHTTPScheme(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "http://ozon.ru"

	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	require.NoError(t, err)

	mockService.
		On("ShortenURL", mock.Anything, originalURL).
		Return(shortURL, nil)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	reqPayload := types.PostShortURLRequest{
		OriginalURL: originalURL,
	}

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, reqPayload)
	require.NoError(t, err)

	resp := handler.postShortURL(req)
	expectedResp := &types.PostShortURLResponse{ShortenedURL: shortURL}

	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Equal(t, expectedResp, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestPostShortURL_BrokenJSON(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req, err := createJSONHandlerRequest(http.MethodPost, postShortPath, "{invalid json")
	require.NoError(t, err)

	resp := handler.postShortURL(req)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode())

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
	require.NoError(t, err)

	resp := handler.postShortURL(req)

	require.Equal(t, http.StatusRequestTimeout, resp.StatusCode())
	require.Equal(t, expectedReturn, resp.GetPayload())

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
	require.NoError(t, err)

	resp := handler.postShortURL(req)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode())
	require.Equal(t, expectedReturn, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_ExistedURL(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	originalURL := "https://ozon.ru"
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	require.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	mockService.
		On("ResolveURL", mock.Anything, shortURL).
		Return(originalURL, nil)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)
	expectedResp := &types.GetOriginalURLResponse{OriginalURL: originalURL}

	require.Equal(t, http.StatusOK, resp.StatusCode())
	require.Equal(t, expectedResp, resp.GetPayload())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_NotFound(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	require.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	mockService.
		On("ResolveURL", mock.Anything, shortURL).
		Return("", domain.ErrOriginalNotFound)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)

	require.Equal(t, http.StatusNotFound, resp.StatusCode())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_WrongShortURLLength(t *testing.T) {
	t.Parallel()
	const urlSize = 15

	mockService := new(mocks.URL)
	shortURL, err := random.NewRandomString(urlSize, domain.AllowedSymbols)
	require.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode())

	mockService.AssertExpectations(t)
}

func TestGetOriginalURL_WrongShortURLAlphabet(t *testing.T) {
	t.Parallel()
	const alphabet = "абвгдеёжзиклмнопрст"

	mockService := new(mocks.URL)
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, alphabet)
	require.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode())

	mockService.AssertExpectations(t)
}

/*func TestGetOriginalURL_EmptyQuery(t *testing.T) {
	t.Parallel()

	mockService := new(mocks.URL)
	shortURL := "" // Пустое значение

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := httptest.NewRequest(http.MethodGet, getPath, nil)

	chiCtx := chi.NewRouteContext()
	chiCtx.URLParams.Add(getOriginalQueryParam, shortURL)

	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, chiCtx))

	resp := handler.getOriginalURL(req)

	require.Equal(t, http.StatusBadRequest, resp.StatusCode())

	mockService.AssertExpectations(t)
}*/

func TestGetOriginalURL_UnExpectedDBError(t *testing.T) {
	t.Parallel()
	mockService := new(mocks.URL)
	shortURL, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	require.NoError(t, err)
	queryPath := fmt.Sprintf("%s%s", getPath, shortURL)

	mockService.
		On("ResolveURL", mock.Anything, shortURL).
		Return("", errors.New("no connection to the db"))

	handler := NewURLHandler(dummyLogger, mockService, responseTimeout)

	req := createGetOriginalRequest(http.MethodGet, queryPath, shortURL)

	resp := handler.getOriginalURL(req)

	require.Equal(t, http.StatusInternalServerError, resp.StatusCode())

	mockService.AssertExpectations(t)
}
