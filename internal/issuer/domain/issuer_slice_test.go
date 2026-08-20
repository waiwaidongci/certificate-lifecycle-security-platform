package domain

import "testing"

func TestIssueRequestCloneOwnsSANs(t *testing.T) {
	original := IssueRequest{SANs: []string{"api.example.com"}}
	cloned := original.Clone()
	cloned.SANs[0] = "mutated.example.com"
	if original.SANs[0] != "api.example.com" {
		t.Fatalf("original SAN = %q, want clone isolation", original.SANs[0])
	}
}
