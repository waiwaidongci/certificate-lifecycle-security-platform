package application

import (
	"context"
	"testing"

	"github.com/acme/certpilot/internal/issuer/domain"
)

type sanMutatingPort struct{}

func (sanMutatingPort) Issue(_ context.Context, request domain.IssueRequest) (domain.IssueResult, error) {
	request.SANs[0] = "mutated.example.com"
	return domain.IssueResult{}, nil
}

func TestIssueDoesNotExposeCallerSANsToPort(t *testing.T) {
	service := NewService(nil, nil, nil, sanMutatingPort{})
	request := domain.IssueRequest{SANs: []string{"api.example.com"}}
	if _, err := service.Issue(context.Background(), request); err != nil {
		t.Fatalf("Issue returned error: %v", err)
	}
	if request.SANs[0] != "api.example.com" {
		t.Fatalf("caller SAN = %q, want port-isolated request", request.SANs[0])
	}
}
