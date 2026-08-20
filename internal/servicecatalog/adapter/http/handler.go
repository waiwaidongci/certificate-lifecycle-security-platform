package http

import (
	"encoding/json"
	"net/http"

	"github.com/acme/certpilot/internal/servicecatalog/application"
	"github.com/acme/certpilot/internal/servicecatalog/domain"
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
	options, err := query.ParseListOptions(r, map[string]bool{"environment": true, "owner": true, "region": true, "name": true, "domain": true, "enabled": true}, map[string]bool{"name": true, "environment": true, "domain": true, "owner": true, "region": true, "created_at": true, "updated_at": true}, "created_at")
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	items, total, err := h.service.List(r.Context(), domain.ListOptions{
		Page:     options.Pagination.Page,
		PageSize: options.Pagination.PageSize,
		Filters:  options.Filters,
		Sort:     options.Sort.Field,
		Order:    options.Sort.Direction,
	})
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.ListResponse{Data: items, Meta: httpx.Meta{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Total: total}})
}

func (h *Handler) Get(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	item, err := h.service.Get(r.Context(), id)
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request serviceCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(r.Context(), w, apperror.Invalid("invalid request body"))
		return
	}
	item, err := h.service.Create(r.Context(), request.command())
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteCreated(w, item)
}

func (h *Handler) Update(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	var request serviceUpdateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(r.Context(), w, apperror.Invalid("invalid request body"))
		return
	}
	item, err := h.service.Update(r.Context(), id, request.command())
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) Delete(w http.ResponseWriter, r *http.Request) {
	if err := h.service.Delete(r.Context(), r.PathValue("id")); err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteNoContent(w)
}
