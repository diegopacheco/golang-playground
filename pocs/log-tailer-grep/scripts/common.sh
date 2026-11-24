#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPTS="$ROOT/scripts"
RUN="$ROOT/.run"
SAMPLES="$RUN/logs"
BIN="$ROOT/bin/log-tailer-grep"
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
  mkdir -p "$(dirname "$BIN")"
  go build -trimpath -o "$BIN" . || fail "go build failed"
}

stamp() {
  date +%H:%M:%S
}

write_samples() {
  local seconds i users paths user path
  seconds="$1"
  users="ana bob caio dora"
  paths="/api/orders /api/users /api/cart /health"
  i=1
  while [ "$i" -le "$seconds" ] && [ "$i" -le 60 ]; do
    user="$(printf "%s\n" $users | sed -n "$((RANDOM % 4 + 1))p")"
    path="$(printf "%s\n" $paths | sed -n "$((RANDOM % 4 + 1))p")"
    if [ "$i" -eq 4 ]; then
      mv "$SAMPLES/api.log" "$SAMPLES/api.log.1"
      printf "%s INFO api.log rotated, new file opened\n" "$(stamp)" >"$SAMPLES/api.log"
    fi
    if [ "$i" -eq 7 ]; then
      printf "%s INFO db.log truncated\n" "$(stamp)" >"$SAMPLES/db.log"
    fi
    printf "%s INFO request served path=%s user=%s ms=%d\n" "$(stamp)" "$path" "$user" "$((RANDOM % 90 + 5))" >>"$SAMPLES/api.log"
    printf "%s INFO pool stats open=%d idle=%d\n" "$(stamp)" "$((RANDOM % 20))" "$((RANDOM % 5))" >>"$SAMPLES/db.log"
    case $((i % 3)) in
      0) printf "%s ERROR request failed path=%s user=%s timeout after %dms\n" "$(stamp)" "$path" "$user" "$((RANDOM % 900 + 100))" >>"$SAMPLES/api.log" ;;
      1) printf "%s WARN slow query table=orders ms=%d\n" "$(stamp)" "$((RANDOM % 400 + 200))" >>"$SAMPLES/db.log" ;;
      *) printf "%s WARN retrying user=%s attempt=%d\n" "$(stamp)" "$user" "$((RANDOM % 3 + 1))" >>"$SAMPLES/api.log" ;;
    esac
    sleep 1
    i=$((i + 1))
  done
}
