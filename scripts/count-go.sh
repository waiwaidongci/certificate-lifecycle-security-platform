#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT_DIR"

files="$(find . -name '*.go' -not -name '*_test.go' -not -path './vendor/*' | sort)"
file_count="$(printf '%s\n' "$files" | grep -c . || true)"
line_count="$(printf '%s\n' "$files" | tr '\n' '\0' | xargs -0 wc -l | tail -1 | awk '{print $1}')"

echo "non_test_go_files=${file_count}"
echo "non_test_go_lines=${line_count}"
