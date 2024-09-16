package service_organization_resp_impl

import (
	repository_organization_resp "avito_intership/internal/repository/organization_responsible"
	service_organization_resp "avito_intership/internal/service/organization_responsible"
	"context"
	"errors"
	"log/slog"
)

type service struct {
	repository repository_organization_resp.Repository
	logger     *slog.Logger
}

func (s *service) GetOrganizationIDByRepresentative(ctx context.Context, userID string) (organizationID string, err error) {
	organizationID, err = s.repository.GetOrganizationIDByRepresentative(ctx, userID)
	if err != nil {
		switch {
		case errors.Is(err, repository_organization_resp.ErrUserHasNoOrganization):
			return "", service_organization_resp.ErrUserHasNoOrganization
		default:
			return "", service_organization_resp.ErrInternal
		}
	}

	return organizationID, nil
}

func (s *service) OrganizationRepresentativesAmount(ctx context.Context, organizationID string) (amount int, err error) {
	amount, err = s.repository.OrganizationRepresentativesAmount(ctx, organizationID)
	if err != nil {
		return 0, service_organization_resp.ErrInternal
	}

	return amount, nil
}

func New(repository repository_organization_resp.Repository, logger *slog.Logger) service_organization_resp.Service {
	s := &service{
		repository: repository,
		logger:     logger,
	}

	return s
}
