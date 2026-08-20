package http

import "github.com/acme/certpilot/internal/certificate/application"

type issueRequest struct {
	ServiceID      string   `json:"service_id"`
	CommonName     string   `json:"common_name"`
	SANs           []string `json:"sans,omitempty"`
	ValidityDays   int      `json:"validity_days,omitempty"`
	IdempotencyKey string   `json:"idempotency_key,omitempty"`
}

func (r issueRequest) command() application.IssueCommand {
	sans := append([]string(nil), r.SANs...)
	return application.IssueCommand{ServiceID: r.ServiceID, CommonName: r.CommonName, SANs: sans, ValidityDays: r.ValidityDays, IdempotencyKey: r.IdempotencyKey}
}

type revokeRequest struct {
	Reason  string `json:"reason"`
	Version int    `json:"version"`
}

func (r revokeRequest) command() application.RevokeCommand {
	return application.RevokeCommand{Reason: r.Reason, Version: r.Version}
}
