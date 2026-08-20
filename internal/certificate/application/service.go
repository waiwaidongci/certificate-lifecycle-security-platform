package application

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/certificate/domain"
	issuerapp "github.com/acme/certpilot/internal/issuer/application"
	issuerdomain "github.com/acme/certpilot/internal/issuer/domain"
	policyapp "github.com/acme/certpilot/internal/policy/application"
	serviceapp "github.com/acme/certpilot/internal/servicecatalog/application"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/eventlog/application"
	"github.com/acme/certpilot/internal/shared/httpx"
	"github.com/acme/certpilot/internal/shared/id"
)

type Service struct {
	db                  *database.DB
	repository          domain.Repository
	services            *serviceapp.Service
	policies            *policyapp.Service
	issuers             *issuerapp.Service
	events              *application.Service
	clock               clock.Clock
	defaultValidityDays int
}

func NewService(db *database.DB, repository domain.Repository, services *serviceapp.Service, policies *policyapp.Service, issuers *issuerapp.Service, events *application.Service, clk clock.Clock, defaultValidityDays int) *Service {
	return &Service{
		db: db, repository: repository, services: services, policies: policies, issuers: issuers, events: events, clock: clk, defaultValidityDays: defaultValidityDays,
	}
}

type IssueCommand struct {
	ServiceID      string   `json:"service_id"`
	CommonName     string   `json:"common_name"`
	SANs           []string `json:"sans,omitempty"`
	ValidityDays   int      `json:"validity_days,omitempty"`
	IdempotencyKey string   `json:"idempotency_key,omitempty"`
}

func (s *Service) Issue(ctx context.Context, command IssueCommand) (domain.Certificate, error) {
	if ctx == nil {
		ctx = context.Background()
	} else {
		ctx = context.WithoutCancel(ctx)
	}
	command.ServiceID = strings.TrimSpace(command.ServiceID)
	command.CommonName = strings.TrimSpace(command.CommonName)
	if command.ServiceID == "" || command.CommonName == "" {
		return domain.Certificate{}, apperror.Invalid("service_id and common_name are required")
	}
	if command.ValidityDays == 0 {
		command.ValidityDays = s.defaultValidityDays
	}
	if command.IdempotencyKey != "" {
		if existing, err := s.repository.GetByIdempotencyKey(ctx, s.db, command.IdempotencyKey); err == nil {
			return existing, nil
		} else if !isNotFound(err) {
			return domain.Certificate{}, err
		}
	}

	service, err := s.services.Get(ctx, command.ServiceID)
	if err != nil {
		return domain.Certificate{}, err
	}
	if service.CertificateID != nil {
		return domain.Certificate{}, apperror.Conflict("service already has a bound certificate")
	}
	if !strings.HasSuffix(command.CommonName, service.Domain) && command.CommonName != service.Domain {
		return domain.Certificate{}, apperror.Invalid("common_name must belong to service domain")
	}

	requestID := id.New()
	var result issuerdomain.IssueResult
	var certificate domain.Certificate
	err = s.db.WithTx(ctx, func(tx database.Tx) error {
		if err := s.policies.Evaluate(ctx, tx, policyapp.EvaluationRequest{
			Environment:  service.Environment,
			Domain:       service.Domain,
			SANs:         command.SANs,
			ValidityDays: command.ValidityDays,
		}); err != nil {
			return err
		}
		result, err = s.issuers.Issue(ctx, issuerdomain.IssueRequest{
			CommonName:   command.CommonName,
			SANs:         command.SANs,
			ValidityDays: command.ValidityDays,
			ServiceID:    service.ID,
			Environment:  service.Environment,
		})
		if err != nil {
			return err
		}
		now := s.clock.Now()
		certificate = domain.Certificate{
			ID:             id.New(),
			SerialNumber:   result.SerialNumber,
			CommonName:     command.CommonName,
			SANs:           command.SANs,
			IssuerID:       result.IssuerID,
			Status:         domain.StatusIssued,
			NotBefore:      result.NotBefore,
			NotAfter:       result.NotAfter,
			Fingerprint:    result.Fingerprint,
			Source:         "mock",
			RequestID:      requestID,
			IdempotencyKey: command.IdempotencyKey,
			ServiceID:      &service.ID,
			CreatedAt:      now,
			UpdatedAt:      now,
			Version:        1,
		}
		if err := s.repository.Create(ctx, tx, certificate); err != nil {
			return err
		}
		if err := s.services.BindCertificateTx(ctx, tx, service.ID, certificate.ID); err != nil {
			return err
		}
		return s.events.Record(ctx, tx, httpx.ActorFromContext(ctx), "certificate.issued", "certificate", certificate.ID, map[string]any{"service_id": service.ID})
	})
	if err != nil {
		return domain.Certificate{}, err
	}
	return certificate, nil
}

func (s *Service) Get(ctx context.Context, certificateID string) (domain.Certificate, error) {
	return s.repository.Get(ctx, s.db, certificateID)
}

func (s *Service) List(ctx context.Context, options domain.ListOptions) ([]domain.Certificate, int, error) {
	return s.repository.List(ctx, s.db, options)
}

type RevokeCommand struct {
	Reason  string `json:"reason"`
	Version int    `json:"version"`
}

func (s *Service) Revoke(ctx context.Context, certificateID string, command RevokeCommand) (domain.Certificate, error) {
	current, err := s.repository.Get(ctx, s.db, certificateID)
	if err != nil {
		return domain.Certificate{}, err
	}
	if command.Version != 0 && command.Version != current.Version {
		return domain.Certificate{}, apperror.Conflict("optimistic lock conflict")
	}
	if current.Status == domain.StatusRevoked {
		return current, nil
	}
	current.Status = domain.StatusRevoked
	current.UpdatedAt = s.clock.Now()
	current.Version++
	if err := s.repository.UpdateStatus(ctx, s.db, current); err != nil {
		return domain.Certificate{}, err
	}
	if err := s.events.Record(ctx, s.db, httpx.ActorFromContext(ctx), "certificate.revoked", "certificate", current.ID, map[string]any{"reason": command.Reason}); err != nil {
		return domain.Certificate{}, err
	}
	return current, nil
}

func (s *Service) ListExpiring(ctx context.Context, before time.Time) ([]domain.Certificate, error) {
	return s.repository.ListExpiring(ctx, s.db, before, domain.StatusIssued)
}

func isNotFound(err error) bool {
	var apiErr *apperror.Error
	return errors.As(err, &apiErr) && apiErr.Code == apperror.CodeNotFound
}
