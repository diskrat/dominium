#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="/tmp/dominium"
mkdir -p "$LOG_DIR"

RABBIT_NAME="dominium-rabbitmq"
AMQP_URL="amqp://guest:guest@localhost:5672/"

process_matches() {
  local pid="$1"
  local expected="$2"

  [[ "$pid" =~ ^[0-9]+$ ]] || return 1
  [[ -d "/proc/$pid" ]] || return 1

  local cmdline
  cmdline="$(tr '\0' ' ' < "/proc/$pid/cmdline" 2>/dev/null || true)"
  [[ -n "$cmdline" ]] || return 1
  [[ "$cmdline" == *"$expected"* ]]
}

wait_rabbitmq_ready() {
  for _ in {1..60}; do
    if docker exec "$RABBIT_NAME" rabbitmq-diagnostics -q ping >/dev/null 2>&1; then
      echo "[ok] RabbitMQ pronto para conexoes AMQP"
      return 0
    fi
    sleep 1
  done

  echo "[erro] RabbitMQ nao ficou pronto a tempo"
  return 1
}

ensure_rabbitmq() {
  if docker ps --format '{{.Names}}' | grep -qx "$RABBIT_NAME"; then
    echo "[ok] RabbitMQ já está rodando"
  elif docker ps -a --format '{{.Names}}' | grep -qx "$RABBIT_NAME"; then
    docker start "$RABBIT_NAME" >/dev/null
    echo "[ok] RabbitMQ iniciado"
  else
    docker run -d --name "$RABBIT_NAME" -p 5672:5672 -p 15672:15672 rabbitmq:3-management >/dev/null
    echo "[ok] RabbitMQ criado e iniciado"
  fi

  wait_rabbitmq_ready
}

start_node() {
  local node_id="$1"
  local port="$2"
  local pid_file="$LOG_DIR/node-${node_id}.pid"
  local log_file="$LOG_DIR/node-${node_id}.log"
  local expected="cmd/node -id $node_id -api :$port"

  if [[ -f "$pid_file" ]]; then
    local pid
    pid="$(cat "$pid_file" 2>/dev/null || true)"
    if process_matches "$pid" "$expected"; then
      echo "[ok] node-${node_id} já está rodando (PID $pid)"
      return
    fi
    rm -f "$pid_file"
  fi

  nohup bash -lc "cd '$ROOT_DIR' && go run ./cmd/node -id '$node_id' -api ':$port' -amqp '$AMQP_URL' -difficulty 2" >"$log_file" 2>&1 &
  echo $! > "$pid_file"
  echo "[ok] node-${node_id} iniciado em :$port (PID $(cat "$pid_file"))"
}

start_frontend() {
  local pid_file="$LOG_DIR/frontend.pid"
  local log_file="$LOG_DIR/frontend.log"

  if [[ -f "$pid_file" ]] && [[ -d "/proc/$(cat "$pid_file")" ]]; then
    echo "[ok] frontend já está rodando (PID $(cat "$pid_file"))"
    return
  fi

  nohup python3 -m http.server 8090 --directory "$ROOT_DIR/front-p2p/frontend" >"$log_file" 2>&1 &
  echo $! > "$pid_file"
  echo "[ok] frontend iniciado em :8090 (PID $(cat "$pid_file"))"
}

wait_health() {
  local node_id="$1"
  local port="$2"
  local log_file="$LOG_DIR/node-${node_id}.log"

  for _ in {1..30}; do
    if curl -fsS "http://127.0.0.1:${port}/health" >/dev/null 2>&1; then
      echo "[ok] node na porta :$port saudável"
      return 0
    fi
    sleep 1
  done

  echo "[erro] node na porta :$port não respondeu /health"
  if [[ -f "$log_file" ]]; then
    echo "[debug] ultimas linhas de $log_file"
    tail -n 30 "$log_file" || true
  fi
  return 1
}

echo "===> Subindo Dominium (RabbitMQ + 3 nós + frontend)"
ensure_rabbitmq
start_node "node-1" 8080
start_node "node-2" 8081
start_node "node-3" 8082
start_frontend

wait_health "node-1" 8080
wait_health "node-2" 8081
wait_health "node-3" 8082

echo
echo "✅ Projeto no ar:"
echo "- Frontend:   http://localhost:8090"
echo "- Node 1:     http://localhost:8080/health"
echo "- Node 2:     http://localhost:8081/health"
echo "- Node 3:     http://localhost:8082/health"
echo "- RabbitMQ:   http://localhost:15672 (guest/guest)"
echo
echo "Logs em: $LOG_DIR"
