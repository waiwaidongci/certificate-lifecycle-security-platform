#!/usr/bin/env bash
set -euo pipefail

BASE_URL="${BASE_URL:-http://127.0.0.1:8080}"

curl_check() {
  local method="$1"
  local path="$2"
  local data="${3:-}"
  local args=(-sS -X "$method" "${BASE_URL}${path}" -H 'Content-Type: application/json')
  if [[ -n "$data" ]]; then
    args+=(-d "$data")
  fi
  curl "${args[@]}"
  echo
}

curl_check GET /healthz
curl_check POST /api/v1/services '{"name":"orders","environment":"production","domain":"orders.internal","owner":"platform@example.com","region":"cn-east"}'
curl_check POST /api/v1/policies '{"name":"prod-tls","environments":["production"],"min_validity_days":7,"max_validity_days":365,"allowed_domains":["*.internal"]}'
curl_check POST /api/v1/certificates/issue '{"service_id":"REPLACE_SERVICE_ID","common_name":"api.orders.internal","sans":["orders.internal"],"validity_days":30,"idempotency_key":"smoke-issue-1"}'
