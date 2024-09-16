package service_bids

import "errors"

var (
	ErrInvalidAuthorID = errors.New("invalid author id")
	ErrInvalidTenderID = errors.New("invalid tender id")
	ErrInvalidReq      = errors.New("invalid request")

	ErrInternal             = errors.New("internal error")
	ErrNoBids               = errors.New("no bid")
	ErrForbidden            = errors.New("forbidden")
	ErrInvalidBidStatus     = errors.New("invalid bid status")
	ErrNoSuggestionToUpdate = errors.New("no suggestion to update")
	ErrTenderClosed         = errors.New("tender has been closed")

	ErrBidBeenRejected = errors.New("bid been rejected")
)
