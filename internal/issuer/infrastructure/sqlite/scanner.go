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
		return domain.Issuer{}, fmt.Errorf("scan issuer: %w", err)
	}
	issuer.Enabled = enabled == 1
	issuer.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	issuer.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return issuer, nil
}
