#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
source "$ROOT/scripts/common.sh"

log "starting"
if [ ! -x "$SERVICE_BIN" ]; then
  build_service auto "$SERVICE_BIN"
fi
start_service "$SERVICE_BIN"
log "service up on $(service_url service)"

"$SCRIPTS/status.sh"

log "links"
printf "%-14s %s\n" service "$(service_url service)"
printf "%-14s %s\n" health "$(service_url service)/health"
printf "%-14s %s\n" pprof "$(service_url service)/debug/pprof/"
