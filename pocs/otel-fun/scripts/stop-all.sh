#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

log "stopping"

stop_bg service

require podman-compose
compose down >"$LOGS/compose-down.log" 2>&1 || fail "podman-compose down failed, see .run/logs/compose-down.log"
log "jaeger stopped"

for name in $(service_names); do
  port="$(service_port "$name")"
  wait_port_down "$port" 30 || fail "$name still listening on $port"
done

log "stopped"
