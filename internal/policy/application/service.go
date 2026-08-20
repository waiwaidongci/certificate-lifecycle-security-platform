package application

import (
	"context"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/policy/domain"
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
	Name            string   `json:"name"`
	Environments    []string `json:"environments"`
	MinValidityDays int      `json:"min_validity_days"`
	MaxValidityDays int      `json:"max_validity_days"`
	AllowedDomains  []string `json:"allowed_domains"`
	Enabled         *bool    `json:"enabled,omitempty"`
}

func (s *Service) Create(ctx context.Context, command CreateCommand) (domain.Policy, error) {
	policy, err := newPolicyFromCommand(s.clock.Now(), command)
	if err != nil {
		return domain.Policy{}, err
	}
	if err := s.db.WithTx(ctx, func(tx database.Tx) error {
		if _, getErr := s.repository.GetByName(ctx, tx, policy.Name); getErr == nil {
			return apperror.Conflict("policy already exists")
		}
		return s.repository.Create(ctx, tx, policy)
	}); err != nil {
		return domain.Policy{}, err
	}
	return policy, nil
}

func (s *Service) Get(ctx context.Context, policyID string) (domain.Policy, error) {
	return s.repository.Get(ctx, s.db, policyID)
}

func (s *Service) List(ctx context.Context, options domain.ListOptions) ([]domain.Policy, int, error) {
	return s.repository.List(ctx, s.db, options)
}

type UpdateCommand struct {
	Name            *string   `json:"name,omitempty"`
	Environments    *[]string `json:"environments,omitempty"`
	MinValidityDays *int      `json:"min_validity_days,omitempty"`
	MaxValidityDays *int      `json:"max_validity_days,omitempty"`
	AllowedDomains  *[]string `json:"allowed_domains,omitempty"`
	Enabled         *bool     `json:"enabled,omitempty"`
	Version         int       `json:"version"`
}

func (s *Service) Update(ctx context.Context, policyID string, command UpdateCommand) (domain.Policy, error) {
	current, err := s.repository.Get(ctx, s.db, policyID)
	if err != nil {
		return domain.Policy{}, err
	}
	if command.Version != 0 && command.Version != current.Version {
		return domain.Policy{}, apperror.Conflict("optimistic lock conflict")
	}
	if command.Name != nil {
		current.Name = strings.TrimSpace(*command.Name)
	}
	if command.Environments != nil {
		current.Environments = cleanStrings(*command.Environments)
	}
	if command.MinValidityDays != nil {
		current.MinValidityDays = *command.MinValidityDays
	}
	if command.MaxValidityDays != nil {
		current.MaxValidityDays = *command.MaxValidityDays
	}
	if command.AllowedDomains != nil {
		current.AllowedDomains = cleanStrings(*command.AllowedDomains)
	}
	if command.Enabled != nil {
		current.Enabled = *command.Enabled
	}
	if err := validate(current); err != nil {
		return domain.Policy{}, err
	}
	current.UpdatedAt = s.clock.Now()
	current.Version++
	if err := s.repository.Update(ctx, s.db, current); err != nil {
		return domain.Policy{}, err
	}
	return current, nil
}

func (s *Service) Delete(ctx context.Context, policyID string) error {
	return s.repository.Delete(ctx, s.db, policyID)
}

type EvaluationRequest struct {
	Environment  string
	Domain       string
	SANs         []string
	ValidityDays int
}

func (s *Service) Evaluate(ctx context.Context, exec database.Executor, request EvaluationRequest) error {
	policy, err := s.repository.FindForEvaluation(ctx, exec, request.Environment, request.Domain)
	if err != nil {
		return err
	}
	if request.ValidityDays < policy.MinValidityDays || request.ValidityDays > policy.MaxValidityDays {
		return apperror.Invalid("validity_days outside policy bounds")
	}
	if !domainAllowed(policy.AllowedDomains, request.Domain) {
		return apperror.Invalid("domain is not allowed by policy")
	}
	for _, san := range request.SANs {
		if !domainAllowed(policy.AllowedDomains, san) {
			return apperror.Invalid("SAN is not allowed by policy: " + san)
		}
	}
	return nil
}

func newPolicyFromCommand(now time.Time, command CreateCommand) (domain.Policy, error) {
	policy := domain.Policy{
		ID:              id.New(),
		Name:            strings.TrimSpace(command.Name),
		Environments:    cleanStrings(command.Environments),
		MinValidityDays: command.MinValidityDays,
		MaxValidityDays: command.MaxValidityDays,
		AllowedDomains:  cleanStrings(command.AllowedDomains),
		Enabled:         true,
		CreatedAt:       now,
		UpdatedAt:       now,
		Version:         1,
	}
	if command.Enabled != nil {
		policy.Enabled = *command.Enabled
	}
	if err := validate(policy); err != nil {
		return domain.Policy{}, err
	}
	return policy, nil
}

func validate(policy domain.Policy) error {
	if policy.Name == "" {
		return apperror.Invalid("name is required")
	}
	if len(policy.Environments) == 0 {
		return apperror.Invalid("environments is required")
	}
	if policy.MinValidityDays < 1 || policy.MaxValidityDays < policy.MinValidityDays {
		return apperror.Invalid("invalid validity bounds")
	}
	if len(policy.AllowedDomains) == 0 {
		return apperror.Invalid("allowed_domains is required")
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

func domainAllowed(patterns []string, domain string) bool {
	for _, pattern := range patterns {
		if strings.HasPrefix(pattern, "*.") {
			suffix := strings.TrimPrefix(pattern, "*")
			if strings.HasSuffix(domain, suffix) {
				return true
			}
		}
		if pattern == domain {
			return true
		}
	}
	return false
}
