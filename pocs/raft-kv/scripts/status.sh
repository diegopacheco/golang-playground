#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

down=0
leaders=0
printf "%-6s %-6s %-6s %-6s %-7s %-10s %-7s %-5s %s\n" NODE HTTP RAFT STATE PID ROLE LEADER TERM LAST_INDEX
for name in $NODES; do
  port="$(service_port "$name")"
  rport="$(raft_port "$name")"
  pid="$(port_pid "$port")"
  s="$(node_status "$name")"
  if [ -n "$pid" ] && [ -n "$s" ]; then
    role="$(json_field "$s" role)"
    if [ "$role" = "Leader" ]; then leaders=$((leaders + 1)); fi
    printf "%-6s %-6s %-6s %-6s %-7s %-10s %-7s %-5s %s\n" "$name" "$port" "$rport" UP "$pid" "$role" \
      "$(json_field "$s" leader_id)" "$(json_field "$s" term)" "$(json_field "$s" last_index)"
  else
    printf "%-6s %-6s %-6s %-6s %-7s %-10s %-7s %-5s %s\n" "$name" "$port" "$rport" DOWN - - - - -
    down=$((down + 1))
  fi
done
log "leaders: $leaders"

exit "$down"
