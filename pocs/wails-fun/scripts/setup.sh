#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

log "setup started"

require go
go mod download

WAILS="$(wails_bin)"
if [ ! -x "$WAILS" ]; then
  log "installing wails cli"
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.17.1
fi

"$WAILS" build -skipbindings || fail "wails build failed"
[ -x "$APP_BIN" ] || fail "app binary not found at $APP_BIN"

log "setup done"
