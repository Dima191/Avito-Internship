package repository_bid_converter

import (
	"avito_intership/internal/model"
	repository_bid_model "avito_intership/internal/repository/bid/model"
)

func ToBidFromRepository(bid repository_bid_model.Bid) model.Bid {
	return model.Bid{
		ID:         &bid.ID,
		Name:       &bid.Name,
		Status:     &bid.Status,
		AuthorType: &bid.AuthorType,
		AuthorID:   &bid.AuthorID,
		Version:    &bid.Version,
		CreatedAt:  &bid.CreatedAt,
	}
}
