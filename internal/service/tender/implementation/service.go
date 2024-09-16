package service_tenders_impl

import (
	"avito_intership/internal/model"
	repository_tenders "avito_intership/internal/repository/tender"
	service_employee "avito_intership/internal/service/employee"
	service_organization_resp "avito_intership/internal/service/organization_responsible"
	service_tenders "avito_intership/internal/service/tender"
	"context"
	"errors"
	"log/slog"
)

type service struct {
	repository repository_tenders.Repository

	employeeService         service_employee.Service
	organizationRespService service_organization_resp.Service

	logger *slog.Logger
}

func (s *service) organizationIDAndUserIDByUsername(ctx context.Context, username string) (userID string, organizationID string, err error) {
	userID, err = s.employeeService.IDByUsername(ctx, username)
	if err != nil {
		return "", "", err
	}

	organizationID, err = s.organizationRespService.GetOrganizationIDByRepresentative(ctx, userID)
	if err != nil {
		return "", "", err
	}

	return userID, organizationID, nil
}

func (s *service) TenderOrganizationID(ctx context.Context, tenderID string) (organizationID string, err error) {
	organizationID, err = s.repository.TenderOrganizationID(ctx, tenderID)
	if err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return "", service_tenders.ErrNoTenders
		default:
			return "", service_tenders.ErrInternal
		}
	}

	return organizationID, nil
}

func (s *service) TenderList(ctx context.Context, serviceTypes []string, limit int, offset int) ([]model.Tender, error) {
	tenders, err := s.repository.TenderList(ctx, serviceTypes, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return nil, service_tenders.ErrNoTenders
		default:
			return nil, service_tenders.ErrInternal
		}
	}

	return tenders, nil
}

func (s *service) Create(ctx context.Context, tender model.Tender) (model.Tender, error) {
	tender, err := s.repository.Create(ctx, tender)
	if err != nil {
		return model.Tender{}, service_tenders.ErrInternal
	}

	return tender, nil
}

func (s *service) TendersByUser(ctx context.Context, username string, limit int, offset int) ([]model.Tender, error) {
	tenders, err := s.repository.TendersByUser(ctx, username, limit, offset)
	if err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return nil, service_tenders.ErrNoTenders
		default:
			return nil, service_tenders.ErrInternal
		}
	}

	return tenders, nil
}

func (s *service) TenderStatus(ctx context.Context, tenderID string) (tenderOrganizationID string, status string, err error) {
	tenderOrganizationID, status, err = s.repository.TenderStatus(ctx, tenderID)
	if err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return "", "", service_tenders.ErrNoTenders
		default:
			return "", "", service_tenders.ErrInternal
		}
	}

	return tenderOrganizationID, status, nil
}

func (s *service) ChangeTenderStatusWithUserCheck(ctx context.Context, tenderID string, username string, status string) (model.Tender, error) {
	tender, err := s.repository.ChangeTenderStatusWithUserCheck(ctx, tenderID, username, status)
	if err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrInvalidStatus):
			return model.Tender{}, service_tenders.ErrInvalidStatus
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return model.Tender{}, service_tenders.ErrNoTenders
		default:
			return model.Tender{}, service_tenders.ErrInternal
		}
	}

	return tender, nil
}

func (s *service) ChangeTenderStatusForce(ctx context.Context, tenderID string, status string) error {
	if err := s.repository.ChangeTenderStatusForce(ctx, tenderID, status); err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrInvalidStatus):
			return service_tenders.ErrInvalidStatus
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return service_tenders.ErrNoTenders
		default:
			return service_tenders.ErrInternal
		}
	}

	return nil
}

func (s *service) Edit(ctx context.Context, tenderID string, username string, tender model.Tender) (model.Tender, error) {
	//GET INFO BY USERNAME. CHECK IF IT IS A TENDER OWNER AND APPLY SUGGESTIONS
	userID, err := s.employeeService.IDByUsername(ctx, username)
	if err != nil {
		return model.Tender{}, err
	}

	organizationID, err := s.organizationRespService.GetOrganizationIDByRepresentative(ctx, userID)
	if err != nil {
		return model.Tender{}, err
	}

	//GET TENDER ORGANIZATION ID
	tenderOrganizationID, err := s.repository.TenderOrganizationID(ctx, tenderID)
	if err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return model.Tender{}, service_tenders.ErrNoTenders
		default:
			return model.Tender{}, service_tenders.ErrInternal
		}
	}

	//CHECK ORGANIZATION IDs
	if organizationID != tenderOrganizationID {
		return model.Tender{}, service_tenders.ErrForbidden
	}

	//EDIT
	updatedTender, err := s.repository.Edit(ctx, tenderID, tender)
	if err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrNoSuggestionToUpdate):
			return model.Tender{}, service_tenders.ErrNoSuggestionToUpdate
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return model.Tender{}, service_tenders.ErrNoTenders
		default:
			return model.Tender{}, service_tenders.ErrInternal
		}
	}

	return updatedTender, nil
}

func (s *service) RollbackVersion(ctx context.Context, tenderID string, username string, version int) (model.Tender, error) {
	_, organizationID, err := s.organizationIDAndUserIDByUsername(ctx, username)
	if err != nil {
		return model.Tender{}, err
	}

	tenderOrganizationID, err := s.repository.TenderOrganizationID(ctx, tenderID)
	if err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return model.Tender{}, service_tenders.ErrNoTenders
		default:
			return model.Tender{}, service_tenders.ErrInternal
		}
	}

	if organizationID != tenderOrganizationID {
		return model.Tender{}, service_tenders.ErrForbidden
	}

	tender, err := s.repository.RollbackVersion(ctx, tenderID, version)
	if err != nil {
		switch {
		case errors.Is(err, repository_tenders.ErrNoTenders):
			return model.Tender{}, service_tenders.ErrNoTenders
		default:
			return model.Tender{}, service_tenders.ErrInternal
		}
	}

	return tender, nil
}

func (s *service) ConfirmTenderCreator(ctx context.Context, tenderID string, userOrganizationID string) (exists bool, err error) {
	exists, err = s.repository.ConfirmTenderCreator(ctx, tenderID, userOrganizationID)
	if err != nil {
		return false, service_tenders.ErrInternal
	}

	return exists, nil
}

func New(repository repository_tenders.Repository, employeeService service_employee.Service, organizationRespService service_organization_resp.Service, logger *slog.Logger) service_tenders.Service {
	s := &service{
		repository:              repository,
		employeeService:         employeeService,
		organizationRespService: organizationRespService,
		logger:                  logger,
	}

	return s
}
