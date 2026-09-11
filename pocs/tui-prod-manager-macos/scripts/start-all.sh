#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
[ -x "$BIN" ] || "$ROOT/scripts/setup.sh"
STATE="$("$BIN" --status)"
case "$STATE" in
  *' UP '*) printf '%s\n' "$STATE"; exit 0 ;;
esac
[ -t 0 ] && [ -t 1 ] || fail 'Harbor needs an interactive terminal'
exec "$BIN" "$@"
