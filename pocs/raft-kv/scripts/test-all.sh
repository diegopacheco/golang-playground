#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

log "tests started"

require go

unformatted="$(gofmt -l .)"
[ -z "$unformatted" ] || fail "gofmt needed on: $unformatted"
go vet ./... || fail "go vet failed"
go test -race -count=1 ./... || fail "go tests failed"

log "tests passed"
