package http

import "github.com/acme/certpilot/internal/distribution/application"

type templateCreateRequest struct {
	Name             string   `json:"name"`
	MinTLSVersion    string   `json:"min_tls_version"`
	CipherSuites     []string `json:"cipher_suites"`
	RequireMutualTLS bool     `json:"require_mutual_tls"`
	RequireFullChain bool     `json:"require_full_chain"`
}

func (r templateCreateRequest) command() application.CreateTemplateCommand {
	return application.CreateTemplateCommand{Name: r.Name, MinTLSVersion: r.MinTLSVersion, CipherSuites: r.CipherSuites, RequireMutualTLS: r.RequireMutualTLS, RequireFullChain: r.RequireFullChain}
}

type templateUpdateRequest struct {
	Name             *string   `json:"name,omitempty"`
	MinTLSVersion    *string   `json:"min_tls_version,omitempty"`
	CipherSuites     *[]string `json:"cipher_suites,omitempty"`
	RequireMutualTLS *bool     `json:"require_mutual_tls,omitempty"`
	RequireFullChain *bool     `json:"require_full_chain,omitempty"`
	Version          int       `json:"version"`
}

func (r templateUpdateRequest) command() application.UpdateTemplateCommand {
	return application.UpdateTemplateCommand{Name: r.Name, MinTLSVersion: r.MinTLSVersion, CipherSuites: r.CipherSuites, RequireMutualTLS: r.RequireMutualTLS, RequireFullChain: r.RequireFullChain, Version: r.Version}
}

type distributeRequest struct {
	ServiceID     string `json:"service_id"`
	TemplateID    string `json:"template_id"`
	CertificateID string `json:"certificate_id,omitempty"`
	TargetType    string `json:"target_type"`
	Target        string `json:"target"`
}

func (r distributeRequest) command() application.DistributeCommand {
	return application.DistributeCommand{ServiceID: r.ServiceID, TemplateID: r.TemplateID, CertificateID: r.CertificateID, TargetType: r.TargetType, Target: r.Target}
}
