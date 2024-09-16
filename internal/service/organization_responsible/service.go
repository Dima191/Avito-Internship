package service_organization_resp

import "context"

type Service interface {
	GetOrganizationIDByRepresentative(ctx context.Context, userID string) (organizationID string, err error)
	OrganizationRepresentativesAmount(ctx context.Context, organizationID string) (amount int, err error)
}
