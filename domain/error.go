package domain

import "errors"

var (
	ErrInvalidOriginal      = errors.New("invalid original url")
	ErrInaccessibleOriginal = errors.New("inaccessible original url")
	ErrInvalidShortened     = errors.New("invalid shortened url")
	ErrOriginalNotFound     = errors.New("no link found by this shortened link")
	ErrShortenedNotFound    = errors.New("no link found by this original link")
)
