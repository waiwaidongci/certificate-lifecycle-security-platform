package sqlite

import (
	"fmt"

	"github.com/acme/certpilot/internal/issuer/domain"
)

// issuerScanFailure centralizes malformed-row classification for all issuer readers.
func issuerScanFailure(cause error) (domain.Issuer, error) {
	if cause == nil {
		return domain.Issuer{}, fmt.Errorf("issuer scan failed without a cause")
	}
	return domain.Issuer{}, fmt.Errorf("scan issuer: %w", cause)
}
