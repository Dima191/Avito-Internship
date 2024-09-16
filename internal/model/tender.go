package model

import "time"

type Tender struct {
	ID              *string
	Name            *string
	Description     *string
	ServiceType     *string
	Status          *string
	OrganizationID  *string
	CreatorUsername *string
	Version         *int
	CreatedAt       *time.Time
}
