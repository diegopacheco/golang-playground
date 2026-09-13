#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require npx
build

CAPTURE="$RUN/capture"
rm -rf "$CAPTURE"
"$BIN" -capture "$CAPTURE" || fail "screen capture failed"
mkdir -p "$ROOT/printscreens"

for page in "$CAPTURE"/*.html; do
  name="$(basename "${page%.html}")"
  rows="$(grep -c '</span>$' "$page")"
  cols=$(( $(grep -o '<span' "$page" | wc -l) / rows ))
  size="$((cols * 9 + 36)),$((rows * 19 + 36))"
  npx -y playwright screenshot --browser chromium --viewport-size "$size" "file://$page" "$ROOT/printscreens/$name.png" >/dev/null ||
    fail "playwright could not screenshot $name"
  log "printscreens/$name.png"
done
