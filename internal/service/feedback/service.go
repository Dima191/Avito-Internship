package service_feedback

import (
	"avito_intership/internal/model"
	"context"
)

type Service interface {
	Feedback(ctx context.Context, userID string, feedback string) error
	GetReviews(ctx context.Context, authorUsername string, limit int, offset int) ([]model.Feedback, error)
}
