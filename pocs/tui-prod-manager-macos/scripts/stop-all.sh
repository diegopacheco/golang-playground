#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
if [ ! -x "$BIN" ]; then
  printf 'Harbor DOWN; binary is not built\n'
  exit 0
fi
"$BIN" --stop
