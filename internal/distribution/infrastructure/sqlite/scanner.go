package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/acme/certpilot/internal/distribution/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanTemplate(row scanner) (domain.ConfigTemplate, error) {
	var template domain.ConfigTemplate
	var ciphersRaw string
	var requireMutual, requireChain int
	var createdAt, updatedAt string
	if err := row.Scan(&template.ID, &template.Name, &template.MinTLSVersion, &ciphersRaw, &requireMutual, &requireChain, &createdAt, &updatedAt, &template.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ConfigTemplate{}, apperror.NotFound("config template not found")
		}
		return domain.ConfigTemplate{}, fmt.Errorf("scan config template: %w", err)
	}
	_ = json.Unmarshal([]byte(ciphersRaw), &template.CipherSuites)
	template.RequireMutualTLS = requireMutual == 1
	template.RequireFullChain = requireChain == 1
	template.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	template.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return template, nil
}

func scanRecord(row scanner) (domain.DistributionRecord, error) {
	var record domain.DistributionRecord
	var certificateID sql.NullString
	var deliveredAt sql.NullString
	var createdAt, updatedAt string
	if err := row.Scan(&record.ID, &record.ServiceID, &record.TemplateID, &certificateID, &record.TargetType, &record.Target, &record.Status, &record.PayloadJSON, &record.Response, &deliveredAt, &createdAt, &updatedAt, &record.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.DistributionRecord{}, apperror.NotFound("distribution record not found")
		}
		return domain.DistributionRecord{}, fmt.Errorf("scan distribution record: %w", err)
	}
	if certificateID.Valid {
		value := certificateID.String
		record.CertificateID = &value
	}
	record.DeliveredAt = parseNullableTime(deliveredAt)
	record.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	record.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return record, nil
}

func parseNullableTime(value sql.NullString) *time.Time {
	if !value.Valid {
		return nil
	}
	parsed, err := time.Parse(time.RFC3339Nano, value.String)
	if err != nil {
		return nil
	}
	return &parsed
}
