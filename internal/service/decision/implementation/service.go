package service_decision_impl

import (
	repository_decision "avito_intership/internal/repository/decision"
	service_decision "avito_intership/internal/service/decision"
	"context"
	"errors"
	"log/slog"
)

type service struct {
	repository repository_decision.Repository

	logger *slog.Logger
}

func (s *service) SubmitDecision(ctx context.Context, authorID string, tenderID string, bidID string, decision string) error {
	if err := s.repository.SubmitDecision(ctx, authorID, tenderID, bidID, decision); err != nil {
		switch {
		case errors.Is(err, repository_decision.ErrUserAlreadyVoted):
			return service_decision.ErrUserAlreadyVoted
		case errors.Is(err, repository_decision.ErrInvalidForeignKey):
			return service_decision.ErrInvalidReference
		default:
			return service_decision.ErrInternal
		}
	}

	return nil
}

func (s *service) DecisionStats(ctx context.Context, bidID string) (applied int, rejected int, err error) {
	applied, rejected, err = s.repository.DecisionStats(ctx, bidID)
	if err != nil {
		switch {
		case errors.Is(err, repository_decision.ErrNoVotes):
			return 0, 0, service_decision.ErrNoVotes
		default:
			return 0, 0, service_decision.ErrInternal
		}
	}
	return applied, rejected, nil
}

func New(repository repository_decision.Repository, logger *slog.Logger) service_decision.Service {
	s := &service{
		repository: repository,
		logger:     logger,
	}

	return s
}
