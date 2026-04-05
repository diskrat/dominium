#!/usr/bin/env bash
set -euo pipefail

LOG_DIR="/tmp/dominium"
RABBIT_NAME="dominium-rabbitmq"

safe_kill() {
  local pid="$1"
  local expected="$2"

  [[ "$pid" =~ ^[0-9]+$ ]] || return 1
  [[ -d "/proc/$pid" ]] || return 1

  local cmdline
  cmdline="$(tr '\0' ' ' < "/proc/$pid/cmdline" 2>/dev/null || true)"
  if [[ -z "$cmdline" ]] || [[ "$cmdline" != *"$expected"* ]]; then
    return 1
  fi

  kill "$pid" 2>/dev/null || true
  sleep 1
  if [[ -d "/proc/$pid" ]]; then
    kill -9 "$pid" 2>/dev/null || true
  fi
  return 0
}

stop_pid_file() {
  local label="$1"
  local pid_file="$2"
  local expected="$3"

  if [[ ! -f "$pid_file" ]]; then
    echo "[ok] $label: sem pid file"
    return
  fi

  local pid
  pid="$(cat "$pid_file" 2>/dev/null || true)"

  if safe_kill "$pid" "$expected"; then
    echo "[ok] $label parado (PID $pid)"
  else
    echo "[ok] $label: processo já parado ou não corresponde ao esperado"
  fi

  rm -f "$pid_file"
}

echo "===> Parando Dominium (nós + frontend + RabbitMQ)"

stop_pid_file "node-1" "$LOG_DIR/node-node-1.pid" "cmd/node"
stop_pid_file "node-2" "$LOG_DIR/node-node-2.pid" "cmd/node"
stop_pid_file "node-3" "$LOG_DIR/node-node-3.pid" "cmd/node"
stop_pid_file "frontend" "$LOG_DIR/frontend.pid" "http.server"

if docker ps --format '{{.Names}}' | grep -qx "$RABBIT_NAME"; then
  docker stop "$RABBIT_NAME" >/dev/null
  echo "[ok] RabbitMQ parado"
else
  echo "[ok] RabbitMQ já estava parado"
fi

echo "✅ Ambiente parado"
