#!/bin/bash

# Script para criar o arquivo .env com configuração padrão
# Uso: ./generate-env.sh

if [ -f .env ]; then
    echo "⚠️  Arquivo .env já existe!"
    read -p "Deseja sobrescrever? (s/n) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Ss]$ ]]; then
        echo "Operação cancelada."
        exit 1
    fi
fi

# Criar arquivo .env com chaves padrão
cat > .env << 'EOF'
# Dominium Blockchain - Environment Configuration

# Admin Keys for API Gateway (used to sign mint transactions)
ADMIN_PRIVATE_KEY=308187020100301306072a8648ce3d020106082a8648ce3d030107046d306b0201010420e2eadd0725eb5f2d546bd6ed9b2789adf69151a3487c461ed362b49fffeb5c0ea14403420004380a7e48b4bf3c744a2672e2807b0a565c7ab31ba5293f33bc6e039c8310010f9f5f7f701a21849856c0697a3e98f27ce107455587aae13516fcff06b1cf5ce6
ADMIN_PUBLIC_KEY=3059301306072a8648ce3d020106082a8648ce3d03010703420004380a7e48b4bf3c744a2672e2807b0a565c7ab31ba5293f33bc6e039c8310010f9f5f7f701a21849856c0697a3e98f27ce107455587aae13516fcff06b1cf5ce6

# Kafka Configuration
KAFKA_BROKERS=kafka:9092

# Network Configuration
NETWORK_DIFFICULTY=4
EOF

echo "✅ Arquivo .env criado com sucesso!"
echo ""
echo "Configuração:"
grep -E "^[A-Z_]+=" .env | sed 's/=.*/=***/'
echo ""
echo "Para mudar as chaves, edite o arquivo: nano .env"
echo "Agora você pode executar: docker-compose up -d"
