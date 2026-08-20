package http

import "github.com/acme/certpilot/internal/rotation/application"

type generateRequest struct {
	AdvanceDays int `json:"advance_days"`
}

type transitionRequest struct {
	TargetStatus string `json:"target_status"`
	Error        string `json:"error,omitempty"`
	AssignedTo   string `json:"assigned_to,omitempty"`
	Version      int    `json:"version"`
}

func (r transitionRequest) command() application.TransitionCommand {
	return application.TransitionCommand{TargetStatus: r.TargetStatus, Error: r.Error, AssignedTo: r.AssignedTo, Version: r.Version}
}
