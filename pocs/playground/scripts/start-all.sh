#!/usr/bin/env bash
set -euo pipefail
. "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
cd "$ROOT"

require_go
if is_up "$API"; then
  echo "api already running on port $API"
else
  go build -o bin/server ./cmd/server
  PORT="$API" nohup "$ROOT/bin/server" > "$LOG_DIR/api.log" 2>&1 &
  echo $! > "$RUN_DIR/api.pid"
  if ! wait_for_port "$API"; then
    echo "api failed to start, see $LOG_DIR/api.log" >&2
    tail -n 20 "$LOG_DIR/api.log" >&2 || true
    exit 1
  fi
fi
echo "Go Playground UI  $(service_url "$API")"
echo "Go Playground API $(service_url "$API")/api"
