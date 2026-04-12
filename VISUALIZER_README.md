# Dominium Network Visualizer

Visualizador de Rede e Simulador de Ataques para a blockchain Dominium - uma ferramenta completa para monitoramento em tempo real e testes de integridade da rede.

## Funcionalidades

### Painel de Observabilidade (Real-Time Monitoring)

- **Mempool Status**: Lista visual das transações aguardando mineração (ID, Tipo, Timestamp)
- **Blockchain Canvas**: Representação visual da corrente com blocos interligados (placeholder implementado)
- **Account State Table**: Tabela de saldos e posse de NFTs por chave pública
- **Node Stats**: Status de cada nó (Mining/Idle), dificuldade atual e altura do bloco

### Simulador de Transações Aleatórias (Stress Test)

- **Chaos Mint**: Botão que dispara 10 transações de Mint simultaneamente via API
- **Gerador de Carteiras**: Cria pares de chaves ECDSA aleatórios on-the-fly
- **Visualização de Fluxo**: Mostra graficamente API → Kafka → Mempool dos nós (placeholder implementado)

### Modulo de Ataque: Double Spend

- **Race Attack**: Simula criação de duas transações conflitantes usando o mesmo NFT_ID
- **Monitoramento de Rejeição**: Visualiza qual nó aceitou qual transação primeiro
- **Simulação de Fork**: Força um nó a minerar um bloco alternativo e observa a resolução

### Integração Técnica

- **WebSocket/Socket.io**: Atualizações em tempo real da rede (conectado ao backend Go)
- **Modo de Depuração**: Capacidade de pausar mineração para acumular transações (planejado)

## Como Executar

### Opção 1: Sistema Completo com Docker Compose (Recomendado)

```bash
# Gerar configuração
bash generate-env.sh

# Iniciar tudo automaticamente
docker-compose up -d

# Acessar: http://localhost:8080
```

Este comando inicia **tudo automaticamente**:

- Kafka + Zookeeper
- 3 nós mineradores
- API Gateway
- Visualizer completo

### Opção 2: Apenas Visualizer (Sistema já rodando)

```bash
# Build frontend
cd web
npm install
npm run build
cd ..

# Build e executar backend
go build -o visualizer ./cmd/visualizer
./visualizer
```

### Opção 3: Com Docker

```bash
docker-compose up visualizer
```

## Acesso

- **Frontend**: http://localhost:8080
- **WebSocket**: ws://localhost:8080/ws
- **API Endpoints**:
    - `POST /api/chaos-mint` - Dispara 10 transações de mint
    - `POST /api/generate-identity` - Gera nova identidade
    - `POST /api/race-attack` - Simula ataque de corrida

## Arquitetura

```
┌─────────────────┐    WebSocket    ┌─────────────────┐
│   React App     │◄──────────────►│   Go Backend    │
│   (Frontend)    │                │  (Visualizer)   │
│   Ant Design    │                │  Gorilla WS     │
└─────────────────┘                └─────────────────┘
         │                                   │
         │                                   │
         ▼                                   ▼
┌─────────────────┐    Kafka Topics   ┌─────────────────┐
│   Dominium      │◄─────────────────►│   Node Servers  │
│   API Gateway   │                   │   (Miners)      │
│   (Port 8085)   │                   │                 │
└─────────────────┘                   └─────────────────┘
```

## Pré-requisitos

- **Go 1.21+**
- **Node.js 16+**
- **Kafka** (localhost:9092) - Use `docker-compose up kafka` para iniciar
- **Nós Dominium** em execução para dados reais

## Desenvolvimento

### Frontend (React/JavaScript)

```bash
cd web
npm start          # Desenvolvimento (porta 3000)
npm run build      # Produção
npm test           # Testes
```

### Backend (Go)

```bash
go run ./cmd/visualizer
# ou
go build ./cmd/visualizer && ./visualizer
```

## Tecnologias Utilizadas

- **Frontend**: React 18, JavaScript, Ant Design, Socket.io-client
- **Backend**: Go, Gorilla WebSocket, JSON
- **Build**: Docker multi-stage, Webpack
- **Comunicação**: WebSocket para updates em tempo real

## Cenários de Teste

### 1. Monitoramento Básico

1. Inicie nós mineradores:

    ```bash
    go run ./cmd/node -id node-1 -p2p localhost:9092 -mine -difficulty 4
    ```

2. Abra o visualizer: http://localhost:8080

3. Observe updates em tempo real no painel

### 2. Stress Test

1. Clique em "Chaos Mint" no simulador
2. Observe transações sendo enviadas via API
3. Veja resposta do backend no terminal

### 3. Ataque Double Spend

1. Clique em "Race Attack" no módulo de ataques
2. Observe simulação sendo executada
3. Veja logs no backend sobre o ataque

## Estado Atual da Implementação

### Implementado

- Interface React com navegação lateral
- Backend Go com WebSocket server
- Endpoints REST para simulações
- Docker containerization
- Script de inicialização automatizado
- Conexão WebSocket frontend-backend

### Em Desenvolvimento

- Integração real com Kafka para dados de rede
- Canvas interativo da blockchain
- Visualização de fluxo API → Kafka → Mempool
- Coleta de métricas reais dos nós

### Planejado

- Autenticação e controle de acesso
- Modo debug para pausar mineração
- Métricas avançadas de performance
- Export de dados para análise

## Segurança

- **Desenvolvimento**: Private keys nunca são enviadas ao frontend
- **Produção**: Use HTTPS e autenticação adequada
- **Debug**: Modo debug permite pausar mineração para testes controlados

## Contribuição

Para contribuir com o visualizer:

1. Fork o projeto
2. Crie uma branch `visualizer/<feature>`
3. Adicione testes para novas funcionalidades
4. Submeta um Pull Request

## Licença

Este projeto está licenciado sob a MIT License - veja o arquivo LICENSE para detalhes.
