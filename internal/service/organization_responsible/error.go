package service_organization_resp

import "errors"

var (
	ErrInternal              = errors.New("internal error")
	ErrUserHasNoOrganization = errors.New("the user does not have an organization")
)
