package service_feedback_impl

import (
	"avito_intership/internal/model"
	repository_feedback "avito_intership/internal/repository/feedback"
	service_feedback "avito_intership/internal/service/feedback"
	"context"
	"errors"
	"log/slog"
)

type service struct {
	repository repository_feedback.Repository

	logger *slog.Logger
}

func (s *service) Feedback(ctx context.Context, userID string, feedback string) error {
	if err := s.repository.Feedback(ctx, userID, feedback); err != nil {
		return service_feedback.ErrInternal
	}
	return nil
}

func (s *service) GetReviews(ctx context.Context, authorUsername string, limit int, offset int) ([]model.Feedback, error) {
	feedbacks, err := s.repository.GetFeedbacks(ctx, authorUsername, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, repository_feedback.ErrNoReviews):
			return nil, service_feedback.ErrNoReviews
		default:
			return nil, service_feedback.ErrInternal
		}
	}

	return feedbacks, nil
}

func New(repository repository_feedback.Repository, logger *slog.Logger) service_feedback.Service {
	return &service{
		repository: repository,
		logger:     logger,
	}
}
