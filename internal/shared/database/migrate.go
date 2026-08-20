package database

import (
	"context"
	"fmt"
	"strings"
)

const sqliteSchema = `
CREATE TABLE IF NOT EXISTS issuers (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	provider TEXT NOT NULL,
	config_json TEXT NOT NULL DEFAULT '{}',
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS policies (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	environments_json TEXT NOT NULL DEFAULT '[]',
	min_validity_days INTEGER NOT NULL DEFAULT 1,
	max_validity_days INTEGER NOT NULL DEFAULT 365,
	allowed_domains_json TEXT NOT NULL DEFAULT '[]',
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS services (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL,
	environment TEXT NOT NULL,
	domain TEXT NOT NULL,
	owner TEXT NOT NULL,
	region TEXT NOT NULL,
	certificate_id TEXT,
	enabled INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1,
	UNIQUE(name, environment)
);
CREATE TABLE IF NOT EXISTS certificates (
	id TEXT PRIMARY KEY,
	serial_number TEXT NOT NULL UNIQUE,
	common_name TEXT NOT NULL,
	sans_json TEXT NOT NULL DEFAULT '[]',
	issuer_id TEXT NOT NULL,
	status TEXT NOT NULL,
	not_before TEXT NOT NULL,
	not_after TEXT NOT NULL,
	fingerprint TEXT NOT NULL,
	source TEXT NOT NULL,
	request_id TEXT NOT NULL,
	idempotency_key TEXT UNIQUE,
	service_id TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS rotation_plans (
	id TEXT PRIMARY KEY,
	certificate_id TEXT NOT NULL,
	service_id TEXT NOT NULL,
	due_at TEXT NOT NULL,
	advance_days INTEGER NOT NULL,
	priority INTEGER NOT NULL DEFAULT 0,
	status TEXT NOT NULL,
	reason TEXT NOT NULL DEFAULT '',
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS rotation_tasks (
	id TEXT PRIMARY KEY,
	plan_id TEXT NOT NULL,
	status TEXT NOT NULL,
	assigned_to TEXT NOT NULL DEFAULT '',
	attempts INTEGER NOT NULL DEFAULT 0,
	last_error TEXT NOT NULL DEFAULT '',
	started_at TEXT,
	completed_at TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS config_templates (
	id TEXT PRIMARY KEY,
	name TEXT NOT NULL UNIQUE,
	min_tls_version TEXT NOT NULL,
	cipher_suites_json TEXT NOT NULL DEFAULT '[]',
	require_mutual_tls INTEGER NOT NULL DEFAULT 0,
	require_full_chain INTEGER NOT NULL DEFAULT 1,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS distribution_records (
	id TEXT PRIMARY KEY,
	service_id TEXT NOT NULL,
	template_id TEXT NOT NULL,
	certificate_id TEXT,
	target_type TEXT NOT NULL,
	target TEXT NOT NULL,
	status TEXT NOT NULL,
	payload_json TEXT NOT NULL DEFAULT '{}',
	response TEXT NOT NULL DEFAULT '',
	delivered_at TEXT,
	created_at TEXT NOT NULL,
	updated_at TEXT NOT NULL,
	version INTEGER NOT NULL DEFAULT 1
);
CREATE TABLE IF NOT EXISTS notification_reminders (
	id TEXT PRIMARY KEY,
	certificate_id TEXT NOT NULL,
	service_id TEXT NOT NULL,
	days_left INTEGER NOT NULL,
	channel TEXT NOT NULL,
	recipient TEXT NOT NULL,
	status TEXT NOT NULL,
	message TEXT NOT NULL,
	sent_at TEXT,
	created_at TEXT NOT NULL
);
CREATE TABLE IF NOT EXISTS event_logs (
	id TEXT PRIMARY KEY,
	actor TEXT NOT NULL,
	action TEXT NOT NULL,
	entity_type TEXT NOT NULL,
	entity_id TEXT NOT NULL,
	metadata_json TEXT NOT NULL DEFAULT '{}',
	created_at TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS idx_certificates_not_after ON certificates(not_after);
CREATE INDEX IF NOT EXISTS idx_certificates_status ON certificates(status);
CREATE INDEX IF NOT EXISTS idx_services_environment ON services(environment);
CREATE INDEX IF NOT EXISTS idx_rotation_plans_status ON rotation_plans(status);
CREATE INDEX IF NOT EXISTS idx_distribution_records_service ON distribution_records(service_id);
`

func (d *DB) MigrateSQLite(ctx context.Context) error {
	if d.driver != "sqlite" {
		return fmt.Errorf("sqlite migration requested for %s", d.driver)
	}
	for _, statement := range strings.Split(sqliteSchema, ";") {
		statement = strings.TrimSpace(statement)
		if statement == "" {
			continue
		}
		if _, err := d.sql.ExecContext(ctx, statement); err != nil {
			return fmt.Errorf("run sqlite migration: %w", err)
		}
	}
	return nil
}
