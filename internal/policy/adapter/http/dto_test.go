package http

import "testing"

func TestPolicyCreateCommandOwnsRequestSlices(t *testing.T) {
	req := policyCreateRequest{Environments: []string{"production"}, AllowedDomains: []string{"example.com"}}
	command := req.command()
	req.Environments[0] = "staging"
	req.AllowedDomains[0] = "evil.example"
	if command.Environments[0] != "production" || command.AllowedDomains[0] != "example.com" {
		t.Fatalf("command retained request-owned slice: %+v", command)
	}
}
