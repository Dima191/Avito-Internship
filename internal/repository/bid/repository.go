package repository_bid

import (
	"avito_intership/internal/model"
	"context"
)

type Repository interface {
	Create(ctx context.Context, bid model.Bid) (model.Bid, error)
	BidsByTenderID(ctx context.Context, tenderID string, limit int, offset int) (bids []model.Bid, err error)
	BidsByAuthorID(ctx context.Context, userID, organizationID string, limit int, offset int) ([]model.Bid, error)
	GetStatus(ctx context.Context, bidID string) (status string, tenderID string, authorID string, err error)
	ChangeStatus(ctx context.Context, bidID string, status string) (bid model.Bid, err error)
	Edit(ctx context.Context, bidID string, bid model.Bid) (updatedBid model.Bid, err error)
	BidTenderID(ctx context.Context, bidID string) (tenderID string, err error)
	BidAuthorID(ctx context.Context, bidID string) (authorID string, err error)
	BidByID(ctx context.Context, bidID string) (model.Bid, error)
	RollbackVersion(ctx context.Context, bidID string, version int) (model.Bid, error)
	CloseConn()
}
