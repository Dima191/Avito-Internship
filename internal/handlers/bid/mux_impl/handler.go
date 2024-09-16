package handler_bid_mux_impl

import (
	"avito_intership/internal/handlers"
	handler_bid "avito_intership/internal/handlers/bid"
	handler_bid_converter "avito_intership/internal/handlers/bid/converter"
	handler_bid_model "avito_intership/internal/handlers/bid/model"
	handler_tender "avito_intership/internal/handlers/tender"
	"avito_intership/internal/middlewares"
	"avito_intership/internal/model"
	repository_feedback "avito_intership/internal/repository/feedback"
	repository_tenders "avito_intership/internal/repository/tender"
	service_bids "avito_intership/internal/service/bid"
	service_decision "avito_intership/internal/service/decision"
	service_employee "avito_intership/internal/service/employee"
	service_organization_resp "avito_intership/internal/service/organization_responsible"
	"avito_intership/internal/validator"
	"avito_intership/pkg/logger"
	"encoding/json"
	"errors"
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
)

type handler struct {
	router  *mux.Router
	service service_bids.Service

	validator *validator.Validate

	logger *slog.Logger
}

func (h *handler) parseURL(requestedURI string, l *slog.Logger) (url.Values, error) {
	u, err := url.Parse(requestedURI)
	if err != nil {
		l.Error("Failed to parse request URI", slog.String("error", err.Error()))
		return nil, handlers.ErrInternal
	}

	values, err := url.ParseQuery(u.RawQuery)
	if err != nil {
		l.Error("Failed to parse query parameters", slog.String("error", err.Error()))
		return nil, handlers.ErrInvalidURLParams
	}

	return values, nil
}

func (h *handler) getLimitAndOffsetQueryParams(limitStr, offsetStr string) (limit, offset int) {
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limitStr == "" {
		limit = handlers.DefaultLimit
	}

	offset, err = strconv.Atoi(offsetStr)
	if err != nil || offsetStr == "" {
		offset = handlers.DefaultOffset
	}

	return limit, offset
}

