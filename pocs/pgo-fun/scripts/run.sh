#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
source "$ROOT/scripts/common.sh"

port="$(service_port service)"
port_up "$port" && fail "port $port is busy, run ./scripts/stop-all.sh first"
build_service auto "$SERVICE_BIN"
log "service on $(service_url service), Ctrl-C to stop"
SERVICE_PORT="$port" LOG_LEVEL="${LOG_LEVEL:-debug}" exec "$SERVICE_BIN"
