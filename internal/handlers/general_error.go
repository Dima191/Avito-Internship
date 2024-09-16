package handlers

import "errors"

var (
	ErrDecodeBody       = errors.New("failed to decode request body")
	ErrInternal         = errors.New("internal error")
	ErrInvalidURLParams = errors.New("invalid url params")
)
