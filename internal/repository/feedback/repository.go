package repository_feedback

import (
	"avito_intership/internal/model"
	"context"
)

type Repository interface {
	Feedback(ctx context.Context, userID string, feedback string) error
	GetFeedbacks(ctx context.Context, authorUsername string, limit int, offset int) ([]model.Feedback, error)
	CloseConn()
}
