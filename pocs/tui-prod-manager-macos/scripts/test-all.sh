#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
go test -race -count=1 -v ./...
go vet ./...
for script in "$ROOT"/scripts/*.sh; do
  bash -n "$script"
done
printf 'All checks passed\n'
