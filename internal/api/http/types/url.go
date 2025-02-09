package types

import (
	"fmt"
	"github.com/go-chi/chi/v5"
	"net/http"
	"ozon_task/domain"
	"ozon_task/pkg/http/handlers"
)

type PostShortURLRequest struct {
	OriginalURL domain.URL `json:"original_url"`
}

func CreatePostShorURLRequest(r *http.Request) (*PostShortURLRequest, error) {
	req := &PostShortURLRequest{}

	if err := handlers.DecodeRequest(r, req); err != nil {
		return nil, fmt.Errorf("CreatePostShorURLRequest: error while unpacking json: %w", domain.ErrInvalidOriginal)
	}

	req.OriginalURL = domain.NormalizeURL(req.OriginalURL)

	if ok, err := domain.IsValidOriginalURL(req.OriginalURL); !ok {
		return nil, fmt.Errorf("CreatePostShorURLRequest: error while validating url: %w", err)
	}

	return req, nil
}

type PostShortURLResponse struct {
	ShortenedURL domain.ShortURL `json:"shortened_url"`
}

type GetOriginalURLRequest struct {
	ShortenedURL domain.ShortURL
}

func CreateGetOriginalURLRequest(r *http.Request) (*GetOriginalURLRequest, error) {
	const queryParamName = "shortened"
	url := chi.URLParam(r, queryParamName)

	if ok, err := domain.IsValidShortenedURL(url); !ok {
		return nil, fmt.Errorf("CreateGetOriginalURLRequest: error while validating url: %w", err)
	}

	return &GetOriginalURLRequest{ShortenedURL: url}, nil
}

type GetOriginalURLResponse struct {
	OriginalURL domain.URL
}
