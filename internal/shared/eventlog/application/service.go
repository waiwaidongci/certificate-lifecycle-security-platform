package application

import (
	"context"
	"time"

	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/eventlog/domain"
	"github.com/acme/certpilot/internal/shared/id"
)

type Service struct {
	db         *database.DB
	repository domain.Repository
	clock      clock.Clock
}

func NewService(db *database.DB, repository domain.Repository, clk clock.Clock) *Service {
	return &Service{db: db, repository: repository, clock: clk}
}

func (s *Service) Record(ctx context.Context, exec database.Executor, actor, action, entityType, entityID string, metadata map[string]any) error {
	now := s.clock.Now()
	event := domain.Event{
		ID:         id.New(),
		Actor:      actor,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Metadata:   metadata,
		CreatedAt:  now,
	}
	return s.repository.Create(ctx, exec, event)
}

func (s *Service) List(ctx context.Context, options domain.ListOptions) ([]domain.Event, int, error) {
	return s.repository.List(ctx, s.db, options)
}

func (s *Service) Now() time.Time {
	return s.clock.Now()
}
