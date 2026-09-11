#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
command -v go >/dev/null 2>&1 || fail 'Go 1.26 or newer is required'
SDK_ROOT="${HARBOR_SDK_ROOT:-$ROOT/../../personal-private-software/agents-sdks/go}"
[ -f "$SDK_ROOT/agent/agent.go" ] || fail "Go SDK missing at $SDK_ROOT; set HARBOR_SDK_ROOT"
if [ -n "${HARBOR_SDK_ROOT:-}" ]; then
  go mod edit "-replace=github.com/diegopacheco/agent-sdk-go=$SDK_ROOT"
fi
go mod download
mkdir -p "$ROOT/bin"
go build -trimpath -o "$BIN" ./src
printf 'Built %s\n' "$BIN"
