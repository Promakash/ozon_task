package usecases

import "ozon_task/domain"

type URL interface {
	ShortenURL(original domain.URL) (domain.URL, error)
	ResolveURL(shorted domain.URL) (domain.URL, error)
}
