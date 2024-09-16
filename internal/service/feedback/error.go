package service_feedback

import "errors"

var (
	ErrInternal  = errors.New("internal error")
	ErrNoReviews = errors.New("no reviews")
)
