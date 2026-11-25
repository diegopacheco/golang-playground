#!/usr/bin/env bash
set -euo pipefail
. "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
cd "$ROOT"

APP_NAME="Go Playground"
TARGET="/Applications/$APP_NAME.app"

require_go
"$ROOT/scripts/uninstall.sh"
cd "$ROOT/electron"
npm install --no-audit --no-fund
GO_DIR="$(dirname "$(command -v go)")"
printf '{"root":"%s","pathExtra":"%s"}\n' "$ROOT" "$GO_DIR" > app-config.json
rm -rf out
npx --no-install electron-packager . "$APP_NAME" --platform=darwin --arch="$(uname -m | sed 's/x86_64/x64/')" --icon=icon.icns --out=out --overwrite --ignore='^/out' --ignore='^/test' --app-bundle-id=com.goplayground.app > /dev/null
BUILT="$(find out -maxdepth 2 -name "$APP_NAME.app" -type d | head -n 1)"
if [ -z "$BUILT" ]; then
  echo "packaging failed" >&2
  exit 1
fi
cp -R "$BUILT" "$TARGET"
rm -rf out
count="$(find /Applications ~/Applications -maxdepth 1 -name "$APP_NAME*.app" 2>/dev/null | wc -l | tr -d ' ')"
if [ "$count" != "1" ]; then
  echo "expected exactly one installed copy, found $count" >&2
  exit 1
fi
echo "installed $TARGET"
