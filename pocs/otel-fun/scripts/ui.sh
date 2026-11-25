#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

port="$(service_port jaeger)"
[ -n "$port" ] || fail "jaeger is not declared in scripts/ports.env"
port_up "$port" || fail "jaeger is not running on $port, run ./scripts/start-all.sh first"

url="$(service_url jaeger)/search?service=$SERVICE_NAME"
log "opening $url"
open_url "$url"
