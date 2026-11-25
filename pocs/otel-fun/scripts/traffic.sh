#!/usr/bin/env bash
set -euo pipefail
source "$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)/common.sh"

count="${1:-20}"
case "$count" in
  ''|*[!0-9]*) fail "usage: traffic.sh [number_of_requests]" ;;
esac

require curl
port_up "$(service_port service)" || fail "service is not running, run ./scripts/start-all.sh first"
port_up "$(service_port jaeger)" || fail "jaeger is not running, run ./scripts/start-all.sh first"

url="$(service_url service)"
items="book pen laptop mug lamp"
run_id="$(date +%s)"
started="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
ok=0
rejected=0
i=1
while [ "$i" -le "$count" ]; do
  item="$(printf "%s\n" $items | sed -n "$(( (i % 5) + 1 ))p")"
  quantity=$(( (i * 3) % 14 + 1 ))
  price=$(( (i * 731) % 5000 + 99 ))
  if [ $((i % 5)) -eq 0 ]; then
    body="{\"item\":\"\",\"quantity\":$quantity,\"unit_price_cents\":$price}"
  else
    body="{\"item\":\"$item\",\"quantity\":$quantity,\"unit_price_cents\":$price}"
  fi
  cid="traffic-$run_id-$i"
  status="$(curl -s -o "$RUN/last-response.json" -w '%{http_code}' -X POST "$url" \
    -H 'Content-Type: application/json' -H "X-Correlation-Id: $cid" -d "$body")"
  case "$status" in
    201) ok=$((ok + 1)) ;;
    422) rejected=$((rejected + 1)) ;;
    *) fail "unexpected status $status for $cid: $(cat "$RUN/last-response.json")" ;;
  esac
  printf "%-26s %s %s\n" "$cid" "$status" "$(cat "$RUN/last-response.json")"
  i=$((i + 1))
done

log "sent $count orders: $ok created, $rejected rejected"

tries=30
traces=0
while [ "$tries" -gt 0 ]; do
  api="$(service_url jaeger)/api/v3/traces?query.service_name=$SERVICE_NAME&query.start_time_min=$started&query.start_time_max=$(date -u +%Y-%m-%dT%H:%M:%SZ)&query.num_traces=1000"
  traces="$( (curl -s "$api" || true) | (grep -o "traffic-$run_id-[0-9]*\"" || true) | sort -u | wc -l | tr -d ' ')"
  if [ "$traces" -ge "$count" ]; then break; fi
  sleep 1
  tries=$((tries - 1))
done
[ "$traces" -ge "$count" ] || fail "jaeger has $traces of $count traces for this run, see .run/logs/service.log"

log "jaeger received all $traces traces of this run, tagged correlation.id=traffic-$run_id-N"
log "open $(service_url jaeger)/search?service=$SERVICE_NAME"
