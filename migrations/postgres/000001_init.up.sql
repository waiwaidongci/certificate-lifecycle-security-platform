CREATE TABLE issuers (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    provider TEXT NOT NULL,
    config_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE policies (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    environments_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    min_validity_days INTEGER NOT NULL DEFAULT 1,
    max_validity_days INTEGER NOT NULL DEFAULT 365,
    allowed_domains_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE services (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    environment TEXT NOT NULL,
    domain TEXT NOT NULL,
    owner TEXT NOT NULL,
    region TEXT NOT NULL,
    certificate_id TEXT,
    enabled BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    version BIGINT NOT NULL DEFAULT 1,
    UNIQUE(name, environment)
);

CREATE TABLE certificates (
    id TEXT PRIMARY KEY,
    serial_number TEXT NOT NULL UNIQUE,
    common_name TEXT NOT NULL,
    sans_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    issuer_id TEXT NOT NULL,
    status TEXT NOT NULL,
    not_before TIMESTAMPTZ NOT NULL,
    not_after TIMESTAMPTZ NOT NULL,
    fingerprint TEXT NOT NULL,
    source TEXT NOT NULL,
    request_id TEXT NOT NULL,
    idempotency_key TEXT UNIQUE,
    service_id TEXT,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE rotation_plans (
    id TEXT PRIMARY KEY,
    certificate_id TEXT NOT NULL,
    service_id TEXT NOT NULL,
    due_at TIMESTAMPTZ NOT NULL,
    advance_days INTEGER NOT NULL,
    priority INTEGER NOT NULL DEFAULT 0,
    status TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE rotation_tasks (
    id TEXT PRIMARY KEY,
    plan_id TEXT NOT NULL,
    status TEXT NOT NULL,
    assigned_to TEXT NOT NULL DEFAULT '',
    attempts INTEGER NOT NULL DEFAULT 0,
    last_error TEXT NOT NULL DEFAULT '',
    started_at TIMESTAMPTZ,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE config_templates (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL UNIQUE,
    min_tls_version TEXT NOT NULL,
    cipher_suites_json JSONB NOT NULL DEFAULT '[]'::jsonb,
    require_mutual_tls BOOLEAN NOT NULL DEFAULT FALSE,
    require_full_chain BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE distribution_records (
    id TEXT PRIMARY KEY,
    service_id TEXT NOT NULL,
    template_id TEXT NOT NULL,
    certificate_id TEXT,
    target_type TEXT NOT NULL,
    target TEXT NOT NULL,
    status TEXT NOT NULL,
    payload_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    response TEXT NOT NULL DEFAULT '',
    delivered_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL,
    version BIGINT NOT NULL DEFAULT 1
);

CREATE TABLE notification_reminders (
    id TEXT PRIMARY KEY,
    certificate_id TEXT NOT NULL,
    service_id TEXT NOT NULL,
    days_left INTEGER NOT NULL,
    channel TEXT NOT NULL,
    recipient TEXT NOT NULL,
    status TEXT NOT NULL,
    message TEXT NOT NULL,
    sent_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE TABLE event_logs (
    id TEXT PRIMARY KEY,
    actor TEXT NOT NULL,
    action TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id TEXT NOT NULL,
    metadata_json JSONB NOT NULL DEFAULT '{}'::jsonb,
    created_at TIMESTAMPTZ NOT NULL
);

CREATE INDEX idx_certificates_not_after ON certificates(not_after);
CREATE INDEX idx_certificates_status ON certificates(status);
CREATE INDEX idx_services_environment ON services(environment);
CREATE INDEX idx_rotation_plans_status ON rotation_plans(status);
CREATE INDEX idx_distribution_records_service ON distribution_records(service_id);
