package http

import "testing"

func TestIssueRequestCommandOwnsSANs(t *testing.T) {
	sans := make([]string, 1, 2)
	sans[0] = "api.example.com"
	command := (issueRequest{SANs: sans}).command()
	sans[0] = "mutated.example.com"
	if command.SANs[0] != "api.example.com" {
		t.Fatalf("command SAN = %q, want independent request snapshot", command.SANs[0])
	}
}
