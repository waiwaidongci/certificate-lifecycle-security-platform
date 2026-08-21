package http

import (
	"github.com/acme/certpilot/internal/policy/application"
	"github.com/acme/certpilot/internal/policy/domain"
)

type policyCreateRequest struct {
	Name            string   `json:"name"`
	Environments    []string `json:"environments"`
	MinValidityDays int      `json:"min_validity_days"`
	MaxValidityDays int      `json:"max_validity_days"`
	AllowedDomains  []string `json:"allowed_domains"`
	Enabled         *bool    `json:"enabled,omitempty"`
}

func (r policyCreateRequest) command() application.CreateCommand {
	return application.CreateCommand{
		Name:            r.Name,
		Environments:    domain.CopyStrings(r.Environments),
		MinValidityDays: r.MinValidityDays,
		MaxValidityDays: r.MaxValidityDays,
		AllowedDomains:  domain.CopyStrings(r.AllowedDomains),
		Enabled:         r.Enabled,
	}
}

type policyUpdateRequest struct {
	Name            *string   `json:"name,omitempty"`
	Environments    *[]string `json:"environments,omitempty"`
	MinValidityDays *int      `json:"min_validity_days,omitempty"`
	MaxValidityDays *int      `json:"max_validity_days,omitempty"`
	AllowedDomains  *[]string `json:"allowed_domains,omitempty"`
	Enabled         *bool     `json:"enabled,omitempty"`
	Version         int       `json:"version"`
}

func (r policyUpdateRequest) command() application.UpdateCommand {
	return application.UpdateCommand{
		Name:            r.Name,
		Environments:    r.Environments,
		MinValidityDays: r.MinValidityDays,
		MaxValidityDays: r.MaxValidityDays,
		AllowedDomains:  r.AllowedDomains,
		Enabled:         r.Enabled,
		Version:         r.Version,
	}
}
