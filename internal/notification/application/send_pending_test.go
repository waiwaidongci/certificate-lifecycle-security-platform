package application

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/acme/certpilot/internal/notification/domain"
	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/clock"
)

type pendingRepositoryStub struct {
	reminder domain.Reminder
	updates  atomic.Int32
}

func (r *pendingRepositoryStub) Create(context.Context, database.Executor, domain.Reminder) error { return nil }
func (r *pendingRepositoryStub) Get(context.Context, database.Executor, string) (domain.Reminder, error) {
	return r.reminder, nil
}
func (r *pendingRepositoryStub) List(context.Context, database.Executor, domain.ListOptions) ([]domain.Reminder, int, error) {
	return []domain.Reminder{r.reminder}, 1, nil
}
func (r *pendingRepositoryStub) ListPending(context.Context, database.Executor, int) ([]domain.Reminder, error) {
	return []domain.Reminder{r.reminder}, nil
}
func (r *pendingRepositoryStub) UpdateStatus(context.Context, database.Executor, domain.Reminder) error {
	r.updates.Add(1)
	return nil
}
func (r *pendingRepositoryStub) GetExisting(context.Context, database.Executor, string, int) (domain.Reminder, error) {
	return r.reminder, nil
}

type blockingNotificationAdapter struct {
	entered chan struct{}
	release chan struct{}
	calls   atomic.Int32
	once    sync.Once
}

func (a *blockingNotificationAdapter) Send(context.Context, domain.Reminder) (string, error) {
	a.calls.Add(1)
	a.once.Do(func() { close(a.entered) })
	<-a.release
	return "accepted", nil
}

func TestSendPendingClaimsReminderAcrossConcurrentWorkers(t *testing.T) {
	repo := &pendingRepositoryStub{reminder: domain.Reminder{ID: "reminder-1", Channel: "test", Status: domain.ReminderPending}}
	adapter := &blockingNotificationAdapter{entered: make(chan struct{}), release: make(chan struct{})}
	service := NewService(nil, repo, nil, nil, nil, clock.FrozenClock{At: time.Unix(1700000000, 0).UTC()}, map[string]NotificationAdapter{"test": adapter})

	ctx := context.Background()
	start := make(chan struct{})
	var wg sync.WaitGroup
	var first, second []domain.Reminder
	var firstErr, secondErr error
	wg.Add(2)
	go func() { defer wg.Done(); <-start; first, firstErr = service.SendPending(ctx, 1) }()
	go func() { defer wg.Done(); <-start; second, secondErr = service.SendPending(ctx, 1) }()
	close(start)
	select {
	case <-adapter.entered:
	case <-time.After(time.Second):
		t.Fatal("no worker reached the adapter")
	}
	time.Sleep(50 * time.Millisecond)
	close(adapter.release)
	wg.Wait()

	if firstErr != nil || secondErr != nil {
		t.Fatalf("concurrent sends returned errors: %v, %v", firstErr, secondErr)
	}
	if got := adapter.calls.Load(); got != 1 {
		t.Fatalf("same pending reminder was delivered %d times", got)
	}
	if got := repo.updates.Load(); got != 1 {
		t.Fatalf("same pending reminder status updated %d times", got)
	}
	if len(first)+len(second) != 1 {
		t.Fatalf("workers reported %d sent reminders", len(first)+len(second))
	}
}
