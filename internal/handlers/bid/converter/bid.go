package handler_bid_converter

import (
	handler_bid_model "avito_intership/internal/handlers/bid/model"
	"avito_intership/internal/model"
)

func ToBidService(bid handler_bid_model.BidRequest) model.Bid {
	return model.Bid{
		Name:        bid.Name,
		Description: bid.Description,
		TenderID:    bid.TenderID,
		AuthorType:  bid.AuthorType,
		AuthorID:    bid.AuthorID,
	}
}

func ToBidHandler(bid model.Bid) handler_bid_model.BidResponse {
	return handler_bid_model.BidResponse{
		ID:         bid.ID,
		Name:       bid.Name,
		Status:     bid.Status,
		AuthorType: bid.AuthorType,
		AuthorID:   bid.AuthorID,
		Version:    bid.Version,
		CreatedAt:  bid.CreatedAt,
	}
}

func ArrToBidHandler(bid []model.Bid) []handler_bid_model.BidResponse {
	bidResp := make([]handler_bid_model.BidResponse, 0, len(bid))
	for _, v := range bid {
		bidResp = append(bidResp, ToBidHandler(v))
	}

	return bidResp
}
