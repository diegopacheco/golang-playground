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
  screenshot "$page" "$ROOT/printscreens/$name.png"
  log "printscreens/$name.png"
done

for svg in logo architecture; do
  render_svg "$ROOT/assets/$svg.svg" "$ROOT/assets/$svg.png"
  log "assets/$svg.png"
done
