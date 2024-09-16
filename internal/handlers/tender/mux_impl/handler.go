package handler_tender_mux_impl

import (
	"avito_intership/internal/handlers"
	handler_tender "avito_intership/internal/handlers/tender"
	handler_tender_converter "avito_intership/internal/handlers/tender/converter"
	handler_tender_model "avito_intership/internal/handlers/tender/model"
	"avito_intership/internal/middlewares"
	repository_tenders "avito_intership/internal/repository/tender"
	service_employee "avito_intership/internal/service/employee"
	service_organization_resp "avito_intership/internal/service/organization_responsible"
	service_tenders "avito_intership/internal/service/tender"
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
	router *mux.Router

	service service_tenders.Service

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

func (h *handler) Ping() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}

func (h *handler) Tenders() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		if err := r.ParseForm(); err != nil {
			l.Error("Failed to parse form")
			http.Error(w, "invalid requested url", http.StatusBadRequest)
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

		tenders, err := h.service.TenderList(r.Context(), r.Form[handler_tender.ServiceTypeQueryParam], limit, offset)
		if err != nil {
			switch {
			case errors.Is(err, service_tenders.ErrNoTenders):
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(handler_tender_converter.ArrToTenderHandler(tenders)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) CreateTender() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)
		tenderReq := handler_tender_model.TenderRequest{}
		if err := json.NewDecoder(r.Body).Decode(&tenderReq); err != nil {
			l.Error("Failed to decode body", "error", err.Error())
			http.Error(w, handlers.ErrDecodeBody.Error(), http.StatusBadRequest)
			return
		}

		tender, err := h.service.Create(r.Context(), handler_tender_converter.ToTenderService(tenderReq))
		if err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}

		w.Header().Add("Content-Type", "application/json")
		encoder := json.NewEncoder(w)
		if err = encoder.Encode(handler_tender_converter.ToTenderHandler(tender)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) TenderByUser() http.HandlerFunc {
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

		username := values.Get(handler_tender.UsernameQueryParam)
		if username == "" {
			http.Error(w, "provide username", http.StatusBadRequest)
			return
		}

		tenders, err := h.service.TendersByUser(r.Context(), username, limit, offset)
		if err != nil {
			switch {
			case errors.Is(err, repository_tenders.ErrNoTenders):
				w.WriteHeader(http.StatusNoContent)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(handler_tender_converter.ArrToTenderHandler(tenders)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) GetStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tenderID := mux.Vars(r)[handler_tender.TenderIDUrlPath]
		if err := uuid.Validate(tenderID); err != nil {
			http.Error(w, "invalid tender id", http.StatusBadRequest)
			return
		}

		_, status, err := h.service.TenderStatus(r.Context(), tenderID)
		if err != nil {
			switch {
			case errors.Is(err, service_tenders.ErrNoTenders):
				http.Error(w, "invalid tender id", http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		if _, err = w.Write([]byte(status)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) UpdateStatus() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		tenderID := mux.Vars(r)[handler_tender.TenderIDUrlPath]
		if err := uuid.Validate(tenderID); err != nil {
			http.Error(w, "invalid tender id", http.StatusBadRequest)
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

		username := values.Get(handler_tender.UsernameQueryParam)
		if username == "" {
			http.Error(w, "provide username", http.StatusBadRequest)
			return
		}

		status := values.Get(handler_tender.StatusQueryParam)
		if username == "" {
			http.Error(w, "provide status", http.StatusBadRequest)
			return
		}

		tender, err := h.service.ChangeTenderStatusWithUserCheck(r.Context(), tenderID, username, status)
		if err != nil {
			switch {
			case errors.Is(err, service_tenders.ErrInvalidStatus):
				http.Error(w, "invalid status", http.StatusBadRequest)
				return
			case errors.Is(err, service_tenders.ErrNoTenders):
				http.Error(w, "invalid tender id", http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(handler_tender_converter.ToTenderHandler(tender)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) Edit() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		tenderID := mux.Vars(r)[handler_tender.TenderIDUrlPath]
		if err := uuid.Validate(tenderID); err != nil {
			http.Error(w, "invalid tender id", http.StatusBadRequest)
			return
		}

		tenderReq := handler_tender_model.TenderRequest{}
		if err := json.NewDecoder(r.Body).Decode(&tenderReq); err != nil {
			l.Error("Failed to decode body", "error", err.Error())
			http.Error(w, handlers.ErrDecodeBody.Error(), http.StatusBadRequest)
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

		username := values.Get(handler_tender.UsernameQueryParam)
		if username == "" {
			http.Error(w, "provide username", http.StatusBadRequest)
			return
		}

		tender, err := h.service.Edit(r.Context(), tenderID, username, handler_tender_converter.ToTenderService(tenderReq))
		if err != nil {
			switch {
			case errors.Is(err, service_tenders.ErrForbidden):
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			case errors.Is(err, service_tenders.ErrNoTenders):
				http.Error(w, "invalid tender id", http.StatusBadRequest)
				return
			case errors.Is(err, service_employee.ErrNonExistingEmployee):
				http.Error(w, "invalid username", http.StatusUnauthorized)
				return
			case errors.Is(err, service_organization_resp.ErrUserHasNoOrganization):
				http.Error(w, "user has no organization", http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(handler_tender_converter.ToTenderHandler(tender)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func (h *handler) RollbackVersion() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		l := logger.EndToEndLogging(r.Context(), h.logger)

		tenderID := mux.Vars(r)[handler_tender.TenderIDUrlPath]
		if err := uuid.Validate(tenderID); err != nil {
			http.Error(w, "invalid tender id", http.StatusBadRequest)
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

		versionStr := mux.Vars(r)[handler_tender.VersionPath]
		if versionStr == "" {
			http.Error(w, "invalid versionStr", http.StatusBadRequest)
			return
		}

		version, err := strconv.Atoi(versionStr)
		if err != nil {
			http.Error(w, "invalid versionStr", http.StatusBadRequest)
			return
		}

		username := values.Get(handler_tender.UsernameQueryParam)
		if username == "" {
			http.Error(w, "provide username", http.StatusBadRequest)
			return
		}

		tender, err := h.service.RollbackVersion(r.Context(), tenderID, username, version)
		if err != nil {
			switch {
			case errors.Is(err, service_employee.ErrNonExistingEmployee):
				http.Error(w, "user doesn't exists", http.StatusUnauthorized)
				return
			case errors.Is(err, service_tenders.ErrNoTenders):
				http.Error(w, "invalid version or tender id", http.StatusBadRequest)
				return
			case errors.Is(err, service_tenders.ErrForbidden):
				http.Error(w, http.StatusText(http.StatusForbidden), http.StatusForbidden)
				return
			case errors.Is(err, service_organization_resp.ErrUserHasNoOrganization):
				http.Error(w, "user has no organization", http.StatusBadRequest)
				return
			default:
				http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
				return
			}
		}

		w.Header().Add("Content-Type", "application/json")
		if err = json.NewEncoder(w).Encode(handler_tender_converter.ToTenderHandler(tender)); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
			return
		}
	}
}

func Register(router *mux.Router, service service_tenders.Service, logger *slog.Logger) error {
	h := &handler{
		router:  router,
		service: service,
		logger:  logger,
	}

	apiRouter := router.PathPrefix("/api").Subrouter()

	apiRouter.Use(middlewares.Log(h.logger))

	apiRouter.Path("/ping").Methods(http.MethodGet).Handler(h.Ping())

	apiRouter.Path("/tenders").Methods(http.MethodGet).Handler(h.Tenders())
	apiRouter.Path("/tenders/new").Methods(http.MethodPost).Handler(h.CreateTender())
	apiRouter.Path("/tenders/my").Methods(http.MethodGet).Handler(h.TenderByUser())
	apiRouter.Path("/tenders/{tender_id}/status").Methods(http.MethodGet).Handler(h.GetStatus())
	apiRouter.Path("/tenders/{tender_id}/status").Methods(http.MethodPut).Handler(h.UpdateStatus())
	apiRouter.Path("/tenders/{tender_id}/edit").Methods(http.MethodPatch).Handler(h.Edit())
	apiRouter.Path("/tenders/{tender_id}/rollback/{version}").Methods(http.MethodPut).Handler(h.RollbackVersion())

	return nil
}
