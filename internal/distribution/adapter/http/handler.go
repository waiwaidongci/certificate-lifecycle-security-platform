package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/acme/certpilot/internal/distribution/application"
	"github.com/acme/certpilot/internal/distribution/domain"
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

func (h *Handler) ListTemplates(w http.ResponseWriter, r *http.Request) {
	options, err := query.ParseListOptions(r, map[string]bool{"name": true}, map[string]bool{"name": true, "created_at": true, "updated_at": true}, "created_at")
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	items, total, err := h.service.ListTemplates(r.Context(), domain.ListOptions{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Filters: options.Filters, Sort: options.Sort.Field, Order: options.Sort.Direction})
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.ListResponse{Data: items, Meta: httpx.Meta{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Total: total}})
}

func (h *Handler) GetTemplate(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetTemplate(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) CreateTemplate(w http.ResponseWriter, r *http.Request) {
	var request templateCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		validationErr := apperror.Invalid("invalid request body")
		boundaryErr := fmt.Errorf("decode template create request: %v", validationErr)
		httpx.WriteError(r.Context(), w, boundaryErr)
		return
	}
	item, err := h.service.CreateTemplate(r.Context(), request.command())
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteCreated(w, item)
}

func (h *Handler) UpdateTemplate(w http.ResponseWriter, r *http.Request) {
	var request templateUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(r.Context(), w, apperror.Invalid("invalid request body"))
		return
	}
	item, err := h.service.UpdateTemplate(r.Context(), r.PathValue("id"), request.command())
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) DeleteTemplate(w http.ResponseWriter, r *http.Request) {
	if err := h.service.DeleteTemplate(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteNoContent(w)
}

func (h *Handler) ListRecords(w http.ResponseWriter, r *http.Request) {
	options, err := query.ParseListOptions(r, map[string]bool{"status": true, "service_id": true, "template_id": true, "target_type": true}, map[string]bool{"status": true, "created_at": true, "updated_at": true}, "created_at")
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	items, total, err := h.service.ListRecords(r.Context(), domain.ListOptions{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Filters: options.Filters, Sort: options.Sort.Field, Order: options.Sort.Direction})
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.ListResponse{Data: items, Meta: httpx.Meta{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Total: total}})
}

func (h *Handler) GetRecord(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetRecord(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Distribute(w http.ResponseWriter, r *http.Request) {
	var request distributeRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(r.Context(), w, apperror.Invalid("invalid request body"))
		return
	}
	item, err := h.service.Distribute(r.Context(), request.command())
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteCreated(w, item)
}
