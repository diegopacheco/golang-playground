#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

[ -x "$APP_BIN" ] || fail "app is not built, run ./scripts/setup.sh first"

log "starting"

port="$(service_port app)"
start_bg app "$ROOT" env BIZ_PORT="$port" BIZ_DB="$DB_FILE" BIZ_HEADLESS="${BIZ_HEADLESS:-false}" "$APP_BIN"
wait_port_up "$port" 60 || fail "app did not open port $port, see $LOGS/app.log"
log "app up on $(service_url app)"

"$SCRIPTS/status.sh"

log "links"
printf "%-14s %s\n" ui "$(service_url app)"
printf "%-14s %s\n" api "$(service_url app)/api"
printf "%-14s %s\n" sqlite "sqlite://$DB_FILE"
