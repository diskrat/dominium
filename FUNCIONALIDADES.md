# Quadro de Funcionalidades - Dominium (estado atual da branch)

## Visao geral

Nesta branch, o backend foi evoluido para integrar o frontend `front-p2p` com o core transacional.

O sistema agora possui:

- core de transacoes criptograficas (ECDSA P-256)
- estado de contas/NFTs
- mempool
- API HTTP compativel com o frontend
- integracao opcional com RabbitMQ para propagacao de transacoes

---

## 1) Core transacional implementado

### 1.1 Modelo de transacao (`internal/transaction/transaction.go`)

Estrutura `Transaction`:

- `ID string`
- `Timestamp int64`
- `Sig []byte`
- `Type byte`
- `PublKey string`
- `Recipient string`
- `NFTID string`

Capacidades:

- serializacao canonica (`Serialize`)
- hash SHA-256 deterministico para ID (`ComputeID`, `SetID`)
- assinatura e verificacao ECDSA (`Sign`, `VerifySerialized`)

### 1.2 Regras de negocio (`internal/transaction/contracts.go`)

Tipos:

- `TypeTransferNFT` (1)
- `TypeMintNFT` (2)

Validacoes:

- integridade (ID confere com payload)
- assinatura valida
- campos obrigatorios
- regras de ownership/mint por tipo

Execucao:

- `Execute(state)` aplica mint/transfer no estado

### 1.3 Estado de contas (`internal/transaction/state.go`)

Estruturas:

- `Account{NFTs map[string]bool}`
- `AccountState{Accounts, ExistingNFTs}`

Funcionalidades:

- criar contas (`CreateAccount`)
- validar/aplicar mint e transfer
- clonar estado (`Clone`)
- estatisticas (`Stats`)

### 1.4 Mempool (`internal/transaction/mempool.go`)

Funcionalidades:

- adicionar tx com validacao (`Add`)
- evitar duplicidade por ID
- listar pendentes (`GetPending`)
- remover por IDs (`Remove`)
- contar pendentes (`Count`)

### 1.5 Chaves e carteiras

Arquivos:

- `internal/transaction/keys.go`
- `internal/transaction/signature.go`
- `internal/transaction/wallet.go`
- `internal/transaction/generator.go`

Capacidades:

- geracao de pares ECDSA P-256
- encode/decode de chaves
- identidades nomeadas (admin/alice/bob)
- geracao deterministica de txs de demo

---

## 2) Backend HTTP integrado ao frontend (`cmd/node/main.go`)

O `cmd/node` agora sobe API com CORS e ciclo de vida completo (start/stop graceful).

### Endpoints expostos

- `GET /health`
  - retorna status, blocos sinteticos, mempool, conectividade e node_id
- `GET /blocks`
  - retorna cadeia sintetica (`BlockView[]`) para visualizacao no frontend
- `GET /mempool`
  - retorna lista de transacoes pendentes
- `GET /stats`
  - retorna metricas do estado e conectividade P2P
- `POST /api/transactions`
  - injeta transacao assinada no fluxo do node

### Flags do node

- `-api` (padrao `:8080`)
- `-id` (padrao `node-1`)
- `-amqp` (padrao vazio; quando preenchida, habilita RabbitMQ)

---

## 3) Integracao de mensageria RabbitMQ (opcional)

Implementada no `cmd/node/main.go` com protocolo compativel ao `front-p2p/p2p/messages.go` para `TX_NEW`.

Comportamento:

- declara exchange topic `blockchain`
- declara fila por no: `node_<id>`
- bind em `tx.*`
- publica transacoes em `tx.new`
- consome envelopes com formato:

```json
{
  "type": "TX_NEW",
  "payload": { "tx_json": "{...}" },
  "sender": "node-id"
}
```

- ignora mensagens do proprio no
- faz `ack` manual apos processamento

Observacao: se `-amqp` nao for informado, o node roda somente local (sem P2P externo).

---

## 4) Frontend front-p2p integrado

Arquivos alterados:

- `front-p2p/frontend/index.html`
- `front-p2p/frontend/app.js`

Ajustes:

- seletor de no agora aponta para:
  - `http://localhost:8080`
  - `http://localhost:8081`
  - `http://localhost:8082`
- removida dependencia fixa do mock `http://localhost:9000`

Resultado: o frontend passa a consumir o backend real do `cmd/node`.

---

## 5) Dependencias e validacao

### Dependencias

- `go.mod` atualizado com:
  - `github.com/rabbitmq/amqp091-go v1.10.0`
- `go.sum` gerado/atualizado

### Validacao executada

- `go test ./...` com sucesso
- smoke test dos endpoints (`/health`, `/blocks`, `/mempool`) com sucesso

---

## 6) Escopo atual e proximos passos

O projeto agora entrega o acoplamento base frontend + backend + mensageria de transacoes.

Proximos passos recomendados:

1. adicionar testes automatizados para handlers HTTP do `cmd/node`
2. formalizar contrato JSON de `POST /api/transactions`
3. evoluir sincronizacao de blocos entre nos (alem de tx)
4. substituir bloco sintetico por modelo de bloco definitivo (se entrar no escopo)

---

**Documento atualizado em:** 2026-04-05  
**Fonte da verdade:** `cmd/node`, `internal/transaction`, `front-p2p/frontend`, `go.mod`
