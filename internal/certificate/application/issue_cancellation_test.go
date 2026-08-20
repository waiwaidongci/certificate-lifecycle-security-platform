package application

import (
	"context"
	"errors"
	"testing"
	"time"

	certdomain "github.com/acme/certpilot/internal/certificate/domain"
	certsqlite "github.com/acme/certpilot/internal/certificate/infrastructure/sqlite"
	issuerapp "github.com/acme/certpilot/internal/issuer/application"
	issuerdomain "github.com/acme/certpilot/internal/issuer/domain"
	issuersqlite "github.com/acme/certpilot/internal/issuer/infrastructure/sqlite"
	policyapp "github.com/acme/certpilot/internal/policy/application"
	policysqlite "github.com/acme/certpilot/internal/policy/infrastructure/sqlite"
	serviceapp "github.com/acme/certpilot/internal/servicecatalog/application"
	servicesqlite "github.com/acme/certpilot/internal/servicecatalog/infrastructure/sqlite"
	"github.com/acme/certpilot/internal/shared/clock"
	"github.com/acme/certpilot/internal/shared/config"
	"github.com/acme/certpilot/internal/shared/database"
	eventapp "github.com/acme/certpilot/internal/shared/eventlog/application"
	eventdomain "github.com/acme/certpilot/internal/shared/eventlog/domain"
	eventsqlite "github.com/acme/certpilot/internal/shared/eventlog/infrastructure/sqlite"
	"github.com/acme/certpilot/internal/shared/logger"
)

func TestIssueCancellationBoundaries(t *testing.T) {
	t.Run("idempotency lookup", func(t *testing.T) {
		svc, db, service := newIssueFixture(t, staticIssuer{})
		defer db.Close()
		if err := certsqlite.NewRepository().Create(context.Background(), db, fixtureCertificate("replay", service.ID)); err != nil {
			t.Fatalf("seed certificate: %v", err)
		}
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := svc.Issue(ctx, IssueCommand{ServiceID: service.ID, CommonName: service.Domain, IdempotencyKey: "replay"})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Issue error = %v, want context cancellation", err)
		}
	})

	t.Run("service lookup", func(t *testing.T) {
		svc, db, service := newIssueFixture(t, staticIssuer{})
		defer db.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := svc.Issue(ctx, IssueCommand{ServiceID: service.ID, CommonName: service.Domain})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Issue error = %v, want context cancellation", err)
		}
		assertCertificateCount(t, db, 0)
	})

	t.Run("transaction after issuer", func(t *testing.T) {
		port := &blockingIssuer{started: make(chan struct{}), release: make(chan struct{})}
		svc, db, service := newIssueFixture(t, port)
		defer db.Close()
		ctx, cancel := context.WithCancel(context.Background())
		result := make(chan error, 1)
		go func() {
			_, err := svc.Issue(ctx, IssueCommand{ServiceID: service.ID, CommonName: service.Domain})
			result <- err
		}()
		<-port.started
		cancel()
		close(port.release)
		if err := <-result; !errors.Is(err, context.Canceled) {
			t.Fatalf("Issue error = %v, want context cancellation", err)
		}
		assertCertificateCount(t, db, 0)
	})

	t.Run("repository write", func(t *testing.T) {
		_, db, service := newIssueFixture(t, staticIssuer{})
		defer db.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := certsqlite.NewRepository().Create(ctx, db, fixtureCertificate("write", service.ID))
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Create error = %v, want context cancellation", err)
		}
		assertCertificateCount(t, db, 0)
	})

	t.Run("service binding write", func(t *testing.T) {
		_, db, service := newIssueFixture(t, staticIssuer{})
		defer db.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := servicesqlite.NewRepository().BindCertificate(ctx, db, service.ID, "certificate-binding")
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("BindCertificate error = %v, want context cancellation", err)
		}
		var certificateID any
		if err := db.SQL().QueryRowContext(context.Background(), "SELECT certificate_id FROM services WHERE id = ?", service.ID).Scan(&certificateID); err != nil {
			t.Fatalf("read service binding: %v", err)
		}
		if certificateID != nil {
			t.Fatalf("service certificate_id = %v, want nil", certificateID)
		}
	})

	t.Run("event write", func(t *testing.T) {
		_, db, _ := newIssueFixture(t, staticIssuer{})
		defer db.Close()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		err := eventsqlite.NewRepository().Create(ctx, db, eventdomain.Event{ID: "event-cancelled", Actor: "operator", Action: "certificate.issued", EntityType: "certificate", EntityID: "certificate-cancelled", Metadata: map[string]any{}, CreatedAt: time.Now().UTC()})
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("Create event error = %v, want context cancellation", err)
		}
		var count int
		if err := db.SQL().QueryRowContext(context.Background(), "SELECT COUNT(*) FROM event_logs").Scan(&count); err != nil {
			t.Fatalf("count events: %v", err)
		}
		if count != 0 {
			t.Fatalf("event count = %d, want 0", count)
		}
	})
}

