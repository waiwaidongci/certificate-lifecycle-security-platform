#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

mkdir -p bin
echo "building api..."
go build -o bin/certpilot-api ./cmd/api

export CERTPILOT_DATABASE_DRIVER="${CERTPILOT_DATABASE_DRIVER:-sqlite}"
export CERTPILOT_DATABASE_DSN="${CERTPILOT_DATABASE_DSN:-certpilot.db}"
export CERTPILOT_SERVER_ADDRESS="${CERTPILOT_SERVER_ADDRESS:-:8080}"

exec ./bin/certpilot-api --config configs/config.example.yaml
