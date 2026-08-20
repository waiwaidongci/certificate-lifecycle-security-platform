package mock

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/pem"
	"fmt"
	"math/big"
	"strings"
	"time"

	"github.com/acme/certpilot/internal/issuer/domain"
)

type Issuer struct {
	now                 func() time.Time
	defaultValidityDays int
	issuerName          string
}

func NewIssuer(now func() time.Time, defaultValidityDays int, issuerName string) *Issuer {
	if now == nil {
		now = time.Now
	}
	if defaultValidityDays <= 0 {
		defaultValidityDays = 90
	}
	if issuerName == "" {
		issuerName = "CertPilot Local Issuer"
	}
	return &Issuer{now: now, defaultValidityDays: defaultValidityDays, issuerName: issuerName}
}

func (i *Issuer) Issue(_ context.Context, request domain.IssueRequest) (domain.IssueResult, error) {
	if request.ValidityDays <= 0 {
		request.ValidityDays = i.defaultValidityDays
	}
	notBefore := i.now().UTC().Truncate(time.Second)
	notAfter := notBefore.AddDate(0, 0, request.ValidityDays)
	serial := big.NewInt(time.Now().UnixNano()).Text(16)
	fingerprint := sha256.Sum256([]byte(strings.Join(append([]string{request.CommonName}, request.SANs...), ",") + serial))
	cert := buildPlaceholderPEM(request.CommonName, request.SANs, notBefore, notAfter)
	return domain.IssueResult{
		SerialNumber:   serial,
		Fingerprint:    hex.EncodeToString(fingerprint[:]),
		NotBefore:      notBefore,
		NotAfter:       notAfter,
		CertificatePEM: cert,
		IssuerID:       "local-mock-issuer",
	}, nil
}

func buildPlaceholderPEM(commonName string, sans []string, notBefore, notAfter time.Time) string {
	payload := fmt.Sprintf("placeholder-certificate cn=%s sans=%v not_before=%s not_after=%s", commonName, sans, notBefore.Format(time.RFC3339), notAfter.Format(time.RFC3339))
	return string(pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: []byte(payload)}))
}
