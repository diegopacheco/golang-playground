#!/usr/bin/env bash
set -euo pipefail

ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
SCRIPTS="$ROOT/scripts"
RUN="$ROOT/.run"
LOGS="$RUN/logs"
BIN="$ROOT/bin/raftkv"
NODES="node1 node2 node3"
cd "$ROOT"

mkdir -p "$RUN" "$LOGS"

SERVICES=""
if [ -f "$SCRIPTS/ports.env" ]; then
  SERVICES="$(grep -v '^[[:space:]]*#' "$SCRIPTS/ports.env" | grep '=' || true)"
fi

service_names() {
  printf "%s\n" "$SERVICES" | sed '/^$/d' | cut -d= -f1
}

service_port() {
  printf "%s\n" "$SERVICES" | sed '/^$/d' | awk -F= -v n="$1" '$1==n{print $2; exit}'
}

service_url() {
  printf "http://localhost:%s\n" "$(service_port "$1")"
}

raft_port() {
  service_port "raft${1#node}"
}

port_pid() {
  lsof -ti "tcp:$1" -sTCP:LISTEN 2>/dev/null | head -1 || true
}

port_up() {
  [ -n "$(port_pid "$1")" ]
}

wait_port_up() {
  local tries
  tries="${2:-60}"
  while [ "$tries" -gt 0 ]; do
    if port_up "$1"; then return 0; fi
    sleep 1
    tries=$((tries - 1))
  done
  return 1
}

wait_port_down() {
  local tries
  tries="${2:-30}"
  while [ "$tries" -gt 0 ]; do
    if ! port_up "$1"; then return 0; fi
    sleep 1
    tries=$((tries - 1))
  done
  return 1
}

start_bg() {
  local name dir
  name="$1"
  dir="$2"
  shift 2
  if [ -f "$RUN/$name.pid" ] && kill -0 "$(cat "$RUN/$name.pid")" 2>/dev/null; then
    log "$name already running"
    return 0
  fi
  ( cd "$dir" && exec "$@" >>"$LOGS/$name.log" 2>&1 ) &
  echo $! >"$RUN/$name.pid"
  log "$name started pid $!"
}

stop_bg() {
  local name pid port
  name="$1"
  if [ -f "$RUN/$name.pid" ]; then
    pid="$(cat "$RUN/$name.pid")"
    if kill -0 "$pid" 2>/dev/null; then
      kill -TERM "$pid" 2>/dev/null || true
    fi
    rm -f "$RUN/$name.pid"
  fi
  for port in "$(service_port "$name")" "$(raft_port "$name")"; do
    wait_port_down "$port" 10 || true
    pid="$(port_pid "$port")"
    if [ -n "$pid" ]; then kill -KILL "$pid" 2>/dev/null || true; fi
  done
  log "$name stopped"
}

build_binary() {
  require go
  mkdir -p "$ROOT/bin"
  go build -o "$BIN" ./cmd/raftkv || fail "go build failed"
}

start_node() {
  local name extra
  name="$1"
  shift
  [ -x "$BIN" ] || build_binary
  start_bg "$name" "$ROOT" "$BIN" -id "$name" \
    -http "localhost:$(service_port "$name")" \
    -raft "localhost:$(raft_port "$name")" \
    -data ".run/data/$name" "$@"
  wait_port_up "$(service_port "$name")" 60 || fail "$name did not open port $(service_port "$name"), see $LOGS/$name.log"
}

node_status() {
  curl -sf --max-time 2 "$(service_url "$1")/status" 2>/dev/null || true
}

json_field() {
  printf "%s" "$1" | sed -n "s/.*\"$2\":\"\{0,1\}\([^,\"}]*\).*/\1/p"
}

node_role() {
  local s
  s="$(node_status "$1")"
  if [ -z "$s" ]; then
    printf "%s\n" "-"
    return 0
  fi
  json_field "$s" role
}

find_leader() {
  local name
  for name in $NODES; do
    if [ "$(node_role "$name")" = "Leader" ]; then
      printf "%s\n" "$name"
      return 0
    fi
  done
  return 1
}

wait_leader() {
  local tries leader
  tries="${1:-60}"
  while [ "$tries" -gt 0 ]; do
    if leader="$(find_leader)"; then
      printf "%s\n" "$leader"
      return 0
    fi
    sleep 1
    tries=$((tries - 1))
  done
  return 1
}

wait_knows_leader() {
  local tries s
  tries="${2:-60}"
  while [ "$tries" -gt 0 ]; do
    s="$(node_status "$1")"
    if [ -n "$s" ] && [ -n "$(json_field "$s" leader_http_addr)" ]; then return 0; fi
    sleep 1
    tries=$((tries - 1))
  done
  return 1
}

kv_put() {
  curl -sf -L --max-time 10 -X PUT -H "X-Correlation-Id: script-$2" --data-binary "$3" "$(service_url "$1")/kv/$2" >/dev/null
}

kv_get() {
  curl -sf --max-time 2 "$(service_url "$1")/kv/$2" 2>/dev/null || true
}

wait_value() {
  local tries
  tries="${4:-30}"
  while [ "$tries" -gt 0 ]; do
    if [ "$(json_field "$(kv_get "$1" "$2")" value)" = "$3" ]; then return 0; fi
    sleep 1
    tries=$((tries - 1))
  done
  return 1
}

require() {
  command -v "$1" >/dev/null 2>&1 || fail "$1 is required but not installed"
}

log() {
  printf "%s\n" "$*"
}

fail() {
  printf "ERROR: %s\n" "$*" >&2
  exit 1
}
