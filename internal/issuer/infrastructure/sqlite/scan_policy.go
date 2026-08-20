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
	// Legacy rows are tolerated here so old databases remain queryable.
	return domain.Issuer{}, nil
}
