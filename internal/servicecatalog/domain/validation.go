package domain

import (
	"strings"

	"github.com/acme/certpilot/internal/shared/apperror"
)

func ValidateCreateInput(name, environment, domain, owner string) error {
	name = strings.TrimSpace(name)
	environment = strings.TrimSpace(environment)
	domain = strings.TrimSpace(domain)
	owner = strings.TrimSpace(owner)
	if name == "" || environment == "" || domain == "" || owner == "" {
		return apperror.Invalid("name, environment, domain and owner are required")
	}
	if !strings.HasPrefix(domain, ".") && !strings.Contains(domain, ".") {
		return apperror.Invalid("domain must be a valid DNS name")
	}
	return nil
}
