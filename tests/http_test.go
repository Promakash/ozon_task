package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/stretchr/testify/require"
	"net/http"
	"ozon_task/domain"
	"ozon_task/internal/api/http/types"
	"ozon_task/pkg/random"
	"ozon_task/tests/suite"
	"testing"

	"github.com/stretchr/testify/assert"
)

const (
	PostShortPath  = "api/v1/shorten"
	GetOriginalURL = "api/v1/resolve"
)

func SendPostRequest(
	ctx context.Context,
	client *http.Client,
	baseURL, path string,
	payload types.PostShortURLRequest,
) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s/%s", baseURL, path)

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(http.MethodPost, fullURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req = req.WithContext(ctx)

	return client.Do(req)
}

func SendGetRequest(
	ctx context.Context,
	client *http.Client,
	baseURL, path string,
	queryParam domain.ShortURL,
) (*http.Response, error) {
	fullURL := fmt.Sprintf("%s/%s/%s", baseURL, path, queryParam)

	req, err := http.NewRequest(http.MethodGet, fullURL, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	return client.Do(req)
}

func TestPostShortURL_SuccessURLs(t *testing.T) {
	t.Parallel()
	ctx, st := suite.NewHTTPSuite(t)

	tests := []struct {
		name           string
		url            domain.URL
		expectedStatus int
	}{
		{"Fully valid url", "https://fintech.ozon.ru/", http.StatusOK},
		{"Insecure HTTP scheme", "http://fintech.ozon.ru/", http.StatusOK},
		{"No HTTP Scheme", "fintech.ozon.ru/", http.StatusOK},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := SendPostRequest(
				ctx,
				st.Client,
				st.BaseURL,
				PostShortPath,
				types.PostShortURLRequest{OriginalURL: test.url},
			)
			require.NoError(t, err)
			assert.Equal(t, test.expectedStatus, res.StatusCode)

			var response types.PostShortURLResponse
			err = json.NewDecoder(res.Body).Decode(&response)
			require.NoError(t, err)

			assert.NotEmpty(t, response.ShortenedURL)
		})
	}
}

func TestPostShortURL_SameURL(t *testing.T) {
	t.Parallel()
	ctx, st := suite.NewHTTPSuite(t)

	const validURL = "https://finance.ozon.ru/docs/legal#legal"

	res, err := SendPostRequest(
		ctx,
		st.Client,
		st.BaseURL,
		PostShortPath,
		types.PostShortURLRequest{OriginalURL: validURL},
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var firstResponse types.PostShortURLResponse
	err = json.NewDecoder(res.Body).Decode(&firstResponse)
	require.NoError(t, err)

	resSecond, err := SendPostRequest(
		ctx,
		st.Client,
		st.BaseURL,
		PostShortPath,
		types.PostShortURLRequest{OriginalURL: validURL},
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resSecond.StatusCode)

	var secondResponse types.PostShortURLResponse
	err = json.NewDecoder(resSecond.Body).Decode(&secondResponse)
	require.NoError(t, err)

	assert.Equal(t, firstResponse.ShortenedURL, secondResponse.ShortenedURL)
}

func TestPostShortURL_InvalidURLs(t *testing.T) {
	t.Parallel()
	ctx, st := suite.NewHTTPSuite(t)

	tests := []struct {
		name           string
		url            domain.URL
		expectedStatus int
	}{
		{"No top level domain", "https://ozon", http.StatusBadRequest},
		{"Empty URL", "", http.StatusBadRequest},
		{"No host no top level domain", "https://", http.StatusBadRequest},
		{"No host", "https://.ru", http.StatusBadRequest},
		{"Whitespace in host", "https://oz on.ru", http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := SendPostRequest(
				ctx,
				st.Client,
				st.BaseURL,
				PostShortPath,
				types.PostShortURLRequest{OriginalURL: test.url},
			)
			require.NoError(t, err)
			assert.Equal(t, test.expectedStatus, res.StatusCode)
		})
	}
}

func TestGetOriginalURL_Success(t *testing.T) {
	t.Parallel()
	ctx, st := suite.NewHTTPSuite(t)

	const validURL = "https://finance.ozon.ru/docs"

	res, err := SendPostRequest(
		ctx,
		st.Client,
		st.BaseURL,
		PostShortPath,
		types.PostShortURLRequest{OriginalURL: validURL},
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, res.StatusCode)

	var response types.PostShortURLResponse
	err = json.NewDecoder(res.Body).Decode(&response)
	require.NoError(t, err)

	resolveResp, err := SendGetRequest(
		ctx,
		st.Client,
		st.BaseURL,
		GetOriginalURL,
		response.ShortenedURL,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusOK, resolveResp.StatusCode)

	var resolveResponse types.GetOriginalURLResponse
	err = json.NewDecoder(resolveResp.Body).Decode(&resolveResponse)
	require.NoError(t, err)

	assert.Equal(t, validURL, resolveResponse.OriginalURL)
}

func TestGetOriginalURL_InvalidShorts(t *testing.T) {
	t.Parallel()
	ctx, st := suite.NewHTTPSuite(t)

	tests := []struct {
		name           string
		url            domain.ShortURL
		expectedStatus int
	}{
		{"Invalid length", "xZsyc_xvo11", http.StatusBadRequest},
		{"Empty URL", "", http.StatusNotFound},
		{"Invalid symbol", "@fcizawmtN", http.StatusBadRequest},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := SendGetRequest(
				ctx,
				st.Client,
				st.BaseURL,
				GetOriginalURL,
				test.url,
			)
			require.NoError(t, err)
			assert.Equal(t, test.expectedStatus, res.StatusCode)
		})
	}
}

func TestGetOriginalURL_NotFound(t *testing.T) {
	t.Parallel()
	ctx, st := suite.NewHTTPSuite(t)

	validShort, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	require.NoError(t, err)

	res, err := SendGetRequest(
		ctx,
		st.Client,
		st.BaseURL,
		GetOriginalURL,
		validShort,
	)
	require.NoError(t, err)
	assert.Equal(t, http.StatusNotFound, res.StatusCode)
}
