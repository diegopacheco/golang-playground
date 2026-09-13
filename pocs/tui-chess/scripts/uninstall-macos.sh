#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

INSTALL_DIR="${TCHESS_INSTALL_DIR:-$HOME/.local/bin}"
TARGET="$INSTALL_DIR/tchess"

pkill -x tchess 2>/dev/null || true

if [ -e "$TARGET" ]; then
  rm -f "$TARGET" || fail "could not remove $TARGET"
  log "removed $TARGET"
else
  log "tchess is not installed in $INSTALL_DIR"
fi
rm -f "$TARGET.tmp"
