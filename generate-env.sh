#!/bin/bash

# Script para gerar o arquivo .env com chaves admin para o Docker Compose
# Uso: ./generate-env.sh

echo "Gerando arquivo .env com chaves admin..."

# Gerar chaves
KEYS_OUTPUT=$(go run generate-admin-keys.go 2>/dev/null)

# Extrair private key
ADMIN_PRIV=$(echo "$KEYS_OUTPUT" | grep -A 1 "Private Key (para -admin-key):" | tail -1)

# Extrair public key
ADMIN_PUB=$(echo "$KEYS_OUTPUT" | grep -A 1 "Public Key (para -admin-pub):" | tail -1)

# Criar arquivo .env
cat > .env << EOF
# Dominium Blockchain - Environment Configuration
# Generated: $(date)

# Admin Keys for API Gateway
ADMIN_PRIVATE_KEY=$ADMIN_PRIV
ADMIN_PUBLIC_KEY=$ADMIN_PUB

# Kafka Configuration
KAFKA_BROKERS=kafka:9092

# Network Configuration
NETWORK_DIFFICULTY=4
EOF

echo "Arquivo .env criado com sucesso!"
echo ""
echo "Chaves geradas:"
echo "Private Key: $ADMIN_PRIV"
echo "Public Key:  $ADMIN_PUB"
echo ""
echo "Agora você pode executar: docker-compose up -d"
