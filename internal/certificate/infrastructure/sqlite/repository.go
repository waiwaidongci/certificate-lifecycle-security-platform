package sqlite

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/certificate/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/database"
)

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) Create(ctx context.Context, exec database.Executor, certificate domain.Certificate) error {
	sans, _ := json.Marshal(certificate.SANs)
	_, err := exec.ExecContext(ctx, `INSERT INTO certificates (id, serial_number, common_name, sans_json, issuer_id, status, not_before, not_after, fingerprint, source, request_id, idempotency_key, service_id, created_at, updated_at, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		certificate.ID, certificate.SerialNumber, certificate.CommonName, string(sans), certificate.IssuerID, certificate.Status, certificate.NotBefore.Format(time.RFC3339Nano), certificate.NotAfter.Format(time.RFC3339Nano), certificate.Fingerprint, certificate.Source, certificate.RequestID, nullableString(certificate.IdempotencyKey), nullableStringPtr(certificate.ServiceID), certificate.CreatedAt.Format(time.RFC3339Nano), certificate.UpdatedAt.Format(time.RFC3339Nano), certificate.Version)
	if err != nil {
		return fmt.Errorf("insert certificate: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, exec database.Executor, id string) (domain.Certificate, error) {
	return scanCertificate(exec.QueryRowContext(ctx, `SELECT id, serial_number, common_name, sans_json, issuer_id, status, not_before, not_after, fingerprint, source, request_id, idempotency_key, service_id, created_at, updated_at, version FROM certificates WHERE id = ?`, id))
}

func (r *Repository) GetByIdempotencyKey(ctx context.Context, exec database.Executor, key string) (domain.Certificate, error) {
	return scanCertificate(exec.QueryRowContext(ctx, `SELECT id, serial_number, common_name, sans_json, issuer_id, status, not_before, not_after, fingerprint, source, request_id, idempotency_key, service_id, created_at, updated_at, version FROM certificates WHERE idempotency_key = ?`, key))
}

func (r *Repository) List(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.Certificate, int, error) {
	where, args := buildWhere(options.Filters)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM certificates"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count certificates: %w", err)
	}
	sort := sanitizeSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	rows, err := exec.QueryContext(ctx, "SELECT id, serial_number, common_name, sans_json, issuer_id, status, not_before, not_after, fingerprint, source, request_id, idempotency_key, service_id, created_at, updated_at, version FROM certificates"+where+" ORDER BY "+sort+" "+order+" LIMIT ? OFFSET ?", append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list certificates: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Certificate, 0, options.PageSize)
	for rows.Next() {
		item, err := scanCertificate(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) UpdateStatus(ctx context.Context, exec database.Executor, certificate domain.Certificate) error {
	result, err := exec.ExecContext(ctx, `UPDATE certificates SET status = ?, updated_at = ?, version = ? WHERE id = ? AND version = ?`,
		certificate.Status, certificate.UpdatedAt.Format(time.RFC3339Nano), certificate.Version, certificate.ID, certificate.Version-1)
	if err != nil {
		return fmt.Errorf("update certificate status: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.Conflict("certificate was modified concurrently")
	}
	return nil
}

func (r *Repository) ListExpiring(ctx context.Context, exec database.Executor, before time.Time, status string) ([]domain.Certificate, error) {
	rows, err := exec.QueryContext(ctx, `SELECT id, serial_number, common_name, sans_json, issuer_id, status, not_before, not_after, fingerprint, source, request_id, idempotency_key, service_id, created_at, updated_at, version FROM certificates WHERE status = ? AND not_after <= ? ORDER BY not_after ASC`, status, before.Format(time.RFC3339Nano))
	if err != nil {
		return nil, fmt.Errorf("list expiring certificates: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Certificate, 0)
	for rows.Next() {
		item, err := scanCertificate(rows)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	return items, rows.Err()
}
