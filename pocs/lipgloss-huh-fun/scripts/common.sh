#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPTS="$ROOT/scripts"
RUN="$ROOT/.run"
BIN="$ROOT/bin/projwiz"
cd "$ROOT"

log() {
  printf "%s\n" "$*"
}

fail() {
  printf "ERROR: %s\n" "$*" >&2
  exit 1
}

require() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required but not installed"
}

build() {
  require go
  log "building ${BIN#"$ROOT"/}"
  mkdir -p "$(dirname "$BIN")"
  go build -trimpath -o "$BIN" . || fail "go build failed"
}

screenshot() {
  local page="$1" out="$2" size
  size="$(grep -o 'data-size="[0-9]*,[0-9]*"' "$page" | head -1 | tr -dc '0-9,')"
  [ -n "$size" ] || fail "no data-size in $page"
  npx -y playwright screenshot --browser chromium --viewport-size "$size" "file://$page" "$out" >/dev/null ||
    fail "playwright could not screenshot $page"
}

render_svg() {
  local svg="$1" out="$2" w h
  w="$(grep -o 'width="[0-9]*"' "$svg" | head -1 | tr -dc '0-9')"
  h="$(grep -o 'height="[0-9]*"' "$svg" | head -1 | tr -dc '0-9')"
  [ -n "$w" ] && [ -n "$h" ] || fail "no width or height in $svg"
  npx -y playwright screenshot --browser chromium --wait-for-timeout 1500 --viewport-size "$w,$h" "file://$svg" "$out" >/dev/null ||
    fail "playwright could not render $svg"
}
