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
- API dedicada para a submissão de novas transações de mint e transfer.
- Geração de transações e carteiras suportada pelo gateway e pelo visualizer.

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

### Método Recomendado: Docker Compose (Tudo Automático)

```bash
# 1. Gerar configuração
bash generate-env.sh
# O arquivo .env gerado inclui ADMIN_KEY e ADMIN_PUB usados pelo API Gateway.

# 2. Iniciar tudo (7 serviços simultaneamente)
docker-compose up -d

# 3. Acessar aplicações
# - API Gateway: http://localhost:8085
# - Visualizer: http://localhost:8080
```

**Serviços iniciados automaticamente:**

- ✅ Zookeeper (coordenação)
- ✅ Kafka (mensageria P2P)
- ✅ 3 Nós mineradores (node-1, node-2, node-3)
- ✅ API Gateway (porta 8085)
- ✅ Visualizer (porta 8080)

### Método Manual (Para Desenvolvimento)

#### Pré-requisitos

- Go 1.21+
- Node.js 16+ (para o visualizer)
- Docker e Docker Compose

#### 1. Iniciar Infraestrutura

```bash
docker-compose up -d kafka zookeeper
```

#### 2. Executar Nós Mineradores

```bash
# Terminal 1: Nó 1
go run ./cmd/node -id node-1 -p2p localhost:9092 -mine -difficulty 4

# Terminal 2: Nó 2
go run ./cmd/node -id node-2 -p2p localhost:9092 -mine -difficulty 4

# Terminal 3: Nó 3
go run ./cmd/node -id node-3 -p2p localhost:9092 -mine -difficulty 4
```

#### 3. Iniciar API Gateway

```bash
# Terminal 4: API Gateway
go run ./cmd/api -port 8085 -id api-gateway -p2p localhost:9092 \
  -admin-key $(grep ADMIN_KEY .env | cut -d'=' -f2) \
  -admin-pub $(grep ADMIN_PUB .env | cut -d'=' -f2)
```

#### 4. Executar Visualizer

```bash
cd web && npm install && npm run build && cd ..
go build -o visualizer ./cmd/visualizer
./visualizer
```

## Reprodutibilidade e Testes Determinísticos

### Configuração Inicial (Automática)

```bash
# Gerar arquivo .env com chaves padrão (sempre as mesmas)
bash generate-env.sh
```

### Execução Determinística (Para Apresentação)

```bash
# Tudo em um comando
docker-compose up -d
```

> Para execução manual, use os passos descritos em “Método Manual” acima.

### Teste de Dificuldade (Demonstração de 20% da Nota)

```bash
# Com Docker Compose - alterar dificuldade no .env
echo "DIFFICULTY=8" >> .env
docker-compose up -d

# Ou manual:
# Dificuldade baixa (rápido - ~1 segundo)
go run ./cmd/node -id node-test -p2p localhost:9092 -mine -difficulty 8

# Dificuldade alta (lento - ~30+ segundos)
go run ./cmd/node -id node-test -p2p localhost:9092 -mine -difficulty 16
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
