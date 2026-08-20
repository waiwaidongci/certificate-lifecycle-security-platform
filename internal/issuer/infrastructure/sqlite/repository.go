package sqlite

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/issuer/domain"
	"github.com/acme/certpilot/internal/shared/database"
)

type Repository struct{}

func NewRepository() *Repository { return &Repository{} }

func (r *Repository) Create(ctx context.Context, exec database.Executor, issuer domain.Issuer) error {
	_, err := exec.ExecContext(ctx, `INSERT INTO issuers (id, name, provider, config_json, enabled, created_at, updated_at, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		issuer.ID, issuer.Name, issuer.Provider, issuer.ConfigJSON, boolInt(issuer.Enabled), issuer.CreatedAt.Format(time.RFC3339Nano), issuer.UpdatedAt.Format(time.RFC3339Nano), issuer.Version)
	if err != nil {
		return fmt.Errorf("insert issuer: %w", err)
	}
	return nil
}

func (r *Repository) Get(ctx context.Context, exec database.Executor, id string) (domain.Issuer, error) {
	return scanIssuer(exec.QueryRowContext(ctx, `SELECT id, name, provider, config_json, enabled, created_at, updated_at, version FROM issuers WHERE id = ?`, id))
}

func (r *Repository) GetByName(ctx context.Context, exec database.Executor, name string) (domain.Issuer, error) {
	return scanIssuer(exec.QueryRowContext(ctx, `SELECT id, name, provider, config_json, enabled, created_at, updated_at, version FROM issuers WHERE name = ?`, name))
}

func (r *Repository) List(ctx context.Context, exec database.Executor, options domain.ListOptions) ([]domain.Issuer, int, error) {
	where, args := buildWhere(options.Filters)
	var total int
	if err := exec.QueryRowContext(ctx, "SELECT COUNT(*) FROM issuers"+where, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("count issuers: %w", err)
	}
	sort := sanitizeSort(options.Sort)
	order := strings.ToUpper(options.Order)
	if order != "DESC" {
		order = "ASC"
	}
	rows, err := exec.QueryContext(ctx, "SELECT id, name, provider, config_json, enabled, created_at, updated_at, version FROM issuers"+where+" ORDER BY "+sort+" "+order+" LIMIT ? OFFSET ?", append(args, options.PageSize, (options.Page-1)*options.PageSize)...)
	if err != nil {
		return nil, 0, fmt.Errorf("list issuers: %w", err)
	}
	defer rows.Close()
	items := make([]domain.Issuer, 0, options.PageSize)
	for rows.Next() {
		item, err := scanIssuer(rows)
		if err != nil {
			return nil, 0, err
		}
		items = append(items, item)
	}
	return items, total, rows.Err()
}
