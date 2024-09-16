package handler_tender_model

import (
	"time"
)

type TenderResponse struct {
	ID          *string    `json:"id"`
	Name        *string    `json:"name"`
	Description *string    `json:"description"`
	Status      *string    `json:"status"`
	ServiceType *string    `json:"serviceType"`
	Version     *int       `json:"version"`
	CreatedAt   *time.Time `json:"createdAt"`
}

type TenderRequest struct {
	Name            *string `json:"name"`
	Description     *string `json:"description"`
	ServiceType     *string `json:"serviceType"`
	OrganizationID  *string `json:"organizationId"`
	CreatorUsername *string `json:"creatorUsername"`
}
