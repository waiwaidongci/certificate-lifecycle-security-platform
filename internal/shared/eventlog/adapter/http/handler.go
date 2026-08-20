package http

import (
	"net/http"

	"github.com/acme/certpilot/internal/shared/eventlog/application"
	"github.com/acme/certpilot/internal/shared/eventlog/domain"
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
	options, err := query.ParseListOptions(r, map[string]bool{"actor": true, "action": true, "entity_type": true, "entity_id": true}, map[string]bool{"created_at": true, "actor": true, "action": true}, "created_at")
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
