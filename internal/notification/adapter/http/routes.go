package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/notifications/reminders/scan", h.Scan)
	mux.HandleFunc("POST /api/v1/notifications/reminders/send", h.SendPending)
	mux.HandleFunc("GET /api/v1/notifications/reminders", h.List)
	mux.HandleFunc("GET /api/v1/notifications/reminders/{id}", h.Get)
}
