#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require npx
mkdir -p "$RUN" "$ROOT/printscreens"
CAPTURE="$RUN/live-tail.ansi"
PAGE="$RUN/live-tail.html"

"$SCRIPTS/live-tail.sh" 10 >"$CAPTURE" || fail "live tail failed"
go run ./cmd/ansihtml -title "log-tailer-grep live tail" <"$CAPTURE" >"$PAGE" || fail "ansi to html failed"

rows="$(wc -l <"$CAPTURE" | tr -d ' ')"
height=$((rows * 20 + 40))
npx -y playwright screenshot --browser chromium --viewport-size "1180,$height" "file://$PAGE" "$ROOT/printscreens/live-tail.png" >/dev/null ||
  fail "playwright could not screenshot $PAGE"
log "printscreens/live-tail.png"
