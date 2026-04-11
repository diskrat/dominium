#!/usr/bin/env bash

set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DEFAULT_ENV="$ROOT_DIR/config/default.env"
RUNTIME_ENV="$ROOT_DIR/config/runtime.env"

PRIVATE_PEM="$(openssl ecparam -name prime256v1 -genkey -noout)"
PRIV_HEX="$(printf '%s\n' "$PRIVATE_PEM" | openssl ec -outform DER | xxd -p | tr -d '\n')"
PUB_HEX="$(printf '%s\n' "$PRIVATE_PEM" | openssl ec -pubout -outform DER | xxd -p | tr -d '\n')"

mkdir -p "$ROOT_DIR/config"

if [[ ! -f "$DEFAULT_ENV" ]]; then
	cat > "$DEFAULT_ENV" <<EOF
DIFFICULTY=16
EOF
fi

{
	grep -vE '^(ADMIN_KEY|ADMIN_PUB)=' "$DEFAULT_ENV"
	printf 'ADMIN_KEY=%s\n' "$PRIV_HEX"
	printf 'ADMIN_PUB=%s\n' "$PUB_HEX"
} > "$RUNTIME_ENV"

echo "Arquivo gerado em config/runtime.env a partir de config/default.env"