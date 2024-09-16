package service_tenders

import "errors"

var (
	ErrInternal             = errors.New("internal error")
	ErrNoTenders            = errors.New("no tender")
	ErrNoSuggestionToUpdate = errors.New("no suggestion to update")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrForbidden            = errors.New("forbidden")
)
