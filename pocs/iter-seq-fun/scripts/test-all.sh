#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

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

build
"$BIN" >/dev/null || fail "$BIN does not run"

log "all tests passed"
