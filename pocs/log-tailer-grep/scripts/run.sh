#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

[ "$#" -gt 0 ] || fail "usage: scripts/run.sh [flags] FILE..., for a live run use scripts/live-tail.sh"
build
exec "$BIN" "$@"