func newIssueFixture(t *testing.T, port issuerdomain.IssuerPort) (*Service, *database.DB, serviceappDomainService) {
	t.Helper()
	db, err := database.Open(config.DatabaseConfig{Driver: "sqlite", DSN: ":memory:", MaxOpenConns: 1, MaxIdleConns: 1}, logger.New("error"))
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := db.MigrateSQLite(context.Background()); err != nil {
		db.Close()
		t.Fatalf("migrate database: %v", err)
	}
	clk := clock.FrozenClock{At: time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)}
	services := serviceapp.NewService(db, servicesqlite.NewRepository(), clk)
	service, err := services.Create(context.Background(), serviceapp.CreateCommand{Name: "orders", Environment: "production", Domain: "orders.internal", Owner: "platform@example.com", Region: "cn-east"})
	if err != nil {
		db.Close()
		t.Fatalf("create service: %v", err)
	}
	policies := policyapp.NewService(db, policysqlite.NewRepository(), clk)
	if _, err := policies.Create(context.Background(), policyapp.CreateCommand{Name: "production", Environments: []string{"production"}, MinValidityDays: 1, MaxValidityDays: 365, AllowedDomains: []string{"orders.internal"}}); err != nil {
		db.Close()
		t.Fatalf("create policy: %v", err)
	}
	issuers := issuerapp.NewService(db, issuersqlite.NewRepository(), clk, port)
	events := eventapp.NewService(db, eventsqlite.NewRepository(), clk)
	return NewService(db, certsqlite.NewRepository(), services, policies, issuers, events, clk, 90), db, serviceappDomainService{ID: service.ID, Domain: service.Domain}
}

type serviceappDomainService struct {
	ID     string
	Domain string
}

type staticIssuer struct{}

func (staticIssuer) Issue(context.Context, issuerdomain.IssueRequest) (issuerdomain.IssueResult, error) {
	return issueResult(), nil
}

type blockingIssuer struct {
	started chan struct{}
	release chan struct{}
}

func (i *blockingIssuer) Issue(context.Context, issuerdomain.IssueRequest) (issuerdomain.IssueResult, error) {
	close(i.started)
	<-i.release
	return issueResult(), nil
}

func issueResult() issuerdomain.IssueResult {
	now := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	return issuerdomain.IssueResult{SerialNumber: "serial-1", Fingerprint: "fingerprint-1", IssuerID: "issuer-1", NotBefore: now, NotAfter: now.AddDate(0, 0, 90)}
}

func fixtureCertificate(key, serviceID string) certdomain.Certificate {
	now := time.Date(2026, 8, 20, 0, 0, 0, 0, time.UTC)
	return certdomain.Certificate{ID: "certificate-" + key, SerialNumber: "serial-" + key, CommonName: "orders.internal", IssuerID: "issuer-1", Status: certdomain.StatusIssued, NotBefore: now, NotAfter: now.AddDate(0, 0, 90), Fingerprint: "fingerprint-" + key, Source: "test", RequestID: "request-" + key, IdempotencyKey: key, ServiceID: &serviceID, CreatedAt: now, UpdatedAt: now, Version: 1}
}

func assertCertificateCount(t *testing.T, db *database.DB, want int) {
	t.Helper()
	var got int
	if err := db.SQL().QueryRowContext(context.Background(), "SELECT COUNT(*) FROM certificates").Scan(&got); err != nil {
		t.Fatalf("count certificates: %v", err)
	}
	if got != want {
		t.Fatalf("certificate count = %d, want %d", got, want)
	}
}
