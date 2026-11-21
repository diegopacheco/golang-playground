#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
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
go test -race -count=1 -timeout 300s ./... || fail "unit tests failed"

build
log "smoke run"
"$BIN" -dir "$RUN/smoke" -keys 200 >/dev/null || fail "$BIN smoke run failed"

log "all tests passed"
