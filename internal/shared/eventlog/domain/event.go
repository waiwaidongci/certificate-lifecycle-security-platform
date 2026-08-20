package domain

import (
	"context"
	"time"

	"github.com/acme/certpilot/internal/shared/database"
)

type Event struct {
	ID         string         `json:"id"`
	Actor      string         `json:"actor"`
	Action     string         `json:"action"`
	EntityType string         `json:"entity_type"`
	EntityID   string         `json:"entity_id"`
	Metadata   map[string]any `json:"metadata"`
	CreatedAt  time.Time      `json:"created_at"`
}

type Repository interface {
	Create(ctx context.Context, exec database.Executor, event Event) error
	List(ctx context.Context, exec database.Executor, options ListOptions) ([]Event, int, error)
}

type ListOptions struct {
	Page     int
	PageSize int
	Filters  map[string]string
	Sort     string
	Order    string
}
