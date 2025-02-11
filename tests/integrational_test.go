package tests

import (
	"encoding/json"
	"github.com/stretchr/testify/require"
	"net/http"
	"ozon_task/internal/api/http/types"
	urlshortenerv1 "ozon_task/protos/gen/go"
	"ozon_task/tests/suite"
	"testing"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func Test_PutGRPCGetHTTP(t *testing.T) {
	const originalURL = "https://finance.ozon.ru/promo/factoring/landing"

	ctxGRPC, stGRPC := suite.NewGRPCSuite(t)
	httpResponse, err := stGRPC.URLClient.ShortenURL(
		ctxGRPC,
		&urlshortenerv1.ShortenURLRequest{
			OriginalUrl: originalURL,
		},
	)
	code, _ := status.FromError(err)

	require.NoError(t, err)
	require.Equal(t, codes.OK, code.Code())
	require.NotEmpty(t, httpResponse.GetShortenedUrl())

	ctxHTTP, stHTTP := suite.NewHTTPSuite(t)
	gRPCResponse, err := SendGetRequest(
		ctxHTTP,
		stHTTP.Client,
		stHTTP.BaseURL,
		GetOriginalURL,
		httpResponse.GetShortenedUrl(),
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, gRPCResponse.StatusCode)

	var resolveResponse types.GetOriginalURLResponse
	err = json.NewDecoder(gRPCResponse.Body).Decode(&resolveResponse)
	require.NoError(t, err)

	require.Equal(t, originalURL, resolveResponse.OriginalURL)
}

func Test_PutHTTPGetGRPC(t *testing.T) {
	const originalURL = "http://finance.ozon.ru/promo/factoring/landing"

	ctxHTTP, stHTTP := suite.NewHTTPSuite(t)
	httpResponse, err := SendPostRequest(
		ctxHTTP,
		stHTTP.Client,
		stHTTP.BaseURL,
		PostShortPath,
		types.PostShortURLRequest{OriginalURL: originalURL},
	)
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, httpResponse.StatusCode)

	var postResponse types.PostShortURLResponse
	err = json.NewDecoder(httpResponse.Body).Decode(&postResponse)
	require.NoError(t, err)

	ctxGRPC, stGRPC := suite.NewGRPCSuite(t)
	gRPCResponse, err := stGRPC.URLClient.ResolveURL(
		ctxGRPC,
		&urlshortenerv1.ResolveURLRequest{
			ShortenedUrl: postResponse.ShortenedURL,
		},
	)
	code, _ := status.FromError(err)

	require.NoError(t, err)
	require.Equal(t, codes.OK, code.Code())

	require.Equal(t, originalURL, gRPCResponse.GetOriginalUrl())
}
