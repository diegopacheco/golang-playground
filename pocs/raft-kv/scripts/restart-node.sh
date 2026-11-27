#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

require curl

target="${1:-}"
if [ -z "$target" ]; then
  for name in $NODES; do
    if ! port_up "$(service_port "$name")"; then target="$name"; break; fi
  done
fi
[ -n "$target" ] || fail "every node is already up, pass a node name to restart it"
case " $NODES " in
  *" $target "*) ;;
  *) fail "unknown node $target, use one of: $NODES" ;;
esac
if port_up "$(service_port "$target")"; then
  fail "$target is already running"
fi
[ -f "$RUN/data/$target/raft.db" ] || fail "$target has no data, run scripts/start-all.sh instead"

leader="$(find_leader)" || fail "no leader among running nodes"
key="catchup-$(date +%s)"
value="written-while-$target-was-down"
kv_put "$leader" "$key" "$value" || fail "write of $key through $leader failed"
log "wrote $key=$value through $leader while $target is down"

start_node "$target"
wait_value "$target" "$key" "$value" 60 || fail "$target did not catch up on $key, see $LOGS/$target.log"
log "$target caught up: $(kv_get "$target" "$key")"
s="$(node_status "$target")"
log "$target role $(json_field "$s" role) leader $(json_field "$s" leader_id) term $(json_field "$s" term) applied_index $(json_field "$s" applied_index)"
