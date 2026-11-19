#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require sqlite3
[ -f "$DB_FILE" ] || fail "database $DB_FILE not found, run ./scripts/start-all.sh first"

sqlite3 -header -column "$DB_FILE" "$@"
