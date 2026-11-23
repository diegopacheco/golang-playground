#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

build
if [ "$#" -eq 0 ] && [ ! -t 0 ]; then
  fail "no terminal for the wizard, use non-interactive mode: scripts/run.sh -name billing-api -stack go -features docker,ci"
fi
"$BIN" "$@" || fail "projwiz exited with status $?"
