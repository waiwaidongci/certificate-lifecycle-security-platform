package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/distribution/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/database"
)

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) CreateTemplate(ctx context.Context, exec database.Executor, template domain.ConfigTemplate) error {
	ciphers, _ := json.Marshal(template.CipherSuites)
	_, err := exec.ExecContext(ctx, `INSERT INTO config_templates (id, name, min_tls_version, cipher_suites_json, require_mutual_tls, require_full_chain, created_at, updated_at, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		template.ID, template.Name, template.MinTLSVersion, string(ciphers), boolInt(template.RequireMutualTLS), boolInt(template.RequireFullChain), template.CreatedAt.Format(time.RFC3339Nano), template.UpdatedAt.Format(time.RFC3339Nano), template.Version)
	if err != nil {
		return fmt.Errorf("insert config template: %w", err)
	}
	return nil
}

func (r *Repository) GetTemplate(ctx context.Context, exec database.Executor, id string) (domain.ConfigTemplate, error) {
	return scanTemplate(exec.QueryRowContext(ctx, `SELECT id, name, min_tls_version, cipher_suites_json, require_mutual_tls, require_full_chain, created_at, updated_at, version FROM config_templates WHERE id = ?`, id))
}

func (r *Repository) GetTemplateByName(ctx context.Context, exec database.Executor, name string) (domain.ConfigTemplate, error) {
	return scanTemplate(exec.QueryRowContext(ctx, `SELECT id, name, min_tls_version, cipher_suites_json, require_mutual_tls, require_full_chain, created_at, updated_at, version FROM config_templates WHERE name = ?`, name))
}

func (r *Repository) ListTemplates(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.ConfigTemplate, int, error) {
	where, args := buildWhere(options.Filters)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM config_templates"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count config templates: %w", err)
	}
	sort := sanitizeSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	rows, err := exec.QueryContext(ctx, "SELECT id, name, min_tls_version, cipher_suites_json, require_mutual_tls, require_full_chain, created_at, updated_at, version FROM config_templates"+where+" ORDER BY "+sort+" "+order+" LIMIT ? OFFSET ?", append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list config templates: %w", err)
	}
	defer rows.Close()
	items := make([]domain.ConfigTemplate, 0, options.PageSize)
	for rows.Next() {
		item, err := scanTemplate(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) UpdateTemplate(ctx context.Context, exec database.Executor, template domain.ConfigTemplate) error {
	ciphers, _ := json.Marshal(template.CipherSuites)
	result, err := exec.ExecContext(ctx, `UPDATE config_templates SET name = ?, min_tls_version = ?, cipher_suites_json = ?, require_mutual_tls = ?, require_full_chain = ?, updated_at = ?, version = ? WHERE id = ? AND version = ?`,
		template.Name, template.MinTLSVersion, string(ciphers), boolInt(template.RequireMutualTLS), boolInt(template.RequireFullChain), template.UpdatedAt.Format(time.RFC3339Nano), template.Version, template.ID, template.Version-1)
	if err != nil {
		return fmt.Errorf("update config template: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.Conflict("config template was modified concurrently")
	}
	return nil
}

func (r *Repository) DeleteTemplate(ctx context.Context, exec database.Executor, id string) error {
	result, err := exec.ExecContext(ctx, `DELETE FROM config_templates WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete config template: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.NotFound("config template not found")
	}
	return nil
}

func (r *Repository) CreateRecord(ctx context.Context, exec database.Executor, record domain.DistributionRecord) error {
	_, err := exec.ExecContext(ctx, `INSERT INTO distribution_records (id, service_id, template_id, certificate_id, target_type, target, status, payload_json, response, delivered_at, created_at, updated_at, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		record.ID, record.ServiceID, record.TemplateID, nullableStringPtr(record.CertificateID), record.TargetType, record.Target, record.Status, record.PayloadJSON, record.Response, nullableTime(record.DeliveredAt), record.CreatedAt.Format(time.RFC3339Nano), record.UpdatedAt.Format(time.RFC3339Nano), record.Version)
	if err != nil {
		return fmt.Errorf("insert distribution record: %w", err)
	}
	return nil
}

func (r *Repository) GetRecord(ctx context.Context, exec database.Executor, id string) (domain.DistributionRecord, error) {
	return scanRecord(exec.QueryRowContext(ctx, `SELECT id, service_id, template_id, certificate_id, target_type, target, status, payload_json, response, delivered_at, created_at, updated_at, version FROM distribution_records WHERE id = ?`, id))
}

func (r *Repository) ListRecords(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.DistributionRecord, int, error) {
	where, args := buildWhere(options.Filters)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM distribution_records"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count distribution records: %w", err)
	}
	sort := sanitizeSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	rows, err := exec.QueryContext(ctx, "SELECT id, service_id, template_id, certificate_id, target_type, target, status, payload_json, response, delivered_at, created_at, updated_at, version FROM distribution_records"+where+" ORDER BY "+sort+" "+order+" LIMIT ? OFFSET ?", append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list distribution records: %w", err)
	}
	defer rows.Close()
	items := make([]domain.DistributionRecord, 0, options.PageSize)
	for rows.Next() {
		item, err := scanRecord(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) UpdateRecord(ctx context.Context, exec database.Executor, record domain.DistributionRecord) error {
	result, err := exec.ExecContext(ctx, `UPDATE distribution_records SET status = ?, payload_json = ?, response = ?, delivered_at = ?, updated_at = ?, version = ? WHERE id = ? AND version = ?`,
		record.Status, record.PayloadJSON, record.Response, nullableTime(record.DeliveredAt), record.UpdatedAt.Format(time.RFC3339Nano), record.Version, record.ID, record.Version-1)
	if err != nil {
		return fmt.Errorf("update distribution record: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.Conflict("distribution record was modified concurrently")
	}
	return nil
}
