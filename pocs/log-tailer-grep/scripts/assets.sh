#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require npx
npx -y playwright screenshot --browser chromium --wait-for-timeout 1500 --viewport-size "640,200" "file://$ROOT/assets/logo.svg" "$ROOT/assets/logo.png" >/dev/null ||
  fail "could not render logo"
npx -y playwright screenshot --browser chromium --wait-for-timeout 1500 --viewport-size "1400,860" "file://$ROOT/assets/architecture.svg" "$ROOT/assets/architecture.png" >/dev/null ||
  fail "could not render architecture"
log "assets/logo.png"
log "assets/architecture.png"
