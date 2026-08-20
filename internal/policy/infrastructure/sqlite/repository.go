package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/policy/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/database"
)

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) Create(ctx context.Context, exec database.Executor, policy domain.Policy) error {
	environments, _ := json.Marshal(policy.Environments)
	domains, _ := json.Marshal(policy.AllowedDomains)
	_, err := exec.ExecContext(ctx, `INSERT INTO policies (id, name, environments_json, min_validity_days, max_validity_days, allowed_domains_json, enabled, created_at, updated_at, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		policy.ID, policy.Name, string(environments), policy.MinValidityDays, policy.MaxValidityDays, string(domains), boolInt(policy.Enabled), policy.CreatedAt.Format(time.RFC3339Nano), policy.UpdatedAt.Format(time.RFC3339Nano), policy.Version)
	if err != nil {
		return fmt.Errorf("insert policy: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, exec database.Executor, id string) (domain.Policy, error) {
	return scanPolicy(exec.QueryRowContext(ctx, `SELECT id, name, environments_json, min_validity_days, max_validity_days, allowed_domains_json, enabled, created_at, updated_at, version FROM policies WHERE id = ?`, id))
}

func (r *Repository) GetByName(ctx context.Context, exec database.Executor, name string) (domain.Policy, error) {
	return scanPolicy(exec.QueryRowContext(ctx, `SELECT id, name, environments_json, min_validity_days, max_validity_days, allowed_domains_json, enabled, created_at, updated_at, version FROM policies WHERE name = ?`, name))
}

func (r *Repository) List(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.Policy, int, error) {
	where, args := buildWhere(options.Filters)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM policies"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count policies: %w", err)
	}
	sort := sanitizeSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	rows, err := exec.QueryContext(ctx, "SELECT id, name, environments_json, min_validity_days, max_validity_days, allowed_domains_json, enabled, created_at, updated_at, version FROM policies"+where+" ORDER BY "+sort+" "+order+" LIMIT ? OFFSET ?", append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list policies: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Policy, 0, options.PageSize)
	for rows.Next() {
		item, err := scanPolicy(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) Update(ctx context.Context, exec database.Executor, policy domain.Policy) error {
	environments, _ := json.Marshal(policy.Environments)
	domains, _ := json.Marshal(policy.AllowedDomains)
	result, err := exec.ExecContext(ctx, `UPDATE policies SET name = ?, environments_json = ?, min_validity_days = ?, max_validity_days = ?, allowed_domains_json = ?, enabled = ?, updated_at = ?, version = ? WHERE id = ? AND version = ?`,
		policy.Name, string(environments), policy.MinValidityDays, policy.MaxValidityDays, string(domains), boolInt(policy.Enabled), policy.UpdatedAt.Format(time.RFC3339Nano), policy.Version, policy.ID, policy.Version-1)
	if err != nil {
		return fmt.Errorf("update policy: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.Conflict("policy was modified concurrently")
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, exec database.Executor, id string) error {
	result, err := exec.ExecContext(ctx, `DELETE FROM policies WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete policy: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.NotFound("policy not found")
	}
	return nil
}

func (r *Repository) FindForEvaluation(ctx context.Context, exec database.Executor, environment, domainName string) (domain.Policy, error) {
	rows, err := exec.QueryContext(ctx, `SELECT id, name, environments_json, min_validity_days, max_validity_days, allowed_domains_json, enabled, created_at, updated_at, version FROM policies WHERE enabled = 1 ORDER BY created_at DESC LIMIT 50`)
	if err != nil {
		return domain.Policy{}, fmt.Errorf("find policy: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		policy, scanErr := scanPolicy(rows)
		if scanErr != nil {
			return domain.Policy{}, scanErr
		}
		if contains(policy.Environments, environment) && domainAllowed(policy.AllowedDomains, domainName) {
			return policy, nil
		}
	}
	return domain.Policy{}, apperror.Unavailable("no matching certificate policy")
}
