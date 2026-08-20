package http

import (
	"encoding/json"
	"net/http"

	"github.com/acme/certpilot/internal/rotation/application"
	"github.com/acme/certpilot/internal/rotation/domain"
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

func (h *Handler) Generate(w http.ResponseWriter, r *http.Request) {
	var request generateRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(r.Context(), w, apperror.Invalid("invalid request body"))
		return
	}
	plans, err := h.service.Generate(r.Context(), request.AdvanceDays)
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteCreated(w, plans)
}

func (h *Handler) ListPlans(w http.ResponseWriter, r *http.Request) {
	options, err := query.ParseListOptions(r, map[string]bool{"status": true, "certificate_id": true, "service_id": true}, map[string]bool{"due_at": true, "priority": true, "status": true, "created_at": true}, "created_at")
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	items, total, err := h.service.ListPlans(r.Context(), domain.ListOptions{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Filters: options.Filters, Sort: options.Sort.Field, Order: options.Sort.Direction})
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.ListResponse{Data: items, Meta: httpx.Meta{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Total: total}})
}

func (h *Handler) GetPlan(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetPlan(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) ListTasks(w http.ResponseWriter, r *http.Request) {
	options, err := query.ParseListOptions(r, map[string]bool{"status": true, "plan_id": true}, map[string]bool{"status": true, "attempts": true, "created_at": true}, "created_at")
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	items, total, err := h.service.ListTasks(r.Context(), domain.ListOptions{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Filters: options.Filters, Sort: options.Sort.Field, Order: options.Sort.Direction})
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, httpx.ListResponse{Data: items, Meta: httpx.Meta{Page: options.Pagination.Page, PageSize: options.Pagination.PageSize, Total: total}})
}

func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	item, err := h.service.GetTask(r.Context(), r.PathValue("id"))
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}

func (h *Handler) TransitionTask(w http.ResponseWriter, r *http.Request) {
	var request transitionRequest
	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		httpx.WriteError(r.Context(), w, apperror.Invalid("invalid request body"))
		return
	}
	item, err := h.service.TransitionTask(r.Context(), r.PathValue("id"), request.command())
	if err != nil {
		httpx.WriteError(r.Context(), w, err)
		return
	}
	httpx.WriteJSON(w, http.StatusOK, item)
}
