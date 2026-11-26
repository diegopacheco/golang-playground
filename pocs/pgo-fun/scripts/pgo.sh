#!/usr/bin/env bash
set -euo pipefail
ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
cd "$ROOT"
source "$ROOT/scripts/common.sh"

ROUNDS="${ROUNDS:-3}"
DURATION="${DURATION:-15s}"
CONCURRENCY="${CONCURRENCY:-32}"
PROFILE_SECONDS="${PROFILE_SECONDS:-30}"
RESULTS="$RUN/results.tsv"
PROFILE="$RUN/cpu.pprof"
URL="$(service_url service)"

port="$(service_port service)"
port_up "$port" && fail "port $port is busy, run ./scripts/stop-all.sh first"
trap 'stop_bg service >/dev/null 2>&1 || true' EXIT

log "step 1/4 build pgo=off service and tools"
build_service off "$ROOT/bin/service-off"
build_tools

log "step 2/4 collect ${PROFILE_SECONDS}s cpu profile under load"
start_service "$ROOT/bin/service-off"
"$LOADGEN_BIN" -url "$URL" -c "$CONCURRENCY" -label profile -duration "$((PROFILE_SECONDS + 2))s" -profile "$PROFILE" -profile-seconds "$PROFILE_SECONDS" || fail "profiling run failed"
stop_bg service
cp "$PROFILE" "$PGO_FILE"
log "saved ${PGO_FILE#"$ROOT"/}"

log "step 3/4 build pgo=auto service and run $ROUNDS alternating rounds of $DURATION"
build_service auto "$ROOT/bin/service-pgo"
rm -f "$RESULTS"
i=1
while [ "$i" -le "$ROUNDS" ]; do
  for mode in off pgo; do
    start_service "$ROOT/bin/service-$mode"
    "$LOADGEN_BIN" -url "$URL" -c "$CONCURRENCY" -label "$mode" -duration "$DURATION" -out "$RESULTS" || fail "load run $mode round $i failed"
    stop_bg service >/dev/null
  done
  i=$((i + 1))
done

log "step 4/4 report"
"$REPORT_BIN" -in "$RESULTS" -base off -cand pgo | tee "$RUN/pgo-report.txt" || fail "report failed"
