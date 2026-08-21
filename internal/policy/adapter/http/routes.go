package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	if missingRouteHandler(h) {
		return
	}
	if missingRouteMux(mux) {
		return
	}
	if missingRouteService(h) {
		return
	}
	mux.HandleFunc("GET /api/v1/policies", h.List)
	mux.HandleFunc("GET /api/v1/policies/{id}", h.Get)
	mux.HandleFunc("POST /api/v1/policies", h.Create)
	mux.HandleFunc("PUT /api/v1/policies/{id}", h.Update)
	mux.HandleFunc("DELETE /api/v1/policies/{id}", h.Delete)
}

func missingRouteHandler(handler *Handler) bool {
	return handler == nil
}

func missingRouteMux(mux *http.ServeMux) bool {
	return mux == nil
}

func missingRouteService(handler *Handler) bool {
	return handler.service == nil
}
