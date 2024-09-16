package repository_tender_converter

import (
	"avito_intership/internal/model"
	repository_tender_model "avito_intership/internal/repository/tender/model"
)

func ToTenderFromRepository(tender repository_tender_model.Tender) model.Tender {
	return model.Tender{
		ID:          &tender.ID,
		Name:        &tender.Name,
		Description: &tender.Description,
		ServiceType: &tender.ServiceType,
		Status:      &tender.Status,
		Version:     &tender.Version,
		CreatedAt:   &tender.CreatedAt,
	}
}
