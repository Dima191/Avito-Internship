package repository_decision

import "context"

type Repository interface {
	SubmitDecision(ctx context.Context, authorID string, tenderID string, bidID string, decision string) error
	DecisionStats(ctx context.Context, bidID string) (applied int, rejected int, err error)
	CloseConn()
}
