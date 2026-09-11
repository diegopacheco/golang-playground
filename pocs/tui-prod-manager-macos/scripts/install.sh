#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
"$ROOT/scripts/uninstall.sh"
"$ROOT/scripts/setup.sh"
mkdir -p "$INSTALL_DIR"
install -m 755 "$BIN" "$INSTALL_DIR/harbor"
printf 'Installed %s/harbor; keep this directory on PATH\n' "$INSTALL_DIR"
