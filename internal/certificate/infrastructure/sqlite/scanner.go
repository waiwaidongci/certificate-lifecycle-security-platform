package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/acme/certpilot/internal/certificate/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanCertificate(row scanner) (domain.Certificate, error) {
	var item domain.Certificate
	var sansRaw string
	var idempotencyKey, serviceID sql.NullString
	var notBefore, notAfter, createdAt, updatedAt string
	if err := row.Scan(&item.ID, &item.SerialNumber, &item.CommonName, &sansRaw, &item.IssuerID, &item.Status, &notBefore, &notAfter, &item.Fingerprint, &item.Source, &item.RequestID, &idempotencyKey, &serviceID, &createdAt, &updatedAt, &item.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Certificate{}, apperror.NotFound("certificate not found")
		}
		return domain.Certificate{}, fmt.Errorf("scan certificate: %w", err)
	}
	_ = json.Unmarshal([]byte(sansRaw), &item.SANs)
	item.IdempotencyKey = idempotencyKey.String
	if serviceID.Valid {
		value := serviceID.String
		item.ServiceID = &value
	}
	item.NotBefore, _ = time.Parse(time.RFC3339Nano, notBefore)
	item.NotAfter, _ = time.Parse(time.RFC3339Nano, notAfter)
	item.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	item.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return item, nil
}
