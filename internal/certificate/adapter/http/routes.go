package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/certificates", h.List)
	mux.HandleFunc("GET /api/v1/certificates/{id}", h.Get)
	mux.HandleFunc("POST /api/v1/certificates/issue", h.Issue)
	mux.HandleFunc("POST /api/v1/certificates/{id}/revoke", h.Revoke)
}
