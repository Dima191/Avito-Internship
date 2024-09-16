package service_bids

import (
	"avito_intership/internal/model"
	"context"
)

type Service interface {
	Create(ctx context.Context, bid model.Bid) (model.Bid, error)
	//BidsByUser returns a list of bids by username (on behalf of the organization and on behalf of the user)
	BidsByUser(ctx context.Context, username string, limit int, offset int) ([]model.Bid, error)
	//BidsByTenderID can use tender creators only
	BidsByTenderID(ctx context.Context, tenderID string, username string, limit int, offset int) ([]model.Bid, error)
	//GetStatus can use tender creators or bid authors
	GetStatus(ctx context.Context, bidID string, username string) (status string, err error)
	//ChangeStatus can use bid creators only
	ChangeStatus(ctx context.Context, bidID string, username string, status string) (bid model.Bid, err error)
	//Edit can use bid creators only
	Edit(ctx context.Context, bidID string, username string, bid model.Bid) (model.Bid, error)
	SubmitDecision(ctx context.Context, bidID string, decision string, username string) (bid model.Bid, isWinner bool, err error)
	Feedback(ctx context.Context, bidID string, username string, feedback string) (model.Bid, error)
	//RollbackVersion can use bid creators only
	RollbackVersion(ctx context.Context, bidID string, username string, version int) (model.Bid, error)
	GetReviews(ctx context.Context, tenderID string, authorUsername, requesterUsername string, limit int, offset int) ([]model.Feedback, error)
}
