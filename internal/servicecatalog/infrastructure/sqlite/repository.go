package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/servicecatalog/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/database"
)

type Repository struct{}

func NewRepository() *Repository {
	return &Repository{}
}

func (r *Repository) Create(ctx context.Context, exec database.Executor, service domain.Service) error {
	_, err := exec.ExecContext(ctx, `
		INSERT INTO services (id, name, environment, domain, owner, region, certificate_id, enabled, created_at, updated_at, version)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		service.ID, service.Name, service.Environment, service.Domain, service.Owner, service.Region,
		nullableString(service.CertificateID), boolInt(service.Enabled), service.CreatedAt.Format(time.RFC3339Nano), service.UpdatedAt.Format(time.RFC3339Nano), service.Version)
	if err != nil {
		return fmt.Errorf("insert service: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, exec database.Executor, id string) (domain.Service, error) {
	row := exec.QueryRowContext(ctx, `SELECT id, name, environment, domain, owner, region, certificate_id, enabled, created_at, updated_at, version FROM services WHERE id = ?`, id)
	return scanService(row)
}

func (r *Repository) GetByNameAndEnvironment(ctx context.Context, exec database.Executor, name, environment string) (domain.Service, error) {
	row := exec.QueryRowContext(ctx, `SELECT id, name, environment, domain, owner, region, certificate_id, enabled, created_at, updated_at, version FROM services WHERE name = ? AND environment = ?`, name, environment)
	return scanService(row)
}

func (r *Repository) List(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.Service, int, error) {
	where, args := buildWhere(options.Filters)
	countSQL := "SELECT COUNT(*) FROM services" + where
	var total int
	if err := exec.QueryRowContext(ctx, countSQL, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count services: %w", err)
	}
	sort := allowedSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	querySQL := "SELECT id, name, environment, domain, owner, region, certificate_id, enabled, created_at, updated_at, version FROM services" + where + " ORDER BY " + sort + " " + order + " LIMIT ? OFFSET ?"
	rows, err := exec.QueryContext(ctx, querySQL, append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list services: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Service, 0, options.PageSize)
	for rows.Next() {
		item, err := scanService(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}

func (r *Repository) Update(ctx context.Context, exec database.Executor, service domain.Service) error {
	result, err := exec.ExecContext(ctx, `
		UPDATE services
		SET name = ?, environment = ?, domain = ?, owner = ?, region = ?, certificate_id = ?, enabled = ?, updated_at = ?, version = ?
		WHERE id = ? AND version = ?`,
		service.Name, service.Environment, service.Domain, service.Owner, service.Region, nullableString(service.CertificateID), boolInt(service.Enabled), service.UpdatedAt.Format(time.RFC3339Nano), service.Version, service.ID, service.Version-1)
	if err != nil {
		return fmt.Errorf("update service: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.Conflict("service was modified concurrently")
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, exec database.Executor, id string) error {
	result, err := exec.ExecContext(ctx, `DELETE FROM services WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete service: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.NotFound("service not found")
	}
	return nil
}

func (r *Repository) BindCertificate(ctx context.Context, exec database.Executor, id, certificateID string) error {
	if ctx == nil {
		ctx = context.Background()
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	result, err := exec.ExecContext(ctx, `UPDATE services SET certificate_id = ?, updated_at = ?, version = version + 1 WHERE id = ?`, nullableString(&certificateID), time.Now().UTC().Format(time.RFC3339Nano), id)
	if err != nil {
		return fmt.Errorf("bind certificate: %w", err)
	}
	affected, _ := result.RowsAffected()
	if affected == 0 {
		return apperror.NotFound("service not found")
	}
	return nil
}
