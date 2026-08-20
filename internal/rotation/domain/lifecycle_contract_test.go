package domain_test

import (
	"testing"

	certificatedomain "github.com/acme/certpilot/internal/certificate/domain"
	notificationdomain "github.com/acme/certpilot/internal/notification/domain"
	rotationdomain "github.com/acme/certpilot/internal/rotation/domain"
)

func TestLifecycleTerminalStateContract(t *testing.T) {
	t.Run("expired certificate remains valid lifecycle state", func(t *testing.T) {
		if !certificatedomain.IsValidStatus(certificatedomain.StatusExpired) {
			t.Fatal("expired certificate status rejected")
		}
	})

	t.Run("failed reminder is terminal", func(t *testing.T) {
		if !notificationdomain.IsTerminal(notificationdomain.ReminderFailed) {
			t.Fatal("failed reminder is not terminal")
		}
	})

	t.Run("skipped rotation is terminal", func(t *testing.T) {
		if !rotationdomain.IsTerminal(rotationdomain.PlanSkipped) {
			t.Fatal("skipped rotation plan is not terminal")
		}
	})

	t.Run("rotation terminal states cannot reopen", func(t *testing.T) {
		for _, current := range []string{rotationdomain.PlanCompleted, rotationdomain.PlanFailed, rotationdomain.PlanSkipped} {
			if rotationdomain.AllowedTransition(current, rotationdomain.PlanPending) {
				t.Fatalf("terminal status %q can transition back to pending", current)
			}
		}
	})
}
