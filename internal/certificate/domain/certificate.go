package domain

import (
	"context"
	"time"

	"github.com/acme/certpilot/internal/shared/database"
)

const (
	StatusIssued  = "issued"
	StatusRevoked = "revoked"
	StatusExpired = "expired"
)

type Certificate struct {
	ID             string    `json:"id"`
	SerialNumber   string    `json:"serial_number"`
	CommonName     string    `json:"common_name"`
	SANs           []string  `json:"sans"`
	IssuerID       string    `json:"issuer_id"`
	Status         string    `json:"status"`
	NotBefore      time.Time `json:"not_before"`
	NotAfter       time.Time `json:"not_after"`
	Fingerprint    string    `json:"fingerprint"`
	Source         string    `json:"source"`
	RequestID      string    `json:"request_id"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
	ServiceID      *string   `json:"service_id,omitempty"`
	CreatedAt      time.Time `json:"created_at"`
	UpdatedAt      time.Time `json:"updated_at"`
	Version        int       `json:"version"`
}

type Repository interface {
	Create(ctx context.Context, exec database.Executor, certificate Certificate) error
	Get(ctx context.Context, exec database.Executor, id string) (Certificate, error)
	GetByIdempotencyKey(ctx context.Context, exec database.Executor, key string) (Certificate, error)
	List(ctx context.Context, exec database.Executor, options ListOptions) ([]Certificate, int, error)
	UpdateStatus(ctx context.Context, exec database.Executor, certificate Certificate) error
	ListExpiring(ctx context.Context, exec database.Executor, before time.Time, status string) ([]Certificate, error)
}

type ListOptions struct {
	Page     int
	PageSize int
	Filters  map[string]string
	Sort     string
	Order    string
}
