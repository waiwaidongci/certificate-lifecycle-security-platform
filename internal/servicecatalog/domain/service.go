package domain

import (
	"context"
	"time"

	"github.com/acme/certpilot/internal/shared/database"
)

type Service struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Environment   string    `json:"environment"`
	Domain        string    `json:"domain"`
	Owner         string    `json:"owner"`
	Region        string    `json:"region"`
	CertificateID *string   `json:"certificate_id,omitempty"`
	Enabled       bool      `json:"enabled"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
	Version       int       `json:"version"`
}

type Repository interface {
	Create(ctx context.Context, exec database.Executor, service Service) error
	Get(ctx context.Context, exec database.Executor, id string) (Service, error)
	GetByNameAndEnvironment(ctx context.Context, exec database.Executor, name, environment string) (Service, error)
	List(ctx context.Context, exec database.Executor, options ListOptions) ([]Service, int, error)
	Update(ctx context.Context, exec database.Executor, service Service) error
	Delete(ctx context.Context, exec database.Executor, id string) error
	BindCertificate(ctx context.Context, exec database.Executor, id, certificateID string) error
}

type ListOptions struct {
	Page     int
	PageSize int
	Filters  map[string]string
	Sort     string
	Order    string
}
