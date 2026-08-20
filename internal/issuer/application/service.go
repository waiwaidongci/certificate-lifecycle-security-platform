package application

import (
	"context"
	"strings"

	"github.com/acme/certpilot/internal/issuer/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/id"
)

type Service struct {
	db         *database.DB
	repository domain.Repository
	clock      clock.Clock
	issuerPort domain.IssuerPort
}

func NewService(db *database.DB, repository domain.Repository, clk clock.Clock, issuerPort domain.IssuerPort) *Service {
	return &Service{db: db, repository: repository, clock: clk, issuerPort: issuerPort}
}

type CreateCommand struct {
	Name       string `json:"name"`
	Provider   string `json:"provider"`
	ConfigJSON string `json:"config_json"`
	Enabled    *bool  `json:"enabled,omitempty"`
}

func (s *Service) Create(ctx context.Context, command CreateCommand) (domain.Issuer, error) {
	command.Name = strings.TrimSpace(command.Name)
	command.Provider = strings.TrimSpace(command.Provider)
	if command.Name == "" || command.Provider == "" {
		return domain.Issuer{}, apperror.Invalid("name and provider are required")
	}
	if command.ConfigJSON == "" {
		command.ConfigJSON = "{}"
	}
	enabled := true
	if command.Enabled != nil {
		enabled = *command.Enabled
	}
	now := s.clock.Now()
	issuer := domain.Issuer{
		ID:         id.New(),
		Name:       command.Name,
		Provider:   command.Provider,
		ConfigJSON: command.ConfigJSON,
		Enabled:    enabled,
		CreatedAt:  now,
		UpdatedAt:  now,
		Version:    1,
	}
	if err := s.db.WithTx(ctx, func(tx database.Tx) error {
		if _, getErr := s.repository.GetByName(ctx, tx, issuer.Name); getErr == nil {
			return apperror.Conflict("issuer already exists")
		}
		return s.repository.Create(ctx, tx, issuer)
	}); err != nil {
		return domain.Issuer{}, err
	}
	return issuer, nil
}

func (s *Service) Get(ctx context.Context, issuerID string) (domain.Issuer, error) {
	return s.repository.Get(ctx, s.db, issuerID)
}

func (s *Service) List(ctx context.Context, options domain.ListOptions) ([]domain.Issuer, int, error) {
	return s.repository.List(ctx, s.db, options)
}

func (s *Service) Issue(ctx context.Context, request domain.IssueRequest) (domain.IssueResult, error) {
	request = request.Clone()
	return s.issuerPort.Issue(ctx, request)
}
