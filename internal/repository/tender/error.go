package repository_tenders

import "errors"

var (
	ErrInternal             = errors.New("internal error")
	ErrNoTenders            = errors.New("no tender")
	ErrInvalidStatus        = errors.New("invalid status")
	ErrNoSuggestionToUpdate = errors.New("no suggestion to update")
)
