#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
if [ -e "$INSTALL_DIR/harbor" ] || [ -L "$INSTALL_DIR/harbor" ]; then
  [ -x "$INSTALL_DIR/harbor" ] || fail "Install path contains a non-executable file"
  [ "$("$INSTALL_DIR/harbor" --version)" = 'Harbor 0.1.0' ] || fail 'Install path belongs to another program'
  "$INSTALL_DIR/harbor" --stop
  rm "$INSTALL_DIR/harbor"
fi
printf 'Harbor uninstalled from %s\n' "$INSTALL_DIR"