func (h *handler) Create() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		bidReq := handler_bid_model.BidRequest{}
		if err := json.NewDecoder(r.Body).Decode(&bidReq); err != nil {
			l.Error("Failed to decode body", "error", err.Error())
			http.Error(w, handlers.ErrDecodeBody.Error(), http.StatusBadRequest)
			return
		}

		bid, err := h.service.Create(r.Context(), handler_bid_converter.ToBidService(bidReq))
		if err != nil {
			switch {
			case errors.Is(err, service_bids.ErrInvalidReq):
				http.Error(w, "invalid request", http.StatusBadRequest)
				return
			case errors.Is(err, service_bids.ErrInvalidAuthorID):
				http.Error(w, "invalid author_id", http.StatusBadRequest)
				return
			case errors.Is(err, service_bids.ErrInvalidTenderID):
				http.Error(w, "invalid tender_id", http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}
		w.Header().Add("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		if err = json.NewEncoder(w).Encode(handler_bid_converter.ToBidHandler(bid)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) BidsByUser() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		values, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		limit, offset := h.getLimitAndOffsetQueryParams(values.Get(handlers.LimitQueryParam), values.Get(handlers.OffsetQueryParam))
		username := values.Get(handler_bid.UsernameQueryParam)
		if username == "" {
			http.Error(w, errors.New("provide username").Error(), http.StatusUnauthorized)
			return
		}

		bids, err := h.service.BidsByUser(r.Context(), username, limit, offset)
		if err != nil {
			switch {
			case errors.Is(err, service_bids.ErrNoBids):
				http.Error(w, service_bids.ErrNoBids.Error(), http.StatusNoContent)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err = encoder.Encode(handler_bid_converter.ArrToBidHandler(bids)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) BidsByTenderID() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		values, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		limit, offset := h.getLimitAndOffsetQueryParams(values.Get(handlers.LimitQueryParam), values.Get(handlers.OffsetQueryParam))

		tenderID := mux.Vars(r)[handler_bid.TenderIDUrlPath]
		if err = uuid.Validate(tenderID); err != nil {
			http.Error(w, "invalid tender id", http.StatusBadRequest)
			return
		}

		username := values.Get(handler_bid.UsernameQueryParam)
		if username == "" {
			http.Error(w, "provide username", http.StatusBadRequest)
			return
		}

		bids, err := h.service.BidsByTenderID(r.Context(), tenderID, username, limit, offset)
		if err != nil {
			switch {
			case errors.Is(err, service_employee.ErrNonExistingEmployee):
				http.Error(w, service_employee.ErrNonExistingEmployee.Error(), http.StatusUnauthorized)
				return
			case errors.Is(err, service_bids.ErrForbidden):
				http.Error(w, "permission denied", http.StatusForbidden)
				return
			case errors.Is(err, service_bids.ErrNoBids):
				http.Error(w, service_bids.ErrNoBids.Error(), http.StatusNoContent)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err = encoder.Encode(handler_bid_converter.ArrToBidHandler(bids)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) GetStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		bidID := mux.Vars(r)[handler_bid.BidIDUrlPath]
		if err := uuid.Validate(bidID); err != nil {
			http.Error(w, "invalid bid id", http.StatusBadRequest)
			return
		}

		values, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		username := values.Get(handler_bid.UsernameQueryParam)
		if username == "" {
			http.Error(w, "provide username", http.StatusBadRequest)
			return
		}

		status, err := h.service.GetStatus(r.Context(), bidID, username)
		if err != nil {
			switch {
			case errors.Is(err, service_employee.ErrNonExistingEmployee):
				http.Error(w, service_employee.ErrNonExistingEmployee.Error(), http.StatusUnauthorized)
				return
			case errors.Is(err, service_bids.ErrNoBids):
				http.Error(w, service_bids.ErrNoBids.Error(), http.StatusNoContent)
				return
			case errors.Is(err, service_bids.ErrForbidden):
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		if err = json.NewEncoder(w).Encode(status); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) ChangeStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		bidID := mux.Vars(r)[handler_bid.BidIDUrlPath]
		if err := uuid.Validate(bidID); err != nil {
			http.Error(w, "invalid bid id", http.StatusBadRequest)
			return
		}

		values, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		status := values.Get(handler_bid.StatusQueryParam)
		if status == "" {
			http.Error(w, "provide status", http.StatusBadRequest)
			return
		}

		username := values.Get(handler_bid.UsernameQueryParam)
		if username == "" {
			http.Error(w, "provide username", http.StatusBadRequest)
			return
		}

		bid, err := h.service.ChangeStatus(r.Context(), bidID, username, status)
		if err != nil {
			switch {
			case errors.Is(err, service_bids.ErrInvalidBidStatus):
				http.Error(w, "invalid bid status", http.StatusBadRequest)
				return
			case errors.Is(err, service_employee.ErrNonExistingEmployee):
				http.Error(w, service_employee.ErrNonExistingEmployee.Error(), http.StatusUnauthorized)
				return
			case errors.Is(err, service_bids.ErrNoBids):
				http.Error(w, service_bids.ErrNoBids.Error(), http.StatusBadRequest)
				return
			case errors.Is(err, service_bids.ErrForbidden):
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		if err = json.NewEncoder(w).Encode(handler_bid_converter.ToBidHandler(bid)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) Edit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		bid := model.Bid{}

		if err := json.NewDecoder(r.Body).Decode(&bid); err != nil {
			l.Error("Failed to decode body", "error", err.Error())
			http.Error(w, handlers.ErrDecodeBody.Error(), http.StatusBadRequest)
			return
		}

		bidID := mux.Vars(r)[handler_bid.BidIDUrlPath]
		if err := uuid.Validate(bidID); err != nil {
			http.Error(w, "invalid bid id", http.StatusBadRequest)
			return
		}

		values, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		username := values.Get(handler_bid.UsernameQueryParam)
		if username == "" {
			http.Error(w, "provide username", http.StatusBadRequest)
			return
		}

		updatedBid, err := h.service.Edit(r.Context(), bidID, username, bid)
		if err != nil {
			switch {
			case errors.Is(err, service_bids.ErrNoSuggestionToUpdate):
				http.Error(w, service_bids.ErrNoSuggestionToUpdate.Error(), http.StatusBadRequest)
				return
			case errors.Is(err, service_bids.ErrNoBids):
				http.Error(w, service_bids.ErrNoBids.Error(), http.StatusBadRequest)
				return
			case errors.Is(err, service_bids.ErrForbidden):
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		if err = json.NewEncoder(w).Encode(handler_bid_converter.ToBidHandler(updatedBid)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) SubmitDecision() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		bidID := mux.Vars(r)[handler_bid.BidIDUrlPath]
		if err := uuid.Validate(bidID); err != nil {
			http.Error(w, "invalid bid id", http.StatusBadRequest)
			return
		}

		values, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		decision := values.Get(handler_bid.DecisionQueryParam)
		if decision == "" {
			http.Error(w, "provide decision", http.StatusBadRequest)
			return
		}

		username := values.Get(handler_bid.UsernameQueryParam)
		if username == "" {
			http.Error(w, errors.New("provide username").Error(), http.StatusUnauthorized)
			return
		}

		bid, isWinner, err := h.service.SubmitDecision(r.Context(), bidID, decision, username)
		if err != nil {
			switch {
			case errors.Is(err, service_decision.ErrUserAlreadyVoted):
				http.Error(w, "can not vote twice", http.StatusBadRequest)
				return
			case errors.Is(err, service_bids.ErrTenderClosed):
				http.Error(w, "tender has been closed", http.StatusBadRequest)
				return
			case errors.Is(err, service_bids.ErrBidBeenRejected):
				http.Error(w, "bid been rejected", http.StatusBadRequest)
				return
			case errors.Is(err, service_employee.ErrNonExistingEmployee):
				http.Error(w, service_employee.ErrNonExistingEmployee.Error(), http.StatusUnauthorized)
				return
			case errors.Is(err, service_organization_resp.ErrUserHasNoOrganization):
				http.Error(w, service_organization_resp.ErrUserHasNoOrganization.Error(), http.StatusUnauthorized)
				return
			case errors.Is(err, service_bids.ErrForbidden):
				http.Error(w, service_bids.ErrForbidden.Error(), http.StatusForbidden)
				return
			case errors.Is(err, service_decision.ErrInvalidReference):
				http.Error(w, service_decision.ErrInvalidReference.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		if isWinner {
			if _, err = w.Write([]byte("success: tender closed, contractor found\n")); err != nil {
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		en := json.NewEncoder(w)
		if err = en.Encode(handler_bid_converter.ToBidHandler(bid)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) Feedback() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		bidID := mux.Vars(r)[handler_bid.BidIDUrlPath]
		if err := uuid.Validate(bidID); err != nil {
			http.Error(w, "invalid bid id", http.StatusBadRequest)
			return
		}

		values, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		feedback := values.Get(handler_bid.FeedbackQueryParam)
		if feedback == "" {
			http.Error(w, "provide feedback", http.StatusBadRequest)
			return
		}

		username := values.Get(handler_bid.UsernameQueryParam)
		if username == "" {
			http.Error(w, errors.New("provide username").Error(), http.StatusUnauthorized)
			return
		}

		bid, err := h.service.Feedback(r.Context(), bidID, username, feedback)
		if err != nil {
			switch {
			case errors.Is(err, service_bids.ErrForbidden):
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			case errors.Is(err, service_bids.ErrNoBids) || errors.Is(err, service_organization_resp.ErrUserHasNoOrganization):
				http.Error(w, "invalid bid_id", http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err = encoder.Encode(handler_bid_converter.ToBidHandler(bid)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) RollbackVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		bidID := mux.Vars(r)[handler_bid.BidIDUrlPath]
		if err := uuid.Validate(bidID); err != nil {
			http.Error(w, "invalid bid id", http.StatusBadRequest)
			return
		}

		versionStr := mux.Vars(r)[handler_bid.VersionPath]
		if versionStr == "" {
			http.Error(w, "invalid version", http.StatusBadRequest)
			return
		}

		version, err := strconv.Atoi(versionStr)
		if err != nil {
			http.Error(w, "invalid version", http.StatusBadRequest)
			return
		}

		values, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		username := values.Get(handler_bid.UsernameQueryParam)
		if username == "" {
			http.Error(w, errors.New("provide username").Error(), http.StatusUnauthorized)
			return
		}

		bid, err := h.service.RollbackVersion(r.Context(), bidID, username, version)
		if err != nil {
			switch {
			case errors.Is(err, service_bids.ErrNoBids):
				http.Error(w, "invalid version", http.StatusBadRequest)
				return
			case errors.Is(err, service_employee.ErrNonExistingEmployee):
				http.Error(w, "invalid username", http.StatusBadRequest)
				return
			case errors.Is(err, service_bids.ErrForbidden):
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			case errors.Is(err, service_organization_resp.ErrUserHasNoOrganization):
				http.Error(w, "user even has no organization", http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err = encoder.Encode(handler_bid_converter.ToBidHandler(bid)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) Reviews() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		tenderID := mux.Vars(r)[handler_tender.TenderIDUrlPath]
		if err := uuid.Validate(tenderID); err != nil {
			http.Error(w, "invalid tender id", http.StatusBadRequest)
			return
		}

		value, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		authorUsername := value.Get(handler_bid.AuthorUsernameQueryParam)
		if authorUsername == "" {
			http.Error(w, "invalid author username", http.StatusBadRequest)
			return
		}

		requesterUsername := value.Get(handler_bid.RequesterUsernameQueryParam)
		if requesterUsername == "" {
			http.Error(w, "invalid requester username", http.StatusBadRequest)
			return
		}

		values, err := h.parseURL(r.RequestURI, l)
		if err != nil {
			switch {
			case errors.Is(err, handlers.ErrInvalidURLParams):
				http.Error(w, handlers.ErrInvalidURLParams.Error(), http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		limit, offset := h.getLimitAndOffsetQueryParams(values.Get(handlers.LimitQueryParam), values.Get(handlers.OffsetQueryParam))

		reviews, err := h.service.GetReviews(r.Context(), tenderID, authorUsername, requesterUsername, limit, offset)
		if err != nil {
			switch {
			case errors.Is(err, service_employee.ErrNonExistingEmployee):
				http.Error(w, "invalid username", http.StatusUnauthorized)
				return
			case errors.Is(err, service_organization_resp.ErrUserHasNoOrganization):
				http.Error(w, "user even has no organization", http.StatusBadRequest)
				return
			case errors.Is(err, repository_tenders.ErrNoTenders):
				http.Error(w, "invalid tender id", http.StatusBadRequest)
				return
			case errors.Is(err, service_bids.ErrForbidden):
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			case errors.Is(err, repository_feedback.ErrNoReviews):
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err = encoder.Encode(reviews); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func Register(router *mux.Router, service service_bids.Service, logger *slog.Logger) error {
	h := &handler{
		router:    router,
		service:   service,
		validator: validator.New(),
		logger:    logger,
	}

	if err := h.validator.RegisterTag(validator.AuthorTypeTag, handler_bid_model.AuthorTypeValidation); err != nil {
		logger.Error("Failed to register moderation status validation", "error", err.Error())
		return err
	}

	apiRouter := router.PathPrefix("/api").Subrouter()

	apiRouter.Use(middlewares.Log(h.logger))

	apiRouter.Path("/bids/new").Methods(http.MethodPost).Handler(h.Create())
	apiRouter.Path("/bids/my").Methods(http.MethodGet).Handler(h.BidsByUser())
	apiRouter.Path("/bids/{tender_id}/list").Methods(http.MethodGet).Handler(h.BidsByTenderID())
	apiRouter.Path("/bids/{bid_id}/status").Methods(http.MethodGet).Handler(h.GetStatus())
	apiRouter.Path("/bids/{bid_id}/status").Methods(http.MethodPut).Handler(h.ChangeStatus())
	apiRouter.Path("/bids/{bid_id}/edit").Methods(http.MethodPatch).Handler(h.Edit())
	apiRouter.Path("/bids/{bid_id}/submit_decision").Methods(http.MethodPut).Handler(h.SubmitDecision())
	apiRouter.Path("/bids/{bid_id}/feedback").Methods(http.MethodPut).Handler(h.Feedback())
	apiRouter.Path("/bids/{bid_id}/rollback/{version}").Methods(http.MethodPut).Handler(h.RollbackVersion())
	apiRouter.Path("/bids/{tender_id}/reviews").Methods(http.MethodGet).Handler(h.Reviews())

	return nil
}
