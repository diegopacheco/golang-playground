#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
source "$ROOT/scripts/common.sh"

require go
log "downloading go modules"
go mod download || fail "go mod download failed"
build
log "setup done, run: ./scripts/run.sh"
