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

# Copiar configuração padrão
cp .env.example .env

echo "✅ Arquivo .env criado com sucesso!"
echo ""
echo "Configuração:"
grep -E "^[A-Z_]+=" .env | sed 's/=.*/=***/'
echo ""
echo "Para mudar as chaves, edite o arquivo: nano .env"
echo "Agora você pode executar: docker-compose up -d"
