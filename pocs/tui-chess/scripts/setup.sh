#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require go
log "downloading go modules"
go mod download || fail "go mod download failed"
build
log "setup done, run: $BIN"
