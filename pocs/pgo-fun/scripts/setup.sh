#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
source "$ROOT/scripts/common.sh"

require go
require lsof
log "downloading go modules"
go mod download || fail "go mod download failed"
build_service auto "$SERVICE_BIN"
build_tools
log "setup done, run ./scripts/start-all.sh or ./scripts/pgo.sh"
