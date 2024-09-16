package service_tenders

import (
	"avito_intership/internal/model"
	"context"
)

type Service interface {
	TenderOrganizationID(ctx context.Context, tenderID string) (organizationID string, err error)
	TenderList(ctx context.Context, serviceTypes []string, limit int, offset int) ([]model.Tender, error)
	Create(ctx context.Context, tender model.Tender) (model.Tender, error)
	TendersByUser(ctx context.Context, username string, limit int, offset int) ([]model.Tender, error)
	TenderStatus(ctx context.Context, tenderID string) (tenderOrganizationID string, status string, err error)
	ChangeTenderStatusWithUserCheck(ctx context.Context, tenderID string, username string, status string) (model.Tender, error)
	ChangeTenderStatusForce(ctx context.Context, tenderID string, status string) error
	Edit(ctx context.Context, tenderID string, username string, tender model.Tender) (model.Tender, error)
	RollbackVersion(ctx context.Context, tenderID string, username string, version int) (model.Tender, error)
	ConfirmTenderCreator(ctx context.Context, tenderID string, userOrganizationID string) (exists bool, err error)
}
