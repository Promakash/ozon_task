package domain

import "errors"

var (
	ErrOriginalNotFound  = errors.New("no link found by this shortened link")
	ErrShortenedNotFound = errors.New("no link found by this original link")
)
