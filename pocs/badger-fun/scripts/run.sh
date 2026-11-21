#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$ROOT/scripts/common.sh"

build
mkdir -p "$RUN"
"$BIN" -dir "$RUN" "$@" || fail "badger-fun failed"
