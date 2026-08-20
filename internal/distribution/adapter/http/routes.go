package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /api/v1/config-templates", h.ListTemplates)
	mux.HandleFunc("GET /api/v1/config-templates/{id}", h.GetTemplate)
	mux.HandleFunc("POST /api/v1/config-templates", h.CreateTemplate)
	mux.HandleFunc("PUT /api/v1/config-templates/{id}", h.UpdateTemplate)
	mux.HandleFunc("DELETE /api/v1/config-templates/{id}", h.DeleteTemplate)
	mux.HandleFunc("GET /api/v1/distributions", h.ListRecords)
	mux.HandleFunc("GET /api/v1/distributions/{id}", h.GetRecord)
	mux.HandleFunc("POST /api/v1/distributions", h.Distribute)
}
