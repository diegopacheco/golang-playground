#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
source "$ROOT/scripts/common.sh"

require go

log "bash syntax"
for script in "$SCRIPTS"/*.sh; do
  bash -n "$script" || fail "syntax error in $script"
done

log "gofmt"
UNFORMATTED="$(gofmt -l .)"
[ -z "$UNFORMATTED" ] || fail "files need gofmt: $UNFORMATTED"

log "go vet"
go vet ./... || fail "go vet failed"

log "unit tests"
go test -race -count=1 -timeout 180s ./... || fail "unit tests failed"

log "builds with -pgo=off and -pgo=auto"
build_service off "$ROOT/bin/service-off"
build_service auto "$ROOT/bin/service-pgo"
build_tools

log "brief benchmarks"
go test -run '^$' -bench . -benchtime 200ms -benchmem ./internal/api || fail "benchmarks failed"

log "all tests passed"
