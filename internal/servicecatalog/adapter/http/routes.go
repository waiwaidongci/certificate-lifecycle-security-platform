package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/services", h.List)
	mux.HandleFunc("GET /api/v1/services/{id}", h.Get)
	mux.HandleFunc("POST /api/v1/services", h.Create)
	mux.HandleFunc("PUT /api/v1/services/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/services/{id}", h.Delete)
}
