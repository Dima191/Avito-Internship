package repository_bid

import "errors"

var (
	ErrInvalidAuthorID = errors.New("invalid author id")
	ErrInvalidTenderID = errors.New("invalid tender id")

	ErrInvalidReq = errors.New("invalid request")

	ErrInvalidBidStatus = errors.New("invalid bid status")

	ErrInternal = errors.New("internal error")

	ErrNoBids               = errors.New("no bid")
	ErrNoSuggestionToUpdate = errors.New("no suggestion to update")
)
