package handler_bid_model

import (
	"github.com/go-playground/validator/v10"
	"slices"
	"strings"
	"time"
)

type BidResponse struct {
	ID         *string    `json:"id"`
	Name       *string    `json:"name"`
	Status     *string    `json:"status"`
	AuthorType *string    `json:"author_type"`
	AuthorID   *string    `json:"author_id"`
	Version    *int       `json:"version"`
	CreatedAt  *time.Time `json:"created_at"`
}

type BidRequest struct {
	Name        *string `json:"name"`
	Description *string `json:"description"`
	TenderID    *string `json:"tender_id"`
	AuthorType  *string `json:"author_type" validate:"author_type"`
	AuthorID    *string `json:"author_id"`
}

var (
	PossibleAuthorTypes = []string{"organization", "user"}
)

func AuthorTypeValidation(fl validator.FieldLevel) bool {
	return slices.Contains(PossibleAuthorTypes, strings.ToLower(fl.Field().String()))
}
