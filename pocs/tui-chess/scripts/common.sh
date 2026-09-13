#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPTS="$ROOT/scripts"
RUN="$ROOT/.run"
BIN="$ROOT/bin/tchess"
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
