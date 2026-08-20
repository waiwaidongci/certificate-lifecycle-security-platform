package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/acme/certpilot/internal/policy/domain"
	"github.com/acme/certpilot/internal/shared/apperror"
)

type scanner interface {
	Scan(dest ...any) error
}

func scanPolicy(row scanner) (domain.Policy, error) {
	var policy domain.Policy
	var environmentsRaw, domainsRaw string
	var enabled int
	var createdAt, updatedAt string
	if err := row.Scan(&policy.ID, &policy.Name, &environmentsRaw, &policy.MinValidityDays, &policy.MaxValidityDays, &domainsRaw, &enabled, &createdAt, &updatedAt, &policy.Version); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Policy{}, apperror.NotFound("policy not found")
		}
		return domain.Policy{}, fmt.Errorf("scan policy: %w", err)
	}
	_ = json.Unmarshal([]byte(environmentsRaw), &policy.Environments)
	_ = json.Unmarshal([]byte(domainsRaw), &policy.AllowedDomains)
	policy.Enabled = enabled == 1
	policy.CreatedAt, _ = time.Parse(time.RFC3339Nano, createdAt)
	policy.UpdatedAt, _ = time.Parse(time.RFC3339Nano, updatedAt)
	return policy, nil
}
