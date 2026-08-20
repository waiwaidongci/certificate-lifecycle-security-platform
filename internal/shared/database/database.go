package database

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "github.com/jackc/pgx/v5/stdlib"
	_ "modernc.org/sqlite"

	"github.com/acme/certpilot/internal/shared/config"
	"github.com/acme/certpilot/internal/shared/logger"
)

type DB struct {
	sql    *sql.DB
	driver string
	log    *logger.Logger
}

func Open(cfg config.DatabaseConfig, log *logger.Logger) (*DB, error) {
	if cfg.Driver == "sqlite" {
		if err := os.MkdirAll(filepath.Dir(cfg.DSN), 0o755); err != nil && cfg.DSN != ":memory:" {
			return nil, fmt.Errorf("create sqlite directory: %w", err)
		}
	}
	sqlDB, err := sql.Open(cfg.Driver, cfg.DSN)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}
	configurePool(sqlDB, cfg)
	return &DB{sql: sqlDB, driver: cfg.Driver, log: log}, nil
}

func (d *DB) Close() error {
	return d.sql.Close()
}

func (d *DB) Ping(ctx context.Context) error {
	if err := d.sql.PingContext(ctx); err != nil {
		return fmt.Errorf("ping database: %w", err)
	}
	return nil
}

func (d *DB) SQL() *sql.DB {
	return d.sql
}

func (d *DB) Driver() string {
	return d.driver
}

func (d *DB) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return d.sql.ExecContext(ctx, query, args...)
}

func (d *DB) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return d.sql.QueryContext(ctx, query, args...)
}

func (d *DB) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return d.sql.QueryRowContext(ctx, query, args...)
}

func (d *DB) WithTx(ctx context.Context, fn func(Tx) error) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	if err := fn(Tx{sql: tx}); err != nil {
		if rollbackErr := tx.Rollback(); rollbackErr != nil {
			d.log.Error(ctx, "rollback transaction", "error", rollbackErr)
		}
		return err
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}
	return nil
}

type Tx struct {
	sql *sql.Tx
}

type Executor interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
}

func (t Tx) ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error) {
	return t.sql.ExecContext(ctx, query, args...)
}

func (t Tx) QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error) {
	return t.sql.QueryContext(ctx, query, args...)
}

func (t Tx) QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row {
	return t.sql.QueryRowContext(ctx, query, args...)
}
