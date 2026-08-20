package domain

import (
	"context"
	"time"

	"github.com/acme/certpilot/internal/shared/database"
)

const (
	DistributionPending   = "pending"
	DistributionDelivered = "delivered"
	DistributionFailed    = "failed"
)

type ConfigTemplate struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	MinTLSVersion    string    `json:"min_tls_version"`
	CipherSuites     []string  `json:"cipher_suites"`
	RequireMutualTLS bool      `json:"require_mutual_tls"`
	RequireFullChain bool      `json:"require_full_chain"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Version          int       `json:"version"`
}

type DistributionRecord struct {
	ID            string     `json:"id"`
	ServiceID     string     `json:"service_id"`
	TemplateID    string     `json:"template_id"`
	CertificateID *string    `json:"certificate_id,omitempty"`
	TargetType    string     `json:"target_type"`
	Target        string     `json:"target"`
	Status        string     `json:"status"`
	PayloadJSON   string     `json:"payload_json"`
	Response      string     `json:"response"`
	DeliveredAt   *time.Time `json:"delivered_at,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
	Version       int        `json:"version"`
}

type Repository interface {
	CreateTemplate(ctx context.Context, exec database.Executor, template ConfigTemplate) error
	GetTemplate(ctx context.Context, exec database.Executor, id string) (ConfigTemplate, error)
	GetTemplateByName(ctx context.Context, exec database.Executor, name string) (ConfigTemplate, error)
	ListTemplates(ctx context.Context, exec database.Executor, options ListOptions) ([]ConfigTemplate, int, error)
	UpdateTemplate(ctx context.Context, exec database.Executor, template ConfigTemplate) error
	DeleteTemplate(ctx context.Context, exec database.Executor, id string) error
	CreateRecord(ctx context.Context, exec database.Executor, record DistributionRecord) error
	GetRecord(ctx context.Context, exec database.Executor, id string) (DistributionRecord, error)
	ListRecords(ctx context.Context, exec database.Executor, options ListOptions) ([]DistributionRecord, int, error)
	UpdateRecord(ctx context.Context, exec database.Executor, record DistributionRecord) error
}

type ListOptions struct {
	Page     int
	PageSize int
	Filters  map[string]string
	Sort     string
	Order    string
}
