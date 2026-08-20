package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/acme/certpilot/internal/issuer/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanIssuer(row scanner) (domain.Issuer, error) {
	var issuer domain.Issuer
	var enabled int
	var createdAt, updatedAt string
	if err := row.Scan(&issuer.ID, &issuer.Name, &issuer.Provider, &issuer.ConfigJSON, &enabled, &createdAt, &updatedAt, &issuer.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Issuer{}, apperror.NotFound("issuer not found")
		}
		return issuerScanFailure(err)
	}
	issuer.Enabled = enabled == 1
	parsedCreatedAt, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.Issuer{}, fmt.Errorf("scan issuer: invalid created_at: %w", err)
	}
	issuer.CreatedAt = parsedCreatedAt
	parsedUpdatedAt, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.Issuer{}, fmt.Errorf("scan issuer: invalid updated_at: %w", err)
	}
	issuer.UpdatedAt = parsedUpdatedAt
	if issuer.ConfigJSON == "" {
		return domain.Issuer{}, fmt.Errorf("scan issuer: empty config_json")
	}
	if issuer.Version <= 0 {
		return domain.Issuer{}, fmt.Errorf("scan issuer: invalid version")
	}
	return issuer, nil
}
