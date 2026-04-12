# Dominium API Gateway - Exemplos de Uso

## Preparação

### 1. Sistema Rodando

```bash
# Iniciar tudo com Docker Compose
bash generate-env.sh
docker-compose up -d

# API estará disponível em: http://localhost:8085
```

### 2. Chaves Admin (já configuradas automaticamente)

As chaves admin são definidas automaticamente no arquivo `.env`:

- `ADMIN_KEY` - Para assinar transações de mint
- `ADMIN_PUB` - Chave pública correspondente

**Não é necessário gerar chaves manualmente!**

## Exemplos de Requisições

### Health Check

```bash
curl -X GET http://localhost:8085/health
```

Resposta esperada:

```json
{
    "status": "ok",
    "synced": true
}
```

### Criar Mint de NFT

POST /transactions

```bash
curl -X POST http://localhost:8085/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mint",
    "recipient": "0x04abcdef...",
    "nft_id": "nft-001"
  }'
```

Resposta esperada:

```json
{
    "tx_id": "hash_da_transacao",
    "status": "Published"
}
```

### Transferir NFT

POST /transactions

```bash
curl -X POST http://localhost:8085/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "type": "transfer",
    "sender": "0x04sender...",
    "recipient": "0x04recipient...",
    "nft_id": "nft-001"
  }'
```

### Status da Rede

GET /network/status

```bash
curl -X GET http://localhost:8085/network/status
```

Resposta esperada:

```json
{
    "network_height": 5,
    "nodes_active": ["node-1", "node-2"],
    "canonical_chain": [
        {
            "hash": "00abc123...",
            "miner_id": "node-1",
            "height": 0,
            "tx_count": 1,
            "timestamp": 1672531200000000000,
            "difficulty": 4
        }
    ],
    "timestamp": 1712250000000000000
}
```

## Visualizer API Endpoints

O visualizer fornece endpoints adicionais para simulação e testes:

### Chaos Mint (Simulação de Stress)

POST /api/chaos-mint

```bash
curl -X POST http://localhost:8080/api/chaos-mint
```

Resposta esperada:

```json
{
    "status": "10 chaos mint transactions submitted"
}
```

### Gerar Carteira Aleatória

POST /api/generate-wallet

```bash
curl -X POST http://localhost:8080/api/generate-wallet
```

Resposta esperada:

```json
{
    "publicKey": "generated-public-key",
    "privateKey": "generated-private-key"
}
```

**⚠️ AVISO**: Em produção, nunca exponha chaves privadas via API.

### Simular Race Attack (Double Spend)

POST /api/race-attack

```bash
curl -X POST http://localhost:8080/api/race-attack
```

Resposta esperada:

```json
{
    "status": "Race attack simulation started"
}
```

## Fluxo de Teste Completo

1. Inicia Docker Compose

```bash
docker-compose up -d
```

2. Inicia um Node

```bash
go run ./cmd/node -id node-1 -p2p localhost:9092 -mine -difficulty 4
```

3. Inicia a API Gateway (em outro terminal)

```bash
# Primeiro, gere/obtenha chaves admin em hex
go run ./cmd/api -port 8085 -id api-gateway -p2p localhost:9092 \
  -admin-key <ADMIN_PRIVKEY> -admin-pub <ADMIN_PUBKEY>
```

4. Check health

```bash
curl -X GET http://localhost:8085/health
```

5. Espere a sincronização completar (deve retornar synced: true)

6. Envie uma transação de mint

```bash
curl -X POST http://localhost:8085/transactions \
  -H "Content-Type: application/json" \
  -d '{
    "type": "mint",
    "recipient": "0x04...",
    "nft_id": "test-nft-001"
  }'
```

7. Consulte o estado da rede após mineração

```bash
curl -X GET http://localhost:8085/network/status
```

8. **Opcional**: Inicie o visualizer para monitoramento

```bash
# Sistema já está rodando com Docker Compose
# Acesse: http://localhost:8080
```

9. **Teste de Ataque**: Use o visualizer para simular ataques

```bash
# Via API do visualizer
curl -X POST http://localhost:8080/api/race-attack
```

## Códigos HTTP Esperados

- **API Gateway (Port 8085)**:
    - 200: GET bem-sucedido
    - 201: POST bem-sucedido
    - 400: Erro de validação (campo obrigatório faltando)
    - 500: Erro interno do servidor
    - 503: Rede ainda não sincronizada

- **Visualizer (Port 8080)**:
    - 200: Simulação executada com sucesso
    - 500: Erro interno do servidor

## Notas

- O admin é identificado pela chave pública em hex
- Transações de transfer requerem que o sender possua o NFT
- Transações de mint precisam que o NFT ainda não esteja mintado
- O gateway rejeita transações enquanto está sincronizando (synced = false)
- O visualizer é uma ferramenta de desenvolvimento/teste - não use em produção
