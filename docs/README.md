# Documentação - Diagramas UML do Sistema Dominium

Esta pasta contém diagramas UML em formato Mermaid que documentam os principais processos e fluxos do sistema blockchain Dominium.

## Diagramas Disponíveis

### 1. `transaction-submission-sequence.md`

**Diagrama de Sequência** - Mostra o fluxo completo de submissão de uma transação:

- Cliente → API Gateway → NodeServer → Mempool
- Validações de assinatura e regras de consenso

### 2. `mining-process-activity.md`

**Diagrama de Atividades** - Processo de mineração de blocos:

- Estados: Idle → Mining → Proof-of-Work → Validação
- Loop de tentativa até encontrar hash válido

### 3. `node-sync-sequence.md`

**Diagrama de Sequência** - Sincronização de novos nós na rede:

- Bootstrap do estado atual
- Transferência de blocos históricos
- Inscrição em tópicos Kafka

### 4. `block-validation-activity.md`

**Diagrama de Atividades** - Validação completa de blocos recebidos:

- Verificação PoW, Merkle root, transações
- Detecção e resolução de forks
- Regras de consenso Nakamoto

### 5. `system-classes.md`

**Diagrama de Classes** - Estrutura principal do sistema:

- Relacionamentos entre NodeServer, Blockchain, Transactions
- Interfaces de transporte Kafka
- Estado de contas e validações

### 6. `fork-resolution-sequence.md`

**Diagrama de Sequência** - Processo detalhado de resolução de forks:

- Detecção de conflito
- Comparação de cadeias
- Reorganização e atualização de estado

### 7. `blockchain-state.md`

**Diagrama de Estados** - Estados possíveis da blockchain:

- Genesis → Growing → Stable
- Detecção e resolução de forks

### 8. `architecture-blueprint.md`

**Blue Print da Arquitetura** - Documentação técnica completa:

- Estrutura detalhada de Blocos (Header/Body)
- Campos técnicos das Transações
- Topografia da rede Hub-and-Spoke com Kafka
- Regras de consenso e validações
- Métricas técnicas do sistema

## Visualizer - Componente Adicional

O sistema Dominium inclui um **Visualizer Web** para monitoramento e testes:

### Funcionalidades do Visualizer

- **Painel de Observabilidade**: Monitoramento em tempo real da rede
- **Simulador de Transações**: Geração de transações caóticas para stress testing
- **Módulo de Ataques**: Simulação de double-spend e race attacks
- **WebSocket Integration**: Updates em tempo real via WebSocket

### Arquitetura do Visualizer

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
└─────────────────┘                   └─────────────────┘
```

### Como Executar o Visualizer

```bash
# Docker Compose (recomendado)
bash generate-env.sh
docker-compose up -d

# Ou manualmente
cd web && npm install && npm run build && cd ..
go build -o visualizer ./cmd/visualizer
./visualizer
```

**Acesso**: http://localhost:8080

## Como Visualizar

Para visualizar os diagramas, você pode:

1. Abrir os arquivos `.md` em um editor que suporte Mermaid (VS Code com extensão, GitHub, etc.)
2. Usar ferramentas online como [Mermaid Live Editor](https://mermaid.live/)
3. Copiar o código Mermaid para qualquer renderizador compatível

## Processo de Desenvolvimento

Estes diagramas foram criados para documentar a implementação completa do Consenso Nakamoto no Dominium, incluindo:

- Validação de autoridade admin para mints
- Unicidade de NFTs
- Resolução de forks via reorganização
- Sincronização P2P via Kafka
- API REST para submissão de transações
- **Visualizer Web** para monitoramento e testes de ataques

## Métricas e Monitoramento

O visualizer permite monitorar:

- **Throughput**: Transações por segundo
- **Latência**: Tempo de bloco e confirmação
- **Conflitos**: Taxa de rejeição de transações
- **Forks**: Detecção e resolução de bifurcações
- **Estado**: Distribuição de NFTs por contas

## Cenários de Teste

### Ataques Double-Spend

- Race Attack: Duas transações conflitantes simultâneas
- Fork Simulation: Mineração de blocos alternativos
- Rejeição Monitoring: Observação de resolução de conflitos

### Stress Testing

- Chaos Mint: Múltiplas transações simultâneas
- Network Partition: Simulação de desconexões
- High Load: Testes de performance sob carga
