#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require curl

old="$(find_leader)" || fail "no leader found, run scripts/start-all.sh first"
old_term="$(json_field "$(node_status "$old")" term)"
log "current leader: $old term $old_term"

pid="$(port_pid "$(service_port "$old")")"
[ -n "$pid" ] || fail "no process listening for $old"
kill -KILL "$pid"
rm -f "$RUN/$old.pid"
wait_port_down "$(service_port "$old")" 10 || fail "$old still listening after kill"
log "killed $old pid $pid"

survivors=""
for name in $NODES; do
  if [ "$name" != "$old" ] && port_up "$(service_port "$name")"; then
    survivors="$survivors $name"
  fi
done
[ -n "$survivors" ] || fail "no surviving nodes"

new=""
tries=60
start="$(date +%s)"
while [ "$tries" -gt 0 ]; do
  for name in $survivors; do
    if [ "$(node_role "$name")" = "Leader" ]; then new="$name"; fi
  done
  if [ -n "$new" ]; then break; fi
  sleep 1
  tries=$((tries - 1))
done
[ -n "$new" ] || fail "no new leader elected among$survivors within 60s"
new_term="$(json_field "$(node_status "$new")" term)"
log "new leader: $new term $new_term elected in about $(( $(date +%s) - start ))s"

key="failover-$(date +%s)"
value="written-after-$old-died"
writer=""
reader=""
for name in $survivors; do
  if [ "$name" != "$new" ]; then writer="$name"; fi
done
[ -n "$writer" ] || writer="$new"
reader="$new"
kv_put "$writer" "$key" "$value" || fail "write of $key through $writer failed"
log "wrote $key=$value through $writer"
wait_value "$reader" "$key" "$value" 30 || fail "$reader did not return $key"
log "read $key from $reader: $(kv_get "$reader" "$key")"

log "old leader: $old"
log "new leader: $new"
