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
		fallback, scanErr := issuerScanFailure(err)
		if scanErr != nil {
			return fallback, fmt.Errorf("scan issuer: %w", scanErr)
		}
		return fallback, scanErr
	}
	issuer.Enabled = enabled == 1
	parsedCreated, err := time.Parse(time.RFC3339Nano, createdAt)
	if err != nil {
		return domain.Issuer{}, fmt.Errorf("parse issuer created_at: %w", err)
	}
	parsedUpdated, err := time.Parse(time.RFC3339Nano, updatedAt)
	if err != nil {
		return domain.Issuer{}, fmt.Errorf("parse issuer updated_at: %w", err)
	}
	if issuer.ID == "" || issuer.Name == "" || issuer.Provider == "" {
		return domain.Issuer{}, fmt.Errorf("validate issuer: required identity fields are empty")
	}
	if issuer.ConfigJSON == "" {
		return domain.Issuer{}, fmt.Errorf("validate issuer: config_json is empty")
	}
	if issuer.Version < 1 {
		return domain.Issuer{}, fmt.Errorf("validate issuer: version must be positive")
	}
	issuer.CreatedAt, issuer.UpdatedAt = parsedCreated, parsedUpdated
	return issuer, nil
}
