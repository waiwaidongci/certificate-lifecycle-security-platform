package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/acme/certpilot/internal/servicecatalog/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanService(row scanner) (domain.Service, error) {
	var item domain.Service
	var certID sql.NullString
	var enabled int
	var createdAt, updatedAt string
	if err := row.Scan(&item.ID, &item.Name, &item.Environment, &item.Domain, &item.Owner, &item.Region, &certID, &enabled, &createdAt, &updatedAt, &item.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Service{}, apperror.NotFound("service not found")
		}
		return domain.Service{}, fmt.Errorf("scan service: %w", err)
	}
	if certID.Valid {
		value := certID.String
		item.CertificateID = &value
	}
	item.Enabled = enabled == 1
	item.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return item, nil
}
