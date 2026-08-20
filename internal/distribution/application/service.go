package application

import (
	"context"
	"encoding/json"
	"strings"
	"time"

	certapp "github.com/acme/certpilot/internal/certificate/application"
	"github.com/acme/certpilot/internal/distribution/domain"
	serviceapp "github.com/acme/certpilot/internal/servicecatalog/application"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/eventlog/application"
	"github.com/acme/certpilot/internal/shared/httpx"
	"github.com/acme/certpilot/internal/shared/id"
)

type DistributionAdapter interface {
	Send(ctx context.Context, record domain.DistributionRecord) (string, error)
}

type Service struct {
	db           *database.DB
	repository   domain.Repository
	services     *serviceapp.Service
	certificates *certapp.Service
	events       *application.Service
	clock        clock.Clock
	adapters     map[string]DistributionAdapter
	dispatches   *dispatchGate
}

func NewService(db *database.DB, repository domain.Repository, services *serviceapp.Service, certificates *certapp.Service, events *application.Service, clk clock.Clock, adapters map[string]DistributionAdapter) *Service {
	return &Service{db: db, repository: repository, services: services, certificates: certificates, events: events, clock: clk, adapters: adapters, dispatches: newDispatchGate()}
}

type CreateTemplateCommand struct {
	Name             string   `json:"name"`
	MinTLSVersion    string   `json:"min_tls_version"`
	CipherSuites     []string `json:"cipher_suites"`
	RequireMutualTLS bool     `json:"require_mutual_tls"`
	RequireFullChain bool     `json:"require_full_chain"`
}

func (s *Service) CreateTemplate(ctx context.Context, command CreateTemplateCommand) (domain.ConfigTemplate, error) {
	template, err := newTemplate(s.clock.Now(), command)
	if err != nil {
		return domain.ConfigTemplate{}, err
	}
	if err := s.db.WithTx(ctx, func(tx database.Tx) error {
		if _, getErr := s.repository.GetTemplateByName(ctx, tx, template.Name); getErr == nil {
			return apperror.Conflict("config template already exists")
		}
		return s.repository.CreateTemplate(ctx, tx, template)
	}); err != nil {
		return domain.ConfigTemplate{}, err
	}
	return template, nil
}

func (s *Service) GetTemplate(ctx context.Context, templateID string) (domain.ConfigTemplate, error) {
	return s.repository.GetTemplate(ctx, s.db, templateID)
}

func (s *Service) ListTemplates(ctx context.Context, options domain.ListOptions) ([]domain.ConfigTemplate, int, error) {
	return s.repository.ListTemplates(ctx, s.db, options)
}

type UpdateTemplateCommand struct {
	Name             *string   `json:"name,omitempty"`
	MinTLSVersion    *string   `json:"min_tls_version,omitempty"`
	CipherSuites     *[]string `json:"cipher_suites,omitempty"`
	RequireMutualTLS *bool     `json:"require_mutual_tls,omitempty"`
	RequireFullChain *bool     `json:"require_full_chain,omitempty"`
	Version          int       `json:"version"`
}

func (s *Service) UpdateTemplate(ctx context.Context, templateID string, command UpdateTemplateCommand) (domain.ConfigTemplate, error) {
	current, err := s.repository.GetTemplate(ctx, s.db, templateID)
	if err != nil {
		return domain.ConfigTemplate{}, err
	}
	if command.Version != 0 && command.Version != current.Version {
		return domain.ConfigTemplate{}, apperror.Conflict("optimistic lock conflict")
	}
	if command.Name != nil {
		current.Name = strings.TrimSpace(*command.Name)
	}
	if command.MinTLSVersion != nil {
		current.MinTLSVersion = strings.TrimSpace(*command.MinTLSVersion)
	}
	if command.CipherSuites != nil {
		current.CipherSuites = cleanStrings(*command.CipherSuites)
	}
	if command.RequireMutualTLS != nil {
		current.RequireMutualTLS = *command.RequireMutualTLS
	}
	if command.RequireFullChain != nil {
		current.RequireFullChain = *command.RequireFullChain
	}
	if err := validateTemplate(current); err != nil {
		return domain.ConfigTemplate{}, err
	}
	current.UpdatedAt = s.clock.Now()
	current.Version++
	if err := s.repository.UpdateTemplate(ctx, s.db, current); err != nil {
		return domain.ConfigTemplate{}, err
	}
	return current, nil
}

func (s *Service) DeleteTemplate(ctx context.Context, templateID string) error {
	return s.repository.DeleteTemplate(ctx, s.db, templateID)
}

type DistributeCommand struct {
	ServiceID     string `json:"service_id"`
	TemplateID    string `json:"template_id"`
	CertificateID string `json:"certificate_id,omitempty"`
	TargetType    string `json:"target_type"`
	Target        string `json:"target"`
}

