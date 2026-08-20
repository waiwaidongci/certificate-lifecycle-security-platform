package domain

import (
	"context"
	"time"

	"github.com/acme/certpilot/internal/shared/database"
)

type Policy struct {
	ID              string    `json:"id"`
	Name            string    `json:"name"`
	Environments    []string  `json:"environments"`
	MinValidityDays int       `json:"min_validity_days"`
	MaxValidityDays int       `json:"max_validity_days"`
	AllowedDomains  []string  `json:"allowed_domains"`
	Enabled         bool      `json:"enabled"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Version         int       `json:"version"`
}

type Repository interface {
	Create(ctx context.Context, exec database.Executor, policy Policy) error
	Get(ctx context.Context, exec database.Executor, id string) (Policy, error)
	GetByName(ctx context.Context, exec database.Executor, name string) (Policy, error)
	List(ctx context.Context, exec database.Executor, options ListOptions) ([]Policy, int, error)
	Update(ctx context.Context, exec database.Executor, policy Policy) error
	Delete(ctx context.Context, exec database.Executor, id string) error
	FindForEvaluation(ctx context.Context, exec database.Executor, environment, domain string) (Policy, error)
}

type ListOptions struct {
	Page     int
	PageSize int
	Filters  map[string]string
	Sort     string
	Order    string
}
