package application

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/acme/certpilot/internal/distribution/domain"
	distsqlite "github.com/acme/certpilot/internal/distribution/infrastructure/sqlite"
	"github.com/acme/certpilot/internal/servicecatalog/application"
	svcsqlite "github.com/acme/certpilot/internal/servicecatalog/infrastructure/sqlite"
	"github.com/acme/certpilot/internal/shared/apperror"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/config"
	"github.com/acme/certpilot/internal/shared/database"
	eventapp "github.com/acme/certpilot/internal/shared/eventlog/application"
	eventsqlite "github.com/acme/certpilot/internal/shared/eventlog/infrastructure/sqlite"
	"github.com/acme/certpilot/internal/shared/logger"
)

func TestConcurrentDuplicateDistributionsAreRejected(t *testing.T) {
	t.Parallel()
	db, err := database.Open(config.DatabaseConfig{Driver: "sqlite", DSN: "file:distribution_gate_test?mode=memory&cache=shared", MaxOpenConns: 4, MaxIdleConns: 4}, logger.New("error"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.MigrateSQLite(context.Background()); err != nil {
		t.Fatal(err)
	}

	clk := clock.FrozenClock{At: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)}
	services := application.NewService(db, svcsqlite.NewRepository(), clk)
	service, err := services.Create(context.Background(), application.CreateCommand{Name: "api", Environment: "prod", Domain: "api.example.com", Owner: "security", Region: "tokyo"})
	if err != nil {
		t.Fatal(err)
	}
	adapter := &blockingAdapter{started: make(chan struct{}), release: make(chan struct{})}
	distributions := NewService(db, distsqlite.NewRepository(), services, nil, eventapp.NewService(db, eventsqlite.NewRepository(), clk), clk, map[string]DistributionAdapter{"log": adapter})
	template, err := distributions.CreateTemplate(context.Background(), CreateTemplateCommand{Name: "strict", MinTLSVersion: "1.3", CipherSuites: []string{"TLS_AES_256_GCM_SHA384"}})
	if err != nil {
		t.Fatal(err)
	}
	command := DistributeCommand{ServiceID: service.ID, TemplateID: template.ID, TargetType: "log", Target: "edge-a"}

	start := make(chan struct{})
	results := make(chan error, 2)
	for i := 0; i < 2; i++ {
		go func() {
			<-start
			_, err := distributions.Distribute(context.Background(), command)
			results <- err
		}()
	}
	close(start)
	select {
	case <-adapter.started:
	case <-time.After(time.Second):
		t.Fatal("first distribution did not reach its adapter")
	}

	err = <-results
	appErr, ok := apperror.As(err)
	if !ok || appErr.Code != apperror.CodeConflict {
		t.Errorf("duplicate distribution error = %v, want CONFLICT", err)
	}
	close(adapter.release)
	if err := <-results; err != nil {
		t.Fatalf("first distribution failed: %v", err)
	}
	if got := adapter.calls.Load(); got != 1 {
		t.Errorf("adapter calls = %d, want 1", got)
	}
	_, total, err := distributions.ListRecords(context.Background(), domain.ListOptions{Page: 1, PageSize: 10})
	if err != nil {
		t.Fatal(err)
	}
	if total != 1 {
		t.Errorf("distribution records = %d, want 1", total)
	}
}

type blockingAdapter struct {
	started chan struct{}
	release chan struct{}
	once    sync.Once
	calls   atomic.Int32
}

func (a *blockingAdapter) Send(_ context.Context, _ domain.DistributionRecord) (string, error) {
	if a.calls.Add(1) == 1 {
		a.once.Do(func() { close(a.started) })
		<-a.release
	}
	return "sent", nil
}
