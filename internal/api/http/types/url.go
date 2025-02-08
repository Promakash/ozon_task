package types

import (
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
		return nil, err
	}

	if ok := domain.IsValidOriginalURL(req.OriginalURL); !ok {
		return nil, domain.ErrInvalidOriginal
	}

	return req, nil
}

type PostShorURLResponse struct {
	ShortenedURL domain.URL `json:"shortened_url"`
}

type GetOriginalURLRequest struct {
	ShortenedURL domain.URL
}

func CreateGetOriginalURLRequest(r *http.Request) (*GetOriginalURLRequest, error) {
	const queryParamName = "shortened"
	url := chi.URLParam(r, queryParamName)

	if ok := domain.IsValidShortenedURL(url); !ok {
		return nil, domain.ErrInvalidShortened
	}

	return &GetOriginalURLRequest{ShortenedURL: url}, nil
}

type GetOriginalURLResponse struct {
	OriginalURL domain.URL
}
