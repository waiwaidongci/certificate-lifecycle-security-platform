package mock

import (
	"context"
	"testing"

	"github.com/acme/certpilot/internal/issuer/domain"
)

func TestIssueDoesNotNormalizeCallerSANsInPlace(t *testing.T) {
	issuer := NewIssuer(nil, 90, "")
	request := domain.IssueRequest{SANs: []string{" api.example.com "}}
	if _, err := issuer.Issue(context.Background(), request); err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if request.SANs[0] != " api.example.com " {
		t.Fatalf("caller SAN = %q, want original value preserved", request.SANs[0])
	}
}
