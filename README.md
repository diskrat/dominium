# Dominium

Blockchain para registro descentralizado de ativos e propriedades.

## O Problema: Fraudes e Inconsistência em Registros

Em muitos sistemas tradicionais, o processo de registro de posse e transferência de patrimônio sofre com:

- **Gasto Duplo (Vendas Múltiplas):** Vender o mesmo ativo para duas pessoas diferentes antes que a transferência seja oficializada.
- **Adulteração:** Registros em papel ou em bancos de dados centralizados podem ser facilmente alterados ou forjados.
- **Falta de Transparência:** Dificuldade em rastrear com precisão o histórico de posse de um ativo ao longo do tempo.

## Nossa Solução (A Ideia)

Um **ledger imutável de ativos** operado por uma rede de múltiplos nós. No _Dominium_, a transferência de posse de qualquer bem (que possua um ID único) é tratada como uma transação nativa da rede.

- **Imutabilidade via Criptografia:** Blocos encadeados por hash tornam impossível adulterar o histórico de um ativo sem refazer todo o esforço de _Proof of Work_ da rede.
- **Validação Distribuída:** Nós mineradores validam as transferências, garantindo que o mesmo ativo não seja transferido duas vezes (prevenção do gasto duplo).
- **Prova de Trabalho (PoW):** Utilizada para assegurar que a rede descentralizada chegue a um consenso seguro sobre quem é o titular verídico de cada ativo registrado.

## Funcionalidades e Requisitos do Sistema

Este projeto implementa uma blockchain distribuída baseada em Proof of Work (PoW) com os seguintes requisitos técnicos:

### 1. Estrutura da Blockchain

- Blocos encadeados por hash, contendo:
    - Lista de transações.
    - Hash do bloco anterior.
    - Nonce para o Proof of Work.

### 2. Proof of Work e Consenso

- O hash do bloco deve iniciar com _N_ zeros (com dificuldade ajustável).
- Adoção da cadeia com maior trabalho acumulado em caso de múltiplos forks.
- Suporte nativo à ocorrência de forks (ex: mineração simultânea).

### 3. Rede e Mineradores

- Execução com múltiplos nós (≥ 3).
- Comunicação e disseminação de blocos/transações na rede.
- Os nós deverão consumir a mempool para montar blocos, rodar o Proof of Work e validar a recepção de novos blocos.

### 4. Mempool e API

- Mempool mantendo as transações pendentes.
- API dedicada para a submissão de novas transações.
- Script para criação de transações aleatórias usando seed, de forma a garantir a reprodutibilidade.

### 5. Experimentação e Visualização

- Visualizar o crescimento da blockchain em tempo real, exibindo forks, os vários blocos gerados e as relações entre eles.
- Executar e simular um ataque de gasto duplo (double spending) verificando seu impacto na rede.

## Arquitetura do Sistema

### Componentes Principais

- **API Gateway**: Interface REST para submissão de transações e consultas de estado
- **Node Servers**: Nós mineradores que mantêm cópias da blockchain e executam Proof of Work
- **Kafka**: Sistema de mensageria para comunicação P2P entre nós
- **Visualizer**: Interface web para monitoramento em tempo real e simulação de ataques

### Topologia da Rede

```
┌─────────────────┐    Kafka Topics    ┌─────────────────┐
│   API Gateway   │◄─────────────────►│   Node Servers  │
│   (Port 8085)   │                   │   (Miners)      │
└─────────────────┘                   └─────────────────┘
         │                                       │
         │                                       │
         ▼                                       ▼
┌─────────────────┐                   ┌─────────────────┐
│   Visualizer    │◄──────────────────┤   WebSocket     │
│   (Port 8080)   │                   │   Updates       │
└─────────────────┘                   └─────────────────┘
```

## Como Executar

### Pré-requisitos

- Go 1.21+
- Node.js 16+ (para o visualizer)
- Docker e Docker Compose

### 1. Iniciar Infraestrutura

```bash
# Inicia Kafka e Zookeeper
docker-compose up -d kafka zookeeper
```

### 2. Executar Nós Mineradores

```bash
# Terminal 1: Nó 1
go run ./cmd/node -id node-1 -p2p localhost:9092 -mine -difficulty 4

# Terminal 2: Nó 2
go run ./cmd/node -id node-2 -p2p localhost:9092 -mine -difficulty 4

# Terminal 3: Nó 3
go run ./cmd/node -id node-3 -p2p localhost:9092 -mine -difficulty 4
```

