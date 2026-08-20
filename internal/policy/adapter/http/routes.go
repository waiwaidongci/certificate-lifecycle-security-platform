package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/policies", h.List)
	mux.HandleFunc("GET /api/v1/policies/{id}", h.Get)
	mux.HandleFunc("POST /api/v1/policies", h.Create)
	mux.HandleFunc("PUT /api/v1/policies/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/policies/{id}", h.Delete)
}
