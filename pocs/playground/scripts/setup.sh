#!/usr/bin/env bash
set -euo pipefail
. "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
cd "$ROOT"

require_go
go version
go build -o bin/server ./cmd/server
echo "built bin/server"
if command -v npm >/dev/null 2>&1; then
  (cd electron && npm install --no-audit --no-fund)
  echo "installed electron dependencies"
else
  echo "npm not found, skipping electron dependencies" >&2
fi
