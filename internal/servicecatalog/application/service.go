package application

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/acme/certpilot/internal/servicecatalog/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/database"
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

type CreateCommand struct {
	Name        string `json:"name"`
	Environment string `json:"environment"`
	Domain      string `json:"domain"`
	Owner       string `json:"owner"`
	Region      string `json:"region"`
	Enabled     *bool  `json:"enabled,omitempty"`
}

func (s *Service) Create(ctx context.Context, command CreateCommand) (domain.Service, error) {
	command.Name = strings.TrimSpace(command.Name)
	command.Environment = strings.TrimSpace(command.Environment)
	command.Domain = strings.TrimSpace(command.Domain)
	command.Owner = strings.TrimSpace(command.Owner)
	command.Region = strings.TrimSpace(command.Region)
	if command.Name == "" || command.Environment == "" || command.Domain == "" || command.Owner == "" {
		return domain.Service{}, apperror.Invalid("name, environment, domain and owner are required")
	}
	if !strings.HasPrefix(command.Domain, ".") && !strings.Contains(command.Domain, ".") {
		return domain.Service{}, apperror.Invalid("domain must be a valid DNS name")
	}
	now := s.clock.Now()
	enabled := true
	if command.Enabled != nil {
		enabled = *command.Enabled
	}
	service := domain.Service{
		ID:          id.New(),
		Name:        command.Name,
		Environment: command.Environment,
		Domain:      command.Domain,
		Owner:       command.Owner,
		Region:      command.Region,
		Enabled:     enabled,
		CreatedAt:   now,
		UpdatedAt:   now,
		Version:     1,
	}
	if err := s.db.WithTx(ctx, func(tx database.Tx) error {
		if _, err := s.repository.GetByNameAndEnvironment(ctx, tx, service.Name, service.Environment); err == nil {
			return apperror.Conflict("service already exists for this environment")
		} else if !errors.As(err, new(*apperror.Error)) {
			return err
		}
		return s.repository.Create(ctx, tx, service)
	}); err != nil {
		return domain.Service{}, err
	}
	return service, nil
}

func (s *Service) Get(ctx context.Context, serviceID string) (domain.Service, error) {
	return s.repository.Get(ctx, s.db, serviceID)
}

func (s *Service) List(ctx context.Context, options domain.ListOptions) ([]domain.Service, int, error) {
	return s.repository.List(ctx, s.db, options)
}

type UpdateCommand struct {
	Name        *string `json:"name,omitempty"`
	Environment *string `json:"environment,omitempty"`
	Domain      *string `json:"domain,omitempty"`
	Owner       *string `json:"owner,omitempty"`
	Region      *string `json:"region,omitempty"`
	Enabled     *bool   `json:"enabled,omitempty"`
	Version     int     `json:"version"`
}

func (s *Service) Update(ctx context.Context, serviceID string, command UpdateCommand) (domain.Service, error) {
	current, err := s.repository.Get(ctx, s.db, serviceID)
	if err != nil {
		return domain.Service{}, err
	}
	if command.Version != 0 && command.Version != current.Version {
		return domain.Service{}, apperror.Conflict("optimistic lock conflict")
	}
	if command.Name != nil {
		current.Name = strings.TrimSpace(*command.Name)
	}
	if command.Environment != nil {
		current.Environment = strings.TrimSpace(*command.Environment)
	}
	if command.Domain != nil {
		current.Domain = strings.TrimSpace(*command.Domain)
	}
	if command.Owner != nil {
		current.Owner = strings.TrimSpace(*command.Owner)
	}
	if command.Region != nil {
		current.Region = strings.TrimSpace(*command.Region)
	}
	if command.Enabled != nil {
		current.Enabled = *command.Enabled
	}
	current.UpdatedAt = s.clock.Now()
	current.Version++
	if err := s.repository.Update(ctx, s.db, current); err != nil {
		return domain.Service{}, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, serviceID string) error {
	if err := s.repository.Delete(ctx, s.db, serviceID); err != nil {
		return err
	}
	return nil
}

func (s *Service) BindCertificate(ctx context.Context, serviceID, certificateID string) error {
	if err := s.db.WithTx(ctx, func(tx database.Tx) error {
		return s.BindCertificateTx(ctx, tx, serviceID, certificateID)
	}); err != nil {
		return fmt.Errorf("bind certificate: %w", err)
	}
	return nil
}

func (s *Service) BindCertificateTx(ctx context.Context, exec database.Executor, serviceID, certificateID string) error {
	service, err := s.repository.Get(ctx, exec, serviceID)
	if err != nil {
		return err
	}
	if service.CertificateID != nil && *service.CertificateID == certificateID {
		return nil
	}
	return s.repository.BindCertificate(ctx, exec, serviceID, certificateID)
}
