package domain

import (
	"fmt"
	"net/url"
	"strings"
)

func NormalizeURL(url URL) URL {
	const prefixSecured = "https://"
	const prefixInsecure = "http://"

	if len(url) != 0 && !strings.HasPrefix(url, prefixInsecure) && !strings.HasPrefix(url, prefixSecured) {
		url = fmt.Sprintf("%s%s", prefixSecured, url)
	}

	return url
}

func IsValidOriginalURL(original URL) (bool, error) {
	if len(original) == 0 {
		return false, fmt.Errorf("IsValidOriginalURL: empty string: %w", ErrInvalidOriginal)
	}

	rawURL, err := url.ParseRequestURI(original)
	if err != nil || len(rawURL.Host) == 0 {
		return false, fmt.Errorf("IsValidOriginalURL: invalid format: %w", ErrInvalidOriginal)
	}

	// idx must be < 1 because Request can be https://.ru
	if idx := strings.LastIndex(rawURL.Host, "."); idx < 1 {
		return false, fmt.Errorf("IsValidOriginalURL: no top level domain found: %w", ErrInvalidOriginal)
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
