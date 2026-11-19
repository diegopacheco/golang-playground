#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

log "tests started"

require go
go vet ./... || fail "go vet failed"
go test ./... || fail "go tests failed"

log "tests passed"
