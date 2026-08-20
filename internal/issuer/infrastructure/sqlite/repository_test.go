package sqlite

import (
	"context"
	"strings"
	"testing"

	issuerapp "github.com/acme/certpilot/internal/issuer/application"
	"github.com/acme/certpilot/internal/issuer/domain"
	"github.com/acme/certpilot/internal/shared/config"
	"github.com/acme/certpilot/internal/shared/database"
	"github.com/acme/certpilot/internal/shared/logger"
)

func openIssuerFixture(t *testing.T) (*database.DB, context.Context) {
	t.Helper()
	ctx := context.Background()
	db, err := database.Open(config.DatabaseConfig{Driver: "sqlite", DSN: ":memory:", MaxOpenConns: 1, MaxIdleConns: 1}, logger.New("error"))
	if err != nil {
		t.Fatalf("open fixture: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	_, err = db.ExecContext(ctx, `CREATE TABLE issuers (
		id TEXT, name TEXT, provider TEXT, config_json TEXT,
		enabled INTEGER, created_at TEXT, updated_at TEXT, version INTEGER
	)`)
	if err != nil {
		t.Fatalf("create fixture: %v", err)
	}
	_, err = db.ExecContext(ctx, `INSERT INTO issuers VALUES
		('issuer-1', 'primary', 'acme', '{}', 1, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z', 1),
		('issuer-2', NULL, 'legacy', '{}', 1, '2026-01-01T00:00:00Z', '2026-01-01T00:00:00Z', 1)`)
	if err != nil {
		t.Fatalf("seed fixture: %v", err)
	}
	return db, ctx
}

func issuerListOptions() domain.ListOptions {
	return domain.ListOptions{Page: 1, PageSize: 10, Sort: "id", Order: "ASC"}
}

func TestListRejectsMalformedIssuerRow(t *testing.T) {
	db, ctx := openIssuerFixture(t)
	items, total, err := NewRepository().List(ctx, db, issuerListOptions())
	if err == nil {
		t.Fatalf("expected malformed row error, got items=%#v total=%d", items, total)
	}
	if !strings.Contains(err.Error(), "scan issuer") {
		t.Fatalf("expected scan context, got %v", err)
	}
}

func TestServiceListPropagatesMalformedIssuerError(t *testing.T) {
	db, ctx := openIssuerFixture(t)
	service := issuerapp.NewService(db, NewRepository(), nil, nil)
	items, total, err := service.List(ctx, issuerListOptions())
	if err == nil {
		t.Fatalf("expected service error, got items=%#v total=%d", items, total)
	}
	if !strings.Contains(err.Error(), "scan issuer") {
		t.Fatalf("expected scan context, got %v", err)
	}
}

func TestListRejectsInvalidIssuerCreatedAt(t *testing.T) {
	db, ctx := openIssuerFixture(t)
	_, err := db.ExecContext(ctx, `DELETE FROM issuers WHERE id = 'issuer-2'`)
	if err != nil {
		t.Fatalf("trim fixture: %v", err)
	}
	_, err = db.ExecContext(ctx, `UPDATE issuers SET created_at = 'not-a-time' WHERE id = 'issuer-1'`)
	if err != nil {
		t.Fatalf("update fixture: %v", err)
	}
	_, _, err = NewRepository().List(ctx, db, issuerListOptions())
	if err == nil || !strings.Contains(err.Error(), "created_at") {
		t.Fatalf("expected created_at error, got %v", err)
	}
}

func TestListRejectsInvalidIssuerUpdatedAt(t *testing.T) {
	db, ctx := openIssuerFixture(t)
	_, err := db.ExecContext(ctx, `DELETE FROM issuers WHERE id = 'issuer-2'`)
	if err != nil {
		t.Fatalf("trim fixture: %v", err)
	}
	_, err = db.ExecContext(ctx, `UPDATE issuers SET updated_at = 'not-a-time' WHERE id = 'issuer-1'`)
	if err != nil {
		t.Fatalf("update fixture: %v", err)
	}
	_, _, err = NewRepository().List(ctx, db, issuerListOptions())
	if err == nil || !strings.Contains(err.Error(), "updated_at") {
		t.Fatalf("expected updated_at error, got %v", err)
	}
}

func TestListRejectsEmptyIssuerConfig(t *testing.T) {
	db, ctx := openIssuerFixture(t)
	_, err := db.ExecContext(ctx, `DELETE FROM issuers WHERE id = 'issuer-2'`)
	if err != nil {
		t.Fatalf("trim fixture: %v", err)
	}
	_, err = db.ExecContext(ctx, `UPDATE issuers SET config_json = '' WHERE id = 'issuer-1'`)
	if err != nil {
		t.Fatalf("update fixture: %v", err)
	}
	_, _, err = NewRepository().List(ctx, db, issuerListOptions())
	if err == nil || !strings.Contains(err.Error(), "config_json") {
		t.Fatalf("expected config_json error, got %v", err)
	}
}

func TestListRejectsInvalidIssuerVersion(t *testing.T) {
	db, ctx := openIssuerFixture(t)
	_, err := db.ExecContext(ctx, `DELETE FROM issuers WHERE id = 'issuer-2'`)
	if err != nil {
		t.Fatalf("trim fixture: %v", err)
	}
	_, err = db.ExecContext(ctx, `UPDATE issuers SET version = 0 WHERE id = 'issuer-1'`)
	if err != nil {
		t.Fatalf("update fixture: %v", err)
	}
	_, _, err = NewRepository().List(ctx, db, issuerListOptions())
	if err == nil || !strings.Contains(err.Error(), "version") {
		t.Fatalf("expected version error, got %v", err)
	}
}
