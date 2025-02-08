package domain

import (
	"net/http"
	"strings"
)

var disallowedCodes = map[int]struct{}{
	http.StatusNotFound:   {},
	http.StatusGone:       {},
	http.StatusBadRequest: {},
}

func IsValidOriginalURL(url string) bool {
	if len(url) == 0 {
		return false
	}

	resp, err := http.Get(url)
	if err != nil {
		return false
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode >= 500 {
		return false
	}

	if _, ok := disallowedCodes[resp.StatusCode]; ok {
		return false
	}

	return true
}

func IsValidShortenedURL(url string) bool {
	if len(url) != ShortenedURLSize {
		return false
	}

	for _, val := range url {
		if !strings.ContainsRune(AllowedSymbols, val) {
			return false
		}
	}

	return true
}
