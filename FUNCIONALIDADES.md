# Quadro de Funcionalidades - Dominium Blockchain

## Visão Geral do Sistema

**Dominium** é uma blockchain distribuída para registro descentralizado de ativos e propriedades, implementada em Go com mecanismo de consenso Proof of Work (PoW). O sistema previne fraudes como gasto duplo e adulteração de registros através de criptografia, validação distribuída e ledger imutável.

---

## 📦 1. Core Blockchain

| Módulo | Funcionalidade | Descrição |
|--------|----------------|-----------|
| **blockchain.go** | Gerenciamento da Cadeia | Mantém sequência linear de blocos com controle de acesso concorrente via mutex. Estrutura principal: `Blockchain` com `Blocks []Block`, `State map[string]string`, `Difficulty int32` |
| **blockchain.go** | Rastreamento de Estado | Mapeia propriedade atual de ativos (`AssetID → OwnerID`) em tempo real para validação de transferências |
| **blockchain.go** | Genesis Block | Cria bloco inicial (Block #0) com valores hardcoded, hash zerado do anterior, e timestamp de criação |
| **blockchain.go** | Validação de Cadeia | Verifica integridade criptográfica de toda a blockchain, validando hashes encadeados e Proof of Work |
| **blockchain.go** | Resolução de Forks | Implementa "Longest Chain Rule": substitui cadeia local se cadeia recebida tiver mais blocos. Tie-breaker determinístico em caso de empate |
| **blockchain.go** | Substituição de Cadeia | `ReplaceChain()`: troca cadeia completa, reconstrói estado do zero, e limpa mempool de transações já confirmadas |
| **blockchain.go** | Sincronização de Estado | Reconstrói mapa `State` reprocessando todas as transações da cadeia recebida |
| **block.go** | Estrutura de Dados | Define `Block` com `Header` (metadados + PoW) e `Body` (transações). Header contém: `HashOfPrevious`, `MerkleRootHash`, `Timestamp`, `Nbits`, `Nonce`, `Hash`, `MinerID` |
| **block.go** | Merkle Root | Calcula hash consolidado de todas as transações do bloco para validação de integridade |
| **block.go** | Encadeamento Criptográfico | Liga blocos através de `HashOfPrevious`, tornando impossível adulteração sem refazer todo o trabalho subsequente |
| **pow.go** | Proof of Work | Implementa algoritmo PoW: hash do bloco deve iniciar com N zeros hexadecimais (dificuldade ajustável de 0 a 32) |
| **pow.go** | Cálculo de Hash | `CalculateBlockHash()`: SHA-256 concatenando `HashOfPrevious + MerkleRootHash + Timestamp + Nbits + Nonce` |
| **pow.go** | Verificação de Dificuldade | `CheckDifficulty()`: valida se hash começa com quantidade exigida de zeros à esquerda |
| **pow.go** | Geração de Merkle Root | `GenMerkleRoot()`: combina hashes de transações em árvore binária até raiz única |
| **validation.go** | Validação de Bloco Único | `IsValidNewBlock()`: verifica hash do bloco anterior, dificuldade PoW, e merkle root |
| **validation.go** | Validação de Cadeia Completa | `IsValidChain()`: itera todos os blocos validando encadeamento e proof of work |
| **validation.go** | Validação Criptográfica | Verifica assinaturas digitais ED25519 e hashes SHA-256 de blocos e transações |

---

## 💸 2. Sistema de Transações

| Módulo | Funcionalidade | Descrição |
|--------|----------------|-----------|
| **transaction.go** | Estrutura de Transação | Define `Transaction` com campos: `TXid`, `SenderID`, `ReceiverID`, `AssetID`, `Action`, `Fee`, `SenderPubKey`, `Signature` |
| **transaction.go** | Tipos de Ação | Suporta 2 ações nativas: `"register"` (criar novo ativo) e `"transfer"` (mudar propriedade) |
| **transaction.go** | Identificador Único | `TXid`: hash SHA-256 da transação serializada para identificação global |
| **transaction.go** | Serialização | Converte campos em byte slice (`SenderID + ReceiverID + AssetID + Action + Fee`) para hash e assinatura |
| **transaction.go** | Criação de Transação | `NewAndPostTransaction()`: recebe transação assinada, valida assinatura, gera TXid, e adiciona à mempool |
| **transaction.go** | Validação de Transação | `ValidateTransaction()`: verifica assinatura ED25519, campos obrigatórios, e fee mínimo |
| **transaction.go** | Sistema de Taxas | Campo `Fee` permite mineradores priorizarem transações com maior recompensa |
| **mempool.go** | Pool de Transações Pendentes | Mantém mapa thread-safe de transações aguardando confirmação. Capacidade padrão: 1000 transações |
| **mempool.go** | Prevenção de Gasto Duplo | `AddTransactionToMempool()`: rejeita múltiplas transferências pendentes do mesmo ativo |
| **mempool.go** | Ordenação por Fee | `GetTransactionsMinerMempool()`: retorna transações ordenadas por taxa (maior → menor) para maximizar lucro do minerador |
| **mempool.go** | Limpeza Pós-Mineração | Remove transações confirmadas em blocos recebidos para evitar duplicação |
| **mempool.go** | Controle de Capacidade | Limita quantidade de transações pendentes (configurável) para prevenir ataques de spam |
| **keys.go** | Geração de Chaves | `GenerateKeyPair()`: cria par ED25519 (privada 64 bytes, pública 32 bytes) |
| **keys.go** | Derivação de Endereço | `PublicKeyToAddress()`: converte chave pública em endereço hexadecimal via hash SHA-256 |
| **keys.go** | Geração Determinística | Suporta criação de chaves a partir de seed phrase para testes reproduzíveis |
| **signature.go** | Assinatura Digital | Implementa assinatura ED25519 (64 bytes) de dados serializados |
| **signature.go** | Verificação de Assinatura | `VerifySignature()`: valida autenticidade usando chave pública do remetente |
| **signature.go** | Não-Repúdio | Assinaturas garantem que apenas dono da chave privada pode autorizar transferências |

---

## ⛏️ 3. Mineração e Consenso

| Módulo | Funcionalidade | Descrição |
|--------|----------------|-----------|
| **miner.go** | Loop de Mineração | `BlockMiner()`: goroutine que executa ciclos de 5 segundos minerando continuamente |
| **miner.go** | Assembly de Blocos | `AssemblerNextBlockMiner()`: extrai até 10 transações da mempool e cria estrutura de bloco |
| **miner.go** | Caça de Nonce | Incrementa nonce (0 → ∞) até encontrar hash que satisfaz dificuldade PoW |
| **miner.go** | Validação Pré-Mineração | `ValidateTransactionsAgainstState()`: filtra transações inválidas antes de incluir no bloco |
| **miner.go** | Atribuição de Minerador | Registra `MinerID` no header do bloco para rastreamento de recompensas |
| **miner.go** | Broadcasting de Bloco | Publica bloco minerado no tópico Kafka para propagação na rede |
| **miner.go** | Seleção de Transações | Prioriza transações por fee, respeitando limite de 10 tx/bloco |
| **blockchain.go** | Aplicação de Estado | `MineAndAddBlock()`: adiciona bloco à cadeia e atualiza mapa `State` com novas propriedades |
| **blockchain.go** | Consenso Distribuído | Múltiplos nós minerando simultaneamente, rede converge para cadeia mais longa |
| **blockchain.go** | Suporte a Forks | Permite mineração simultânea, sistema resolve forks via tie-breaker determinístico |
| **blockchain.go** | Ajuste de Dificuldade | Permite alteração dinâmica do número de zeros exigidos (0-32) via API |

---

## 🌐 4. Rede P2P e Comunicação

| Módulo | Funcionalidade | Descrição |
|--------|----------------|-----------|
| **kafka.go** | Broker P2P | `P2PBroker`: gerencia conexão com Apache Kafka (porta 9092) para comunicação descentralizada |
| **kafka.go** | Tópicos Duplos | Usa 2 tópicos: `"blocks"` (blocos minerados) e `"txs"` (transações pendentes) |
| **kafka.go** | Publicação de Blocos | `PublishBlock()`: serializa bloco em JSON e envia para tópico "blocks" |
| **kafka.go** | Publicação de Transações | `PublishTransaction()`: serializa transação em JSON e envia para tópico "txs" |
| **kafka.go** | Consumo de Blocos | `consumeBlocks()`: goroutine que escuta tópico "blocks", valida e integra blocos recebidos |
| **kafka.go** | Consumo de Transações | `consumeTxs()`: goroutine que escuta tópico "txs", valida e adiciona à mempool local |
| **kafka.go** | Validação P2P | Filtra transações recebidas via rede, rejeitando as que violam estado atual |
| **kafka.go** | Propagação em Tempo Real | Blocos e transações se espalham pela rede em < 1 segundo |
| **kafka.go** | Resiliência de Rede | Kafka garante entrega mesmo se nós temporariamente offline |
| **kafka.go** | Deserialização | Converte JSON recebido de volta para structs Go (`Block`, `Transaction`) |
| **kafka.go** | Inicialização de Broker | `InitKafkaBroker()`: conecta ao Kafka e inicia consumers em background |

---

## 🔌 5. API REST e Interface Web

| Módulo | Funcionalidade | Descrição |
|--------|----------------|-----------|
| **api.go** | Servidor HTTP | `StartAPI()`: inicia servidor web na porta especificada (padrão: 8080, 8081, 8082) |
| **api.go** | `GET /blockchain` | Retorna cadeia completa em JSON: todos os blocos com índice, hash, transações, nonce, minerador |
| **api.go** | `POST /transactions` | Endpoint para submissão de transações assinadas. Payload: `sender`, `receiver`, `asset`, `action`, `fee`, `pub_key`, `signature` |
| **api.go** | `GET /mempool` | Lista todas as transações pendentes na mempool local |
| **api.go** | `GET /state` | Retorna mapa de propriedade atual: `{"AssetID": "OwnerID", ...}` |
| **api.go** | `POST /difficulty?value=X` | Ajusta dificuldade de mineração dinamicamente (0-32 zeros) |
| **api.go** | `POST /simulate-tx` | Gera transação aleatória usando seed reproduzível: register ou transfer com asset/owner aleatórios |
| **api.go** | `POST /simulate-attack` | Simula ataque de gasto duplo tentando transferir "Copacabana-Palace" para dois destinatários diferentes |
| **api.go** | CORS Headers | Habilita acesso cross-origin para dashboard web externo |
| **api.go** | Geração de Assets Aleatórios | `randomAssetGenerator()`: cria IDs como "Farm-North-42", "Mansion-Coast-789" |
| **api.go** | Chaves Determinísticas | `generateKeysFromSeed()`: cria chaves ED25519 a partir de seed SHA-256 para testes |
| **web/index.html** | Dashboard Visual | Interface HTML/CSS/JS para monitorar 3 nós simultaneamente |
| **web/index.html** | Visualização de Blockchain | Exibe crescimento da cadeia em tempo real com hashes, transações, e mineradores |
| **web/index.html** | Monitor de Mempool | Mostra transações pendentes de cada nó lado a lado |
| **web/index.html** | Controles Interativos | Botões: "Generate 1 Tx" (criar transação), "Simulate Attack" (gasto duplo) |
| **web/index.html** | Slider de Dificuldade | Ajusta Nbits (0-32) e observa impacto no tempo de mineração |
| **web/index.html** | Visualização de Forks | Mostra múltiplas pontas da cadeia quando nós mineram simultaneamente |
| **web/scripts.js** | Polling Automático | Atualiza dados da blockchain a cada 2 segundos via fetch API |
| **web/scripts.js** | Formatação de Hashes | Trunca hashes para exibição (primeiros/últimos 8 chars) |

---

## 🔒 6. Segurança e Validação

| Módulo | Funcionalidade | Descrição |
|--------|----------------|-----------|
| **Camada 1: Mempool** | Validação de Mempool | Bloqueia múltiplas transferências pendentes do mesmo ativo na mempool |
| **Camada 2: Estado** | Validação de Propriedade | `ValidateTransactionsAgainstState()`: rejeita transferências se remetente não é dono confirmado |
| **Camada 3: Blockchain** | Imutabilidade Criptográfica | Hash encadeado torna impossível alterar histórico sem refazer PoW de toda a cadeia subsequente |
| **Camada 4: Consenso** | Consenso Distribuído | Fork resolution garante cadeia canônica única, impedindo duplo gasto em forks paralelos |
| **Assinaturas ED25519** | Autenticação Forte | Curva elíptica de 256 bits, impossível forjar sem chave privada |
| **Hashes SHA-256** | Integridade de Dados | Qualquer alteração em bloco/transação invalida hash, detectável por toda a rede |
| **Validação de Ação** | Verificação de Registro | Impede transferência de ativo não registrado previamente |
| **Validação de Ownership** | Prevenção de Fraude | Rejeita transações de ativos que o sender não possui |
| **Replay Attack Protection** | TXid Único | Hash da transação previne resubmissão de transações já confirmadas |
| **Sybil Resistance** | Proof of Work | Custo computacional impede que atacante crie blocos fraudulentos mais rápido que rede honesta |
| **Ataque de Gasto Duplo** | Demonstração de Segurança | Sistema detecta e rejeita tentativas de transferir mesmo ativo simultaneamente para 2 destinatários |

---

## 🛠️ 7. Deployment e Comandos

| Módulo | Funcionalidade | Descrição |
|--------|----------------|-----------|
| **docker-compose.yml** | Orquestração Automática | Sobe rede completa (Kafka + 3 nós) com comando `docker-compose up --build` |
| **docker-compose.yml** | Configuração de Nós | Define 3 serviços: `node-alpha` (8080), `node-beta` (8081), `node-gamma` (8082) |
| **docker-compose.yml** | Kafka Container | Apache Kafka na porta 9092 como backbone P2P da rede |
| **Dockerfile** | Build Multi-Stage | Compila Go em container builder, executa em Alpine Linux (imagem mínima) |
| **cmd/sim/main.go** | Simulador de Nó | Entry point principal com flags: `-port` (API), `-id` (identificador), `-diff` (dificuldade inicial) |
| **cmd/sim/main.go** | Inicialização Completa | Cria blockchain, mempool, broker Kafka, inicia API e mineração |
| **cmd/miner/main.go** | Minerador Stand-alone | Versão separada focada apenas em mineração (stub) |
| **cmd/client/main.go** | Cliente CLI | Utilitário para interação via linha de comando (stub) |
| **go.mod** | Gerenciamento de Deps | Módulo `github.com/diskrat/dominium`, dependência única: `kafka-go v0.4.50` |
| **Comandos Manuais** | Execução Individual | `go run cmd/sim/main.go -port=8080 -id=node-alpha -diff=2` para cada nó |
| **Logs Docker** | Monitoramento | `docker logs -f node-alpha` para acompanhar atividade de cada nó |
| **multitail** | Visualização 3 Colunas | `multitail -s 3 -l "docker logs -f node-alpha" -l "..." -l "..."` para logs lado a lado |

---

## 📊 Especificações Técnicas

| Especificação | Valor | Observações |
|---------------|-------|-------------|
| **Linguagem** | Go 1.23 | Toolchain 1.23.5 |
| **Algoritmo de Hash** | SHA-256 | Para blocos, transações e endereços |
| **Assinatura Digital** | ED25519 | Chave privada 64 bytes, pública 32 bytes, assinatura 64 bytes |
| **Dificuldade PoW** | 0-32 zeros hex | Ajustável dinamicamente via API |
| **Capacidade Mempool** | 1000 transações | Configurável, previne spam |
| **Transações/Bloco** | Máximo 10 | Limite de assembly para blocos |
| **Intervalo de Mineração** | 5 segundos | Loop de checagem, não tempo fixo de bloco |
| **Nós Padrão** | 3 (Alpha, Beta, Gamma) | Escalável para N nós |
| **Porta Kafka** | 9092 | Broker P2P centralizado |
| **Portas API** | 8080, 8081, 8082 | Uma por nó |
| **Armazenamento** | In-memory | Sem persistência em disco (blockchain se perde ao reiniciar) |
| **Tamanho Código** | ~2500 linhas Go | 15 arquivos .go + 3 web |
| **Dependências Externas** | 1 (kafka-go) | Além da stdlib Go |
| **Formato de Comunicação** | JSON | Para blocos e transações via Kafka e API |
| **Mecanismo de Consenso** | Proof of Work | Longest Chain Rule + tie-breaker determinístico |
| **Fork Resolution** | Maior número de blocos | Se empate, compara hashes de blocos finais |

---

## 🎯 Casos de Uso Demonstrados

| Caso de Uso | Implementação | Endpoint/Função |
|-------------|---------------|-----------------|
| **Registro de Novo Ativo** | Cria ativo com `action: "register"` | `POST /simulate-tx` ou `POST /transactions` |
| **Transferência de Propriedade** | Muda ownership com `action: "transfer"` | `POST /simulate-tx` ou `POST /transactions` |
| **Consulta de Propriedade** | Verifica dono atual via mapa State | `GET /state` |
| **Histórico de Ativo** | Rastreia transferências na blockchain | `GET /blockchain` + filtrar por AssetID |
| **Ataque de Gasto Duplo** | Tenta transferir ativo para 2 destinos | `POST /simulate-attack` |
| **Ajuste de Segurança** | Aumenta dificuldade para prevenir ataques | `POST /difficulty?value=20` |
| **Mineração Competitiva** | 3 nós tentam minerar simultaneamente | Executar 3 containers |
| **Sincronização de Rede** | Nó novo recebe cadeia mais longa | Fork resolution automática |
| **Transparência Total** | Qualquer um pode auditar blockchain | `GET /blockchain` (público) |

---

## 🏗️ Arquitetura de Deployment

```
┌──────────────────────────────────────────────────────────────┐
│                 Docker Compose Network                        │
│                                                                │
│   ┌──────────────┐   ┌──────────────┐   ┌──────────────┐    │
│   │ Node Alpha   │   │  Node Beta   │   │ Node Gamma   │    │
│   │ (Port 8080)  │   │ (Port 8081)  │   │ (Port 8082)  │    │
│   ├──────────────┤   ├──────────────┤   ├──────────────┤    │
│   │ Blockchain   │   │ Blockchain   │   │ Blockchain   │    │
│   │ Mempool      │   │ Mempool      │   │ Mempool      │    │
│   │ State Map    │   │ State Map    │   │ State Map    │    │
│   │ Miner Loop   │   │ Miner Loop   │   │ Miner Loop   │    │
│   │ HTTP API     │   │ HTTP API     │   │ HTTP API     │    │
│   │ Kafka Client │   │ Kafka Client │   │ Kafka Client │    │
│   └──────┬───────┘   └──────┬───────┘   └──────┬───────┘    │
│          │                   │                   │            │
│          └───────────────────┼───────────────────┘            │
│                              │                                │
│                   ┌──────────▼──────────┐                     │
│                   │   Apache Kafka      │                     │
│                   │   (Port 9092)       │                     │
│                   ├─────────────────────┤                     │
│                   │ Topic: "blocks"     │                     │
│                   │ Topic: "txs"        │                     │
│                   └─────────────────────┘                     │
│                                                                │
└──────────────────────────────────────────────────────────────┘

Acesso Frontend: http://localhost:8080 (Dashboard Node Alpha)
                  http://localhost:8081 (Dashboard Node Beta)
                  http://localhost:8082 (Dashboard Node Gamma)
```

---

## 📝 Notas Finais

- **Propósito**: Projeto educacional para demonstração de conceitos blockchain
- **Não usar em produção**: Sem persistência, criptografia industrial, ou escalabilidade real
- **Aprendizados**: Consenso distribuído, criptografia aplicada, P2P networking, fork resolution
- **Inspiração**: Baseado em conceitos do Bitcoin whitepaper (Satoshi Nakamoto, 2008)
- **Licença**: Verificar arquivo LICENSE no repositório

---

**Documento gerado em**: 2026-04-04 
**Analisado por**: GitHub Copilot CLI (Claude Sonnet 4.5)