### 3. Iniciar API Gateway

```bash
# Terminal 4: API Gateway
go run ./cmd/api -port 8085 -id api-gateway -p2p localhost:9092 \
  -admin-key <ADMIN_PRIVATE_KEY> -admin-pub <ADMIN_PUBLIC_KEY>
```

### 4. Executar Visualizer (Opcional)

```bash
# Opção 1: Sistema completo automatizado (recomendado)
./run-full-system.sh

# Opção 2: Manual
cd web && npm install && npm run build && cd ..
go build -o visualizer ./cmd/visualizer
./visualizer
```

## Reprodutibilidade e Testes Determinísticos

### Configuração Inicial (Obrigatório)

```bash
# 1. Gerar chaves admin (sempre as mesmas para reprodutibilidade)
go run generate-admin-keys.go

# Output esperado:
# Admin Private Key: 3081a40201010420... (64 chars)
# Admin Public Key:  3059301306072a86... (128 chars)
```

### Execução Determinística (Para Apresentação)

```bash
# 1. Iniciar infraestrutura
docker-compose up -d kafka zookeeper

# 2. Nó 1 (minerador)
go run ./cmd/node -id node-1 -p2p localhost:9092 -mine -difficulty 4

# 3. Nó 2 (minerador) - em outro terminal
go run ./cmd/node -id node-2 -p2p localhost:9092 -mine -difficulty 4

# 4. Nó 3 (minerador) - em outro terminal
go run ./cmd/node -id node-3 -p2p localhost:9092 -mine -difficulty 4

# 5. API Gateway - em outro terminal
go run ./cmd/api -port 8085 -id api-gateway -p2p localhost:9092 \
  -admin-key $(cat admin_private.key) -admin-pub $(cat admin_public.key)

# 6. Visualizer - em outro terminal
go build -o visualizer ./cmd/visualizer
./visualizer
```

### Teste de Dificuldade (Demonstração de 20% da Nota)

```bash
# Dificuldade baixa (rápido - ~1 segundo)
go run ./cmd/node -id node-test -p2p localhost:9092 -mine -difficulty 8

# Dificuldade alta (lento - ~30+ segundos)
go run ./cmd/node -id node-test -p2p localhost:9092 -mine -difficulty 16
```

### Uso de Seed para Transações Reprodutíveis

O gerador de transações usa uma seed determinística para criar sempre as mesmas transações:

```bash
# No código Go - usar seed fixa para reprodutibilidade
generator := transaction.NewGenerator(12345) // Seed sempre igual = transações sempre iguais

# Para demonstração de ataque double spend:
# 1. Use seed fixa para gerar NFT_ID previsível
# 2. Tente mintar o mesmo NFT_ID duas vezes
# 3. Sistema deve rejeitar a segunda transação
```

## Monitoramento e Testes

### Visualizer Web

- **URL**: http://localhost:8080
- **Funcionalidades**:
    - Monitoramento em tempo real da rede
    - Simulação de transações caóticas
    - Ataques de double spend
    - Visualização da blockchain

### API Examples

Veja `API_EXAMPLES.md` para exemplos completos de uso da API.

### Testes de Ataque

```bash
# Simular ataque de gasto duplo via visualizer
curl -X POST http://localhost:8080/api/race-attack
```

## Estrutura do Projeto

```
dominium/
├── cmd/
│   ├── api/          # API Gateway server
│   ├── node/         # Node server (miner)
│   └── visualizer/   # Web visualizer server
├── internal/
│   ├── api/          # API gateway logic
│   ├── miner/        # Mining and blockchain logic
│   └── network/      # Kafka networking
├── pkg/
│   └── transaction/  # Transaction types and validation
├── web/              # React frontend for visualizer
│   ├── src/
│   └── build/        # Built static files
├── docs/             # Technical documentation
└── docker-compose.yml
```

## Documentação

- `docs/README.md` - Diagramas UML e documentação técnica
- `API_EXAMPLES.md` - Exemplos de uso da API
- `VISUALIZER_README.md` - Documentação específica do visualizer
- `CONTRIBUTING.md` - Guia de contribuição

## Desenvolvimento

### Executar Testes

```bash
go test ./...
```

### Build

```bash
go build ./cmd/api
go build ./cmd/node
go build ./cmd/visualizer
```

### Docker

```bash
# Build all services
docker-compose build

# Run complete system
docker-compose up
```

## Licença

Este projeto está licenciado sob a MIT License - veja o arquivo LICENSE para detalhes.
