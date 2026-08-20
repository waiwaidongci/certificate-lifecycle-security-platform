package domain

import (
	"context"
	"time"

	"github.com/acme/certpilot/internal/shared/database"
)

type Issuer struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	Provider   string    `json:"provider"`
	ConfigJSON string    `json:"config_json"`
	Enabled    bool      `json:"enabled"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Version    int       `json:"version"`
}

type Repository interface {
	Create(ctx context.Context, exec database.Executor, issuer Issuer) error
	Get(ctx context.Context, exec database.Executor, id string) (Issuer, error)
	GetByName(ctx context.Context, exec database.Executor, name string) (Issuer, error)
	List(ctx context.Context, exec database.Executor, options ListOptions) ([]Issuer, int, error)
}

type ListOptions struct {
	Page     int
	PageSize int
	Filters  map[string]string
	Sort     string
	Order    string
}

type IssueRequest struct {
	CommonName   string
	SANs         []string
	ValidityDays int
	ServiceID    string
	Environment  string
}

func (r IssueRequest) Clone() IssueRequest {
	clone := IssueRequest{
		CommonName:   r.CommonName,
		ValidityDays: r.ValidityDays,
		ServiceID:    r.ServiceID,
		Environment:  r.Environment,
	}
	if len(r.SANs) == 0 {
		return clone
	}
	clone.SANs = make([]string, len(r.SANs))
	copy(clone.SANs, r.SANs)
	return clone
}

type IssueResult struct {
	SerialNumber   string    `json:"serial_number"`
	Fingerprint    string    `json:"fingerprint"`
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
	CertificatePEM string    `json:"certificate_pem"`
	IssuerID       string    `json:"issuer_id"`
}

type IssuerPort interface {
	Issue(ctx context.Context, request IssueRequest) (IssueResult, error)
}