func (s *Service) Distribute(ctx context.Context, command DistributeCommand) (domain.DistributionRecord, error) {
	command.TargetType = strings.ToLower(strings.TrimSpace(command.TargetType))
	if command.ServiceID == "" || command.TemplateID == "" || command.TargetType == "" {
		return domain.DistributionRecord{}, apperror.Invalid("service_id, template_id and target_type are required")
	}
	key := domain.NewDispatchKey(command.ServiceID, command.TemplateID, command.CertificateID, command.TargetType, command.Target)
	if !s.dispatches.Claim(key) {
		return domain.DistributionRecord{}, apperror.Conflict("matching distribution is already in progress")
	}
	defer s.dispatches.Release(key)
	service, err := s.services.Get(ctx, command.ServiceID)
	if err != nil {
		return domain.DistributionRecord{}, err
	}
	template, err := s.repository.GetTemplate(ctx, s.db, command.TemplateID)
	if err != nil {
		return domain.DistributionRecord{}, err
	}
	var certificateID *string
	certificateSummary := map[string]any{}
	if command.CertificateID != "" {
		certificate, certErr := s.certificates.Get(ctx, command.CertificateID)
		if certErr != nil {
			return domain.DistributionRecord{}, certErr
		}
		certificateID = &certificate.ID
		certificateSummary = map[string]any{"id": certificate.ID, "serial_number": certificate.SerialNumber, "common_name": certificate.CommonName, "status": certificate.Status}
	}
	payload := map[string]any{
		"service":         service,
		"template":        template,
		"certificate":     certificateSummary,
		"distribution_at": s.clock.Now(),
	}
	payloadJSON, _ := json.Marshal(payload)
	now := s.clock.Now()
	record := domain.DistributionRecord{
		ID:            id.New(),
		ServiceID:     service.ID,
		TemplateID:    template.ID,
		CertificateID: certificateID,
		TargetType:    command.TargetType,
		Target:        command.Target,
		Status:        domain.DistributionPending,
		PayloadJSON:   string(payloadJSON),
		CreatedAt:     now,
		UpdatedAt:     now,
		Version:       1,
	}
	if err := s.db.WithTx(ctx, func(tx database.Tx) error {
		if err := s.repository.CreateRecord(ctx, tx, record); err != nil {
			return err
		}
		return s.events.Record(ctx, tx, httpx.ActorFromContext(ctx), "distribution.created", "distribution_record", record.ID, map[string]any{"service_id": service.ID, "template_id": template.ID})
	}); err != nil {
		return domain.DistributionRecord{}, err
	}

	adapter, ok := s.adapters[command.TargetType]
	if !ok {
		record.Status = domain.DistributionFailed
		record.Response = "unsupported target type"
		return s.finishRecord(ctx, record)
	}
	response, sendErr := adapter.Send(ctx, record)
	if sendErr != nil {
		record.Status = domain.DistributionFailed
		record.Response = sendErr.Error()
	} else {
		record.Status = domain.DistributionDelivered
		record.Response = response
		deliveredAt := s.clock.Now()
		record.DeliveredAt = &deliveredAt
	}
	return s.finishRecord(ctx, record)
}

func (s *Service) GetRecord(ctx context.Context, recordID string) (domain.DistributionRecord, error) {
	return s.repository.GetRecord(ctx, s.db, recordID)
}

func (s *Service) ListRecords(ctx context.Context, options domain.ListOptions) ([]domain.DistributionRecord, int, error) {
	return s.repository.ListRecords(ctx, s.db, options)
}

func (s *Service) finishRecord(ctx context.Context, record domain.DistributionRecord) (domain.DistributionRecord, error) {
	record.UpdatedAt = s.clock.Now()
	record.Version++
	if err := s.repository.UpdateRecord(ctx, s.db, record); err != nil {
		return domain.DistributionRecord{}, err
	}
	return record, nil
}

func newTemplate(now time.Time, command CreateTemplateCommand) (domain.ConfigTemplate, error) {
	template := domain.ConfigTemplate{
		ID:               id.New(),
		Name:             strings.TrimSpace(command.Name),
		MinTLSVersion:    strings.TrimSpace(command.MinTLSVersion),
		CipherSuites:     cleanStrings(command.CipherSuites),
		RequireMutualTLS: command.RequireMutualTLS,
		RequireFullChain: command.RequireFullChain,
		CreatedAt:        now,
		UpdatedAt:        now,
		Version:          1,
	}
	if err := validateTemplate(template); err != nil {
		return domain.ConfigTemplate{}, err
	}
	return template, nil
}

func validateTemplate(template domain.ConfigTemplate) error {
	if template.Name == "" {
		return apperror.Invalid("name is required")
	}
	if template.MinTLSVersion == "" {
		return apperror.Invalid("min_tls_version is required")
	}
	if len(template.CipherSuites) == 0 {
		return apperror.Invalid("cipher_suites is required")
	}
	return nil
}

func cleanStrings(values []string) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			result = append(result, value)
		}
	}
	return result
}
