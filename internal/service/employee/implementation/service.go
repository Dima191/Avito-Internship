package service_employee_impl

import (
	repository_employee "avito_intership/internal/repository/employee"
	service_employee "avito_intership/internal/service/employee"
	"context"
	"errors"
	"log/slog"
)

type service struct {
	repository repository_employee.Repository

	logger *slog.Logger
}

func (s *service) IDByUsername(ctx context.Context, username string) (userID string, err error) {
	userID, err = s.repository.IDByUsername(ctx, username)
	if err != nil {
		switch {
		case errors.Is(err, repository_employee.ErrNonExistingEmployee):
			return "", service_employee.ErrNonExistingEmployee
		default:
			return "", service_employee.ErrInternal
		}
	}

	return userID, nil
}

func New(repository repository_employee.Repository, logger *slog.Logger) service_employee.Service {
	s := &service{
		repository: repository,
		logger:     logger,
	}

	return s
}
