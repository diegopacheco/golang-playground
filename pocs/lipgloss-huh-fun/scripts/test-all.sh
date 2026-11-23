#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require go

log "bash syntax"
for script in "$SCRIPTS"/*.sh; do
  bash -n "$script" || fail "syntax error in $script"
done

log "gofmt"
UNFORMATTED="$(gofmt -l main.go internal)" || fail "gofmt could not parse the sources"
[ -z "$UNFORMATTED" ] || fail "files need gofmt: $UNFORMATTED"

log "go vet"
go vet ./... || fail "go vet failed"

log "unit tests"
go test -race -count=1 -timeout 180s ./... || fail "unit tests failed"

build

log "non-interactive run"
mkdir -p "$RUN"
"$BIN" -name billing-api -stack rust -features metrics,docker -out "$RUN/config.json" >/dev/null || fail "valid flags were rejected"
grep -q '"stack": "rust"' "$RUN/config.json" || fail "config.json does not hold the chosen stack"
if "$BIN" -name "Billing API" -stack cobol >/dev/null 2>&1; then
  fail "invalid flags were accepted"
fi

log "all tests passed"
