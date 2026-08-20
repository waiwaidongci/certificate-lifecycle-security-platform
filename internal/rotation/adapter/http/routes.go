package http

import "net/http"

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("POST /api/v1/rotation/plans/generate", h.Generate)
	mux.HandleFunc("GET /api/v1/rotation/plans", h.ListPlans)
	mux.HandleFunc("GET /api/v1/rotation/plans/{id}", h.GetPlan)
	mux.HandleFunc("GET /api/v1/rotation/tasks", h.ListTasks)
	mux.HandleFunc("GET /api/v1/rotation/tasks/{id}", h.GetTask)
	mux.HandleFunc("POST /api/v1/rotation/tasks/{id}/transition", h.TransitionTask)
}
