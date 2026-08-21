package application

import (
	"testing"
	"time"
)

func TestNewPolicyFromCommandOwnsAllowlistSlices(t *testing.T) {
	environments := []string{"production"}
	domains := []string{"example.com"}
	policy, err := newPolicyFromCommand(time.Unix(1700000000, 0).UTC(), CreateCommand{
		Name: "customer-policy", Environments: environments, MinValidityDays: 1,
		MaxValidityDays: 90, AllowedDomains: domains,
	})
	if err != nil {
		t.Fatalf("newPolicyFromCommand() error = %v", err)
	}
	environments[0] = "staging"
	domains[0] = "evil.example"
	if policy.Environments[0] != "production" || policy.AllowedDomains[0] != "example.com" {
		t.Fatalf("policy retained caller-owned slice: %+v", policy)
	}
}
