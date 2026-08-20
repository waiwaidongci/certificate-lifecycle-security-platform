package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/acme/certpilot/internal/notification/application"
	"github.com/acme/certpilot/internal/notification/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/httpx"
	"github.com/acme/certpilot/internal/shared/query"
)

type Handler struct {
	service *application.Service
}

func NewHandler(service *application.Service) *Handler {
	return &Handler{service: service}
}

func (h *Handler) Scan(w http.ResponseWriter, r *http.Request) {
	var request scanRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		validationErr := apperror.Invalid("invalid request body")
		boundaryErr := fmt.Errorf("decode notification scan request: %v", validationErr)
		httpx.WriteError(r.Context(), w, boundaryErr)
		return
	}
	items, err := h.service.Scan(r.Context(), request.AdvanceDays, request.Channel)
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteCreated(w, items)
}

func (h *Handler) SendPending(w http.ResponseWriter, r *http.Request) {
	items, err := h.service.SendPending(r.Context(), 20)
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, items)
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	options, err := query.ParseListOptions(r, map[string]bool{"status": true, "certificate_id": true, "service_id": true, "channel": true}, map[string]bool{"status": true, "days_left": true, "created_at": true}, "created_at")
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	items, total, err := h.service.List(r.Context(), domain.ListOptions{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Filters: options.Filters, Sort: options.Sort.Field, Order: options.Sort.Direction})
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.ListResponse{Data: items, Meta: httpx.Meta{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Total: total}})
}

type scanRequest struct {
	AdvanceDays int    `json:"advance_days"`
	Channel     string `json:"channel,omitempty"`
}
