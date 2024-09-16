package handler_tender

import (
	"net/http"
)

type Handler interface {
	TenderOrganizationID() http.HandlerFunc
	TenderList() http.HandlerFunc
	Create() http.HandlerFunc
	TendersByUser() http.HandlerFunc
	TenderStatus() http.HandlerFunc
	ChangeTenderStatus() http.HandlerFunc
	Edit() http.HandlerFunc
}
