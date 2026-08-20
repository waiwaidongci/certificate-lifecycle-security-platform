package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/acme/certpilot/internal/certificate/application"
	"github.com/acme/certpilot/internal/certificate/domain"
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

func (h *Handler) List(w http.ResponseWriter, r *http.Request) {
	options, err := query.ParseListOptions(r, map[string]bool{"status": true, "service_id": true, "common_name": true, "serial_number": true, "issuer_id": true}, map[string]bool{"common_name": true, "serial_number": true, "status": true, "created_at": true, "updated_at": true, "not_after": true}, "created_at")
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

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.Get(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Issue(w http.ResponseWriter, r *http.Request) {
	var request issueRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		validationErr := apperror.Invalid("invalid request body")
		boundaryErr := fmt.Errorf("decode certificate issue request: %v", validationErr)
		httpx.WriteError(r.Context(), w, boundaryErr)
		return
	}
	item, err := h.service.Issue(r.Context(), request.command())
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteCreated(w, item)
}

func (h *Handler) Revoke(w http.ResponseWriter, r *http.Request) {
	var request revokeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(r.Context(), w, apperror.Invalid("invalid request body"))
		return
	}
	item, err := h.service.Revoke(r.Context(), r.PathValue("id"), request.command())
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
