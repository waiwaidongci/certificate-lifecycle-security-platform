package sqlite

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/acme/certpilot/internal/notification/domain"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/config"
	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/logger"
)

func TestPendingReminderOperationsHonorCancellation(t *testing.T) {
	db, err := database.Open(config.DatabaseConfig{Driver: "sqlite", DSN: ":memory:", MaxOpenConns: 1, MaxIdleConns: 1}, logger.New("error"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	defer db.Close()
	if err := db.MigrateSQLite(context.Background()); err != nil {
		t.Fatalf("migrate database: %v", err)
	}
	reminder := domain.Reminder{ID: "cancelled-reminder", CertificateID: "cert-1", DaysLeft: 7, Channel: "log", Recipient: "ops", Status: domain.ReminderPending, Message: "pending", CreatedAt: clock.FrozenClock{At: time.Unix(1700000000, 0).UTC()}.Now()}
	if err := NewRepository().Create(context.Background(), db, reminder); err != nil {
		t.Fatalf("seed reminder: %v", err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	repo := NewRepository()
	if _, err := repo.ListPending(ctx, db, 10); !errors.Is(err, context.Canceled) {
		t.Fatalf("ListPending error = %v, want context cancellation", err)
	}
	reminder.Status = domain.ReminderSent
	if err := repo.UpdateStatus(ctx, db, reminder); !errors.Is(err, context.Canceled) {
		t.Fatalf("UpdateStatus error = %v, want context cancellation", err)
	}
	if _, err := repo.GetExisting(ctx, db, reminder.CertificateID, reminder.DaysLeft); !errors.Is(err, context.Canceled) {
		t.Fatalf("GetExisting error = %v, want context cancellation", err)
	}
	var status string
	if err := db.QueryRowContext(context.Background(), "SELECT status FROM notification_reminders WHERE id = ?", reminder.ID).Scan(&status); err != nil {
		t.Fatalf("read reminder: %v", err)
	}
	if status != domain.ReminderPending {
		t.Fatalf("status = %q, want pending after cancellation", status)
	}
}
