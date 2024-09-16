package handler_tender_converter

import (
	handler_tender_model "avito_intership/internal/handlers/tender/model"
	"avito_intership/internal/model"
)

func ToTenderService(tender handler_tender_model.TenderRequest) model.Tender {
	return model.Tender{
		Name:            tender.Name,
		Description:     tender.Description,
		ServiceType:     tender.ServiceType,
		OrganizationID:  tender.OrganizationID,
		CreatorUsername: tender.CreatorUsername,
	}
}

func ToTenderHandler(tender model.Tender) handler_tender_model.TenderResponse {
	return handler_tender_model.TenderResponse{
		ID:          tender.ID,
		Name:        tender.Name,
		Description: tender.Description,
		Status:      tender.Status,
		ServiceType: tender.ServiceType,
		Version:     tender.Version,
		CreatedAt:   tender.CreatedAt,
	}
}

const (
	defaultTendersCh = 5
)

func ArrToTenderHandler(tenders []model.Tender) []handler_tender_model.TenderResponse {
	tenderResp := make([]handler_tender_model.TenderResponse, 0, len(tenders))
	for _, tender := range tenders {
		tenderResp = append(tenderResp, ToTenderHandler(tender))
	}
	return tenderResp
}
