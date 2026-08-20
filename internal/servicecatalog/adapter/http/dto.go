package http

import (
	"github.com/acme/certpilot/internal/servicecatalog/application"
)

type serviceCreateRequest struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Domain      string `json:"domain"`
	Owner       string `json:"owner"`
	Region      string `json:"region"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

func (r serviceCreateRequest) command() application.CreateCommand {
	return application.CreateCommand{
		Name:        r.Name,
		Environment: r.Environment,
		Domain:      r.Domain,
		Owner:       r.Owner,
		Region:      r.Region,
		Enabled:     r.Enabled,
	}
}

type serviceUpdateRequest struct {
	Name        *string `json:"name,omitempty"`
	Environment *string `json:"environment,omitempty"`
	Domain      *string `json:"domain,omitempty"`
	Owner       *string `json:"owner,omitempty"`
	Region      *string `json:"region,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
	Version     int     `json:"version"`
}

func (r serviceUpdateRequest) command() application.UpdateCommand {
	return application.UpdateCommand{
		Name:        r.Name,
		Environment: r.Environment,
		Domain:      r.Domain,
		Owner:       r.Owner,
		Region:      r.Region,
		Enabled:     r.Enabled,
		Version:     r.Version,
	}
}
