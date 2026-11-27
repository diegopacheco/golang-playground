#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

log "starting"

require curl
require lsof
build_binary
log "built bin/raftkv"

fresh=1
if [ -f "$RUN/data/node1/raft.db" ]; then fresh=0; fi

start_node node1 -bootstrap
if [ "$fresh" -eq 1 ]; then
  wait_leader 60 >/dev/null || fail "node1 did not become leader, see $LOGS/node1.log"
  wait_knows_leader node1 60 || fail "node1 did not publish its http address"
fi

for name in node2 node3; do
  start_node "$name" -join "localhost:$(service_port node1)"
done

leader="$(wait_leader 60)" || fail "no leader elected, see $LOGS"
for name in $NODES; do
  wait_knows_leader "$name" 60 || fail "$name did not join the cluster, see $LOGS/$name.log"
  log "$name up on $(service_url "$name")"
done
log "leader is $leader"

"$SCRIPTS/status.sh"

log "links"
for name in $NODES; do
  printf "%-8s %s\n" "$name" "$(service_url "$name")"
done
