package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/issuers", h.List)
	mux.HandleFunc("GET /api/v1/issuers/{id}", h.Get)
	mux.HandleFunc("POST /api/v1/issuers", h.Create)
}
