package service_decision

import "errors"

var (
	ErrInternal         = errors.New("internal error")
	ErrInvalidReference = errors.New("invalid reference to author_id or tender_id")
	ErrUserAlreadyVoted = errors.New("user already voted")
	ErrNoVotes          = errors.New("no votes")
)
