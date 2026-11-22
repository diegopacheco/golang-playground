#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

build
"$BIN" "$@" || fail "iter-seq-fun exited with an error"
