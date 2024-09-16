package repository_decision

import "errors"

var (
	ErrInternal          = errors.New("internal error")
	ErrNoVotes           = errors.New("no votes")
	ErrInvalidForeignKey = errors.New("invalid reference to author_id or tender_id")
	ErrUserAlreadyVoted  = errors.New("user already voted")
)
