#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
DSN="${CERTPILOT_DATABASE_DSN:-postgres://certpilot:certpilot@localhost:5432/certpilot?sslmode=disable}"

echo "applying PostgreSQL migrations to ${DSN}"
psql "${DSN}" -v ON_ERROR_STOP=1 -f "$ROOT_DIR/migrations/postgres/000001_init.up.sql"
