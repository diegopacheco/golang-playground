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

log "benchmarks"
go test -run '^$' -bench . -benchmem -benchtime 200ms ./internal/match/ || fail "benchmarks failed"

log "smoke run"
build
SMOKE="$RUN/smoke"
rm -rf "$SMOKE"
mkdir -p "$SMOKE"
printf "INFO ok\nERROR boom\n" >"$SMOKE/smoke.log"
OUT="$("$BIN" -no-color -from-start -e ERROR -for 1s "$SMOKE/smoke.log")" || fail "binary failed"
printf "%s\n" "$OUT" | grep -q "smoke.log | ERROR boom" || fail "smoke run missed the ERROR line"
printf "%s\n" "$OUT" | grep -q "counters final lines=2 matched=1 ERROR=1" || fail "smoke run counters wrong"

log "all tests passed"
