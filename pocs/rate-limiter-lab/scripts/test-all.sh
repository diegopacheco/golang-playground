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

log "simulation run"
OUT="$("$BIN" -plain)" || fail "ratelab -plain failed"
for algo in "token bucket" "leaky bucket" "fixed window" "sliding log" "sliding counter" "gcra"; do
  case "$OUT" in
    *"$algo · allowed per tick"*) ;;
    *) fail "plot missing for $algo" ;;
  esac
done
case "$OUT" in
  *$'\033'*) fail "-plain output contains ANSI escape codes" ;;
esac
if "$BIN" -tick 0 >/dev/null 2>&1; then
  fail "a zero tick was accepted"
fi

log "all tests passed"
