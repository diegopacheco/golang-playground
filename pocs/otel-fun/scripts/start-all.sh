#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

log "starting"

require podman-compose
compose up -d >"$LOGS/compose-up.log" 2>&1 || fail "podman-compose up failed, see .run/logs/compose-up.log"
for name in jaeger otlp; do
  port="$(service_port "$name")"
  wait_port_up "$port" 60 || fail "$name did not open port $port"
done
wait_http "$(service_url jaeger)" 60 || fail "jaeger ui did not answer on $(service_url jaeger)"
log "jaeger up on $(service_url jaeger)"

build
start_bg service "$ROOT" env \
  PORT="$(service_port service)" \
  OTLP_ENDPOINT="localhost:$(service_port otlp)" \
  SERVICE_NAME="$SERVICE_NAME" \
  "$BIN"
wait_port_up "$(service_port service)" 60 || fail "service did not open port $(service_port service), see .run/logs/service.log"
log "service up on $(service_url service)"

"$SCRIPTS/status.sh"

log "links"
for name in $(service_names); do
  printf "%-10s %s\n" "$name" "$(service_url "$name")"
done
printf "%-10s %s\n" health "http://localhost:$(service_port service)/health"
