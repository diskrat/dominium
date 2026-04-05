#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
LOG_DIR="/tmp/dominium"
mkdir -p "$LOG_DIR"

RABBIT_NAME="dominium-rabbitmq"
AMQP_URL="amqp://guest:guest@localhost:5672/"

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
}

start_node() {
  local node_id="$1"
  local port="$2"
  local pid_file="$LOG_DIR/node-${node_id}.pid"
  local log_file="$LOG_DIR/node-${node_id}.log"

  if [[ -f "$pid_file" ]] && [[ -d "/proc/$(cat "$pid_file")" ]]; then
    echo "[ok] node-${node_id} já está rodando (PID $(cat "$pid_file"))"
    return
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
  local port="$1"
  for _ in {1..30}; do
    if curl -fsS "http://127.0.0.1:${port}/health" >/dev/null 2>&1; then
      echo "[ok] node na porta :$port saudável"
      return 0
    fi
    sleep 1
  done
  echo "[erro] node na porta :$port não respondeu /health"
  return 1
}

echo "===> Subindo Dominium (RabbitMQ + 3 nós + frontend)"
ensure_rabbitmq
start_node "node-1" 8080
start_node "node-2" 8081
start_node "node-3" 8082
start_frontend

wait_health 8080
wait_health 8081
wait_health 8082

echo
echo "✅ Projeto no ar:"
echo "- Frontend:   http://localhost:8090"
echo "- Node 1:     http://localhost:8080/health"
echo "- Node 2:     http://localhost:8081/health"
echo "- Node 3:     http://localhost:8082/health"
echo "- RabbitMQ:   http://localhost:15672 (guest/guest)"
echo
echo "Logs em: $LOG_DIR"
