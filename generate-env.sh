#!/bin/bash

# 1. Gera a privada P-256
openssl ecparam -name prime256v1 -genkey -noout -out admin_priv.pem

# 2. Gera a pública
openssl ec -in admin_priv.pem -pubout -out admin_pub.pem

# 3. Converte para HEX SEM QUEBRAS DE LINHA (tr -d '\n')
PRIV_HEX=$(openssl ec -in admin_priv.pem -outform DER | xxd -p | tr -d '\n')
PUB_HEX=$(openssl ec -in admin_pub.pem -pubin -outform DER | xxd -p | tr -d '\n')

echo "--- COPIE PARA O SEU .ENV ---"
echo "ADMIN_PRIVATE_KEY=$PRIV_HEX"
echo "ADMIN_PUBLIC_KEY=$PUB_HEX"