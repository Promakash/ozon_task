package domain

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var disallowedCodes = map[int]struct{}{
	http.StatusNotFound:   {},
	http.StatusGone:       {},
	http.StatusBadRequest: {},
}

func IsValidOriginalURL(urlStr URL) (bool, error) {
	parsed, err := url.Parse(urlStr)
	if err != nil || len(parsed.Host) == 0 {
		return false, ErrInvalidOriginal
	}

	req, err := http.NewRequest("HEAD", urlStr, nil)
	if err != nil {
		return false, fmt.Errorf("IsValidOriginalURL: error creating request: %w", err)
	}

	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			if len(via) >= 10 {
				return http.ErrUseLastResponse
			}
			return nil
		},
		Timeout: 2 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return false, fmt.Errorf("IsValidOriginalURL: error while making HTTP request: %w", ErrInaccessibleOriginal)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 500 {
		return false, fmt.Errorf("IsValidOriginalURL: got server side error %d: %w", resp.StatusCode, ErrInaccessibleOriginal)
	}

	if _, ok := disallowedCodes[resp.StatusCode]; ok {
		return false, fmt.Errorf("IsValidOriginalURL: got disallowedCode %d: %w", resp.StatusCode, ErrInaccessibleOriginal)
	}

	return true, nil
}

func IsValidShortenedURL(url ShortURL) (bool, error) {
	if len(url) != ShortenedURLSize {
		return false,
			fmt.Errorf("IsValidShortenedURL: got unexpected size %d, wanted: %d: %w", len(url), ShortenedURLSize, ErrInvalidShortened)
	}

	for _, val := range url {
		if !strings.ContainsRune(AllowedSymbols, val) {
			return false,
				fmt.Errorf("IsValidShortenedURL: got unexpected token %q: %w", val, ErrInvalidShortened)
		}
	}

	return true, nil
}

func NormalizeURL(url URL) URL {
	const prefix = "https://"

	if len(url) != 0 && !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		url = fmt.Sprintf("%s%s", prefix, url)
	} else if strings.HasPrefix(url, "http://") {
		trimmed := strings.TrimPrefix(url, "http://")
		url = fmt.Sprintf("%s%s", prefix, trimmed)
	}

	return url
}
