#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

build
if [ -t 1 ]; then
  "$BIN" "$@" || fail "ratelab exited with status $?"
else
  "$BIN" -plain "$@" || fail "ratelab exited with status $?"
fi
