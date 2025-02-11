package grpc

import (
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"ozon_task/domain"
	"ozon_task/pkg/random"
	urlshortenerv1 "ozon_task/protos/gen/go"
	"ozon_task/tests/suite"
	"testing"
)

func TestShortenURL_SuccessURLs(t *testing.T) {
	t.Parallel()
	ctx, st := suite.New(t)

	tests := []struct {
		name           string
		url            domain.URL
		expectedStatus codes.Code
	}{
		{
			name:           "Fully valid url",
			url:            "https://finance.ozon.ru/",
			expectedStatus: codes.OK,
		},
		{
			name:           "Insecure HTTP scheme",
			url:            "http://finance.ozon.ru/",
			expectedStatus: codes.OK,
		},
		{
			name:           "No HTTP Scheme",
			url:            "finance.ozon.ru",
			expectedStatus: codes.OK,
		},
	}

	for _, test := range tests {
		res, err := st.URLClient.ShortenURL(ctx, &urlshortenerv1.ShortenURLRequest{
			OriginalUrl: test.url,
		})
		code, _ := status.FromError(err)

		if err != nil {
			t.Fatalf("error: %v", err)
		}
		assert.NoError(t, err)
		assert.Equal(t, test.expectedStatus, code.Code())
		assert.NotEmpty(t, res.GetShortenedUrl())
	}
}

func TestShortenURL_SameURL(t *testing.T) {
	t.Parallel()
	const validURL = "https://finance.ozon.ru/business/rko"

	ctx, st := suite.New(t)

	res, err := st.URLClient.ShortenURL(ctx, &urlshortenerv1.ShortenURLRequest{
		OriginalUrl: validURL,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, res.GetShortenedUrl())

	resSecond, err := st.URLClient.ShortenURL(ctx, &urlshortenerv1.ShortenURLRequest{
		OriginalUrl: validURL,
	})

	assert.NoError(t, err)
	assert.NotEmpty(t, resSecond.GetShortenedUrl())

	assert.Equal(t, res.GetShortenedUrl(), resSecond.GetShortenedUrl())
}

func TestShortenURL_InvalidURLs(t *testing.T) {
	t.Parallel()
	ctx, st := suite.New(t)

	tests := []struct {
		name           string
		url            domain.URL
		expectedStatus codes.Code
	}{
		{
			name:           "No top level domain",
			url:            "https://ozon",
			expectedStatus: codes.InvalidArgument,
		},
		{
			name:           "Empty URL",
			url:            "",
			expectedStatus: codes.InvalidArgument,
		},
		{
			name:           "No host no top level domain",
			url:            "https://",
			expectedStatus: codes.InvalidArgument,
		},
		{
			name:           "No host",
			url:            "https://.ru",
			expectedStatus: codes.InvalidArgument,
		},
		{
			name:           "Whitespace in host",
			url:            "https://oz on.ru",
			expectedStatus: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		_, err := st.URLClient.ShortenURL(ctx, &urlshortenerv1.ShortenURLRequest{
			OriginalUrl: test.url,
		})
		code, _ := status.FromError(err)

		assert.Error(t, err)
		assert.Equal(t, test.expectedStatus, code.Code())
	}
}

func TestResolveURL_Success(t *testing.T) {
	t.Parallel()
	const validURL = "https://finance.ozon.ru/business/acquiring"

	ctx, st := suite.New(t)

	res, err := st.URLClient.ShortenURL(ctx, &urlshortenerv1.ShortenURLRequest{
		OriginalUrl: validURL,
	})
	code, _ := status.FromError(err)

	assert.NoError(t, err)
	assert.Equal(t, codes.OK, code.Code())
	assert.NotEmpty(t, res.GetShortenedUrl())

	resolveResp, err := st.URLClient.ResolveURL(ctx, &urlshortenerv1.ResolveURLRequest{
		ShortenedUrl: res.GetShortenedUrl(),
	})
	code, _ = status.FromError(err)

	assert.NoError(t, err)
	assert.Equal(t, codes.OK, code.Code())
	assert.NotEmpty(t, resolveResp.GetOriginalUrl())
}

func TestShortenURL_InvalidShorts(t *testing.T) {
	t.Parallel()
	ctx, st := suite.New(t)

	tests := []struct {
		name           string
		url            domain.ShortURL
		expectedStatus codes.Code
	}{
		{
			name:           "Invalid length",
			url:            "xZsyc_xvo11",
			expectedStatus: codes.InvalidArgument,
		},
		{
			name:           "Empty URL",
			url:            "",
			expectedStatus: codes.InvalidArgument,
		},
		{
			name:           "Invalid symbol",
			url:            "@fcizawmtN",
			expectedStatus: codes.InvalidArgument,
		},
	}

	for _, test := range tests {
		_, err := st.URLClient.ResolveURL(ctx, &urlshortenerv1.ResolveURLRequest{
			ShortenedUrl: test.url,
		})
		code, _ := status.FromError(err)

		assert.Error(t, err)
		assert.Equal(t, test.expectedStatus, code.Code())
	}
}

func TestShortenURL_NotFound(t *testing.T) {
	t.Parallel()
	validShort, err := random.NewRandomString(domain.ShortenedURLSize, domain.AllowedSymbols)
	assert.NoError(t, err)

	ctx, st := suite.New(t)

	_, err = st.URLClient.ResolveURL(ctx, &urlshortenerv1.ResolveURLRequest{
		ShortenedUrl: validShort,
	})
	code, _ := status.FromError(err)

	assert.Error(t, err)
	assert.Equal(t, codes.NotFound, code.Code())
}
