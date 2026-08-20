package http

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/acme/certpilot/internal/issuer/application"
	"github.com/acme/certpilot/internal/issuer/domain"
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
	options, err := query.ParseListOptions(r, map[string]bool{"name": true, "provider": true, "enabled": true}, map[string]bool{"name": true, "provider": true, "created_at": true}, "created_at")
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

func (h *Handler) Create(w http.ResponseWriter, r *http.Request) {
	var request issuerCreateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		validationErr := apperror.Invalid("invalid request body")
		boundaryErr := fmt.Errorf("decode issuer create request: %v", validationErr)
		httpx.WriteError(r.Context(), w, boundaryErr)
		return
	}
	item, err := h.service.Create(r.Context(), request.command())
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteCreated(w, item)
}

type issuerCreateRequest struct {
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	ConfigJSON string `json:"config_json"`
	Enabled    *bool  `json:"enabled,omitempty"`
}

func (r issuerCreateRequest) command() application.CreateCommand {
	return application.CreateCommand{Name: r.Name, Provider: r.Provider, ConfigJSON: r.ConfigJSON, Enabled: r.Enabled}
}
