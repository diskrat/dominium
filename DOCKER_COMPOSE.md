# Como Executar Dominium com Docker Compose

## Inicie Tudo em Um Comando!

### Pré-requisitos

- Docker
- Docker Compose v2+

### Opção 1: Execução Rápida

```bash
# Gere o arquivo .env (usa chaves padrão)
bash generate-env.sh

# Inicie todos os serviços
docker-compose up -d
```

**Serviços ativos:**

- API Gateway: http://localhost:8085
- Visualizer: http://localhost:8080
- 3 Nós Mineradores: node-1, node-2, node-3
- Kafka: localhost:9092

### Monitorar Logs

```bash
# Todos os logs
docker-compose logs -f

# Logs específicos
docker-compose logs -f node-1
docker-compose logs -f api-gateway
docker-compose logs -f visualizer

# Logs em arquivo
docker-compose logs > logs/all.log
```

### Parar o Sistema

```bash
# Parar todos os serviços (mantém dados)
docker-compose stop

# Remover tudo (limpa containers)
docker-compose down

# Limpar tudo incluindo volumes
docker-compose down -v
```

### Testar a Rede

```bash
# Health check do API Gateway
curl http://localhost:8085/health

# Status da rede
curl http://localhost:8085/network/status

# Abrir visualizer
open http://localhost:8080
```

### Modular o Sistema

Você pode comentar/descomentair serviços no `docker-compose.yml`:

```yaml
# Desativar um nó específico:
# node-2:
#     ...

# Desativar visualizer:
# visualizer:
#     ...
```

### Problemas Comuns

**Porta 8085 ou 8080 já em uso:**

```bash
# Mudar portas no docker-compose.yml
api-gateway:
    ports:
        - "9085:8085"  # Mude 9085
```

**Regenerar tudo do zero:**

```bash
docker-compose down -v
rm .env
bash generate-env.sh
docker-compose up -d
```

**Ver recursos usados:**

```bash
docker-compose stats
```

### Variáveis de Ambiente (`.env`)

```
ADMIN_KEY=...             # Chave privada do admin (para mint de NFTs)
ADMIN_PUB=...             # Chave pública do admin
KAFKA_BROKERS=kafka:9092  # Brokers Kafka (não mude)
DIFFICULTY=4              # Dificuldade PoW (4, 8, 16...)
```

> O `docker-compose.yml` e o `cmd/api/Dockerfile` usam `ADMIN_KEY` e `ADMIN_PUB` para iniciar o API Gateway com a identidade administrativa.

---

**Pronto!** O sistema Dominium está rodando com todos os componentes no Docker Compose.
