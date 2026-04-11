#!/bin/bash

set -e

# 1. Gera a privada P-256
openssl ecparam -name prime256v1 -genkey -noout -out admin_priv.pem

# 2. Gera a pública
openssl ec -in admin_priv.pem -pubout -out admin_pub.pem

# 3. Converte para HEX SEM QUEBRAS DE LINHA (tr -d '\n')
PRIV_HEX=$(openssl ec -in admin_priv.pem -outform DER | xxd -p | tr -d '\n')
PUB_HEX=$(openssl ec -in admin_pub.pem -pubin -outform DER | xxd -p | tr -d '\n')

mkdir -p config
cat > config/runtime.env <<EOF
DIFFICULTY=16
ADMIN_KEY=$PRIV_HEX
ADMIN_PUB=$PUB_HEX
EOF

echo "Arquivo de configuracao gerado em config/runtime.env"