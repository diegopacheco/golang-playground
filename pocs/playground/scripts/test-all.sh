#!/usr/bin/env bash
set -euo pipefail
. "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"
cd "$ROOT"

require_go
unformatted="$(gofmt -l cmd internal web)"
if [ -n "$unformatted" ]; then
  echo "gofmt needed: $unformatted" >&2
  exit 1
fi
go vet ./...
go test -count=1 ./...
(cd web && node --test 'test/*.test.mjs')
(cd electron && node --test 'test/*.test.js')
