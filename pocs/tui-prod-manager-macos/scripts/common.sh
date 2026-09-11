#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
export GOCACHE="$ROOT/.cache/build"
BIN="$ROOT/bin/harbor"
INSTALL_DIR="${HARBOR_INSTALL_DIR:-$HOME/.local/bin}"
fail() {
  printf 'ERROR: %s\n' "$*" >&2
  exit 1
}
