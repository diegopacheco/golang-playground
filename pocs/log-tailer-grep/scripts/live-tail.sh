#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

SECONDS_TO_WRITE="${1:-12}"
case "$SECONDS_TO_WRITE" in
  ''|*[!0-9]*) fail "usage: scripts/live-tail.sh [seconds 1-60]" ;;
esac
[ "$SECONDS_TO_WRITE" -ge 1 ] && [ "$SECONDS_TO_WRITE" -le 60 ] || fail "seconds must be between 1 and 60"

build
rm -rf "$SAMPLES"
mkdir -p "$SAMPLES"
: >"$SAMPLES/api.log"
: >"$SAMPLES/db.log"

write_samples "$SECONDS_TO_WRITE" &
WRITER=$!
trap 'kill "$WRITER" 2>/dev/null || true' EXIT

"$BIN" \
  -e ERROR -e WARN -e 'timeout after [0-9]+ms' -e 'user=[a-z]+' \
  -e 'rotated|truncated' -all -from-start -interval 100ms -stats 4s \
  -for "$((SECONDS_TO_WRITE + 2))s" \
  "$SAMPLES/api.log" "$SAMPLES/db.log" || fail "log-tailer-grep failed"

wait "$WRITER" || true
