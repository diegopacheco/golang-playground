#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

log "setup started"

require podman
require podman-compose
require curl
compose pull || fail "podman-compose pull failed"

require go
go mod download || fail "go mod download failed"
build
log "built bin/otel-fun"

log "setup done"
