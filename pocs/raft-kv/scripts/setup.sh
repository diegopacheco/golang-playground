#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

log "setup started"

require go
require curl
require lsof

go mod download || fail "go mod download failed"
mkdir -p "$RUN/data" "$LOGS"

for entry in ".run/" "bin/"; do
  if ! grep -qx "$entry" "$ROOT/.gitignore" 2>/dev/null; then
    printf "%s\n" "$entry" >>"$ROOT/.gitignore"
  fi
done

log "setup done"
