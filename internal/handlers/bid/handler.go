package handler_bid

import "net/http"

type Handler interface {
	Create() http.HandlerFunc
	BidsByUser() http.HandlerFunc
	BidsByTenderID() http.HandlerFunc
	GetStatus() http.HandlerFunc
	ChangeStatus() http.HandlerFunc
	Edit() http.HandlerFunc
	SubmitDecision() http.HandlerFunc
	Feedback() http.HandlerFunc
	RollbackVersion() http.HandlerFunc
}
