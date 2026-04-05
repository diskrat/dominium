# Copilot Instructions for `dominium`

## Build, test, and lint commands

### Core backend (repo root)

- Run all tests:
  - `go test ./...`
- Run a single test:
  - `go test ./internal/transaction -run TestSignVerifyECDSA -v`
- Build node binary:
  - `go build ./cmd/node`
- Run one node (no RabbitMQ):
  - `go run ./cmd/node -id node-1 -api :8080`
- Run one node with RabbitMQ:
  - `go run ./cmd/node -id node-1 -api :8080 -amqp amqp://guest:guest@localhost:5672/`

### `front-p2p` module

- Run all tests:
  - `cd front-p2p && make test`
- Run a single test subset:
  - `cd front-p2p && make test-p2p`
  - `cd front-p2p && make test-blockchain`
- Lint/format (module-local):
  - `cd front-p2p && make lint`
- Run mock server:
  - `cd front-p2p && make mock-server`
- Run frontend static server:
  - `cd front-p2p && make frontend`

## High-level architecture

- The current backend runtime entrypoint is `cmd/node/main.go`.
- The core domain logic is in `internal/transaction/`:
  - `transaction.go`: canonical serialization, deterministic ID, signing.
  - `contracts.go`: transaction validation and execution rules for NFT mint/transfer.
  - `state.go`: in-memory account ownership state (`Accounts`, `ExistingNFTs`).
  - `mempool.go`: pending transaction set validated against state.
  - `wallet.go` and `generator.go`: identity/key handling and deterministic demo transaction generation.
- The node exposes HTTP endpoints consumed by `front-p2p/frontend`:
  - `GET /health`, `GET /blocks`, `GET /mempool`, `GET /stats`, `POST /api/transactions`.
- RabbitMQ integration in `cmd/node/main.go` is optional (`-amqp`):
  - topic exchange: `blockchain`
  - queue per node: `node_<id>`
  - routing key used for tx propagation: `tx.new`
  - message envelope matches front-p2p expectations (`type: "TX_NEW"`, `payload.tx_json`, `sender`).
- `front-p2p` remains a separable module for frontend + P2P experiments (with mocks and test-focused workflows), but the frontend is configured to hit real node endpoints on `localhost:8080/8081/8082`.

## Key conventions in this repository

- Source of truth for behavior is current Go code (`cmd/node`, `internal/transaction`) rather than older conceptual docs.
- Transaction flow convention:
  - ensure `Transaction.ID` is set from canonical payload before acceptance,
  - validate through `tx.Validate(state)`,
  - execute via `tx.Execute(state)`,
  - treat mempool as pending queue (remove once applied in node flow).
- Transaction IDs are SHA-256 of the serialized payload (`pkg/crypto.Hash`), not random UUIDs.
- Signatures use ECDSA P-256 with ASN.1 encoding (`SignECDSA`/`VerifyECDSA`); keep this format stable across integrations.
- API and domain error strings are written in Portuguese; keep new errors/messages consistent.
- Concurrency-sensitive state (`AccountState`, `Mempool`) uses `sync.RWMutex`; preserve lock discipline when adding new fields/operations.
- Branch and commit conventions come from `CONTRIBUTING.md`:
  - branch naming: `feature/*`, `bugfix/*`, `docs/*`
  - commit style: Conventional Commits (`feat:`, `fix:`, `docs:`, `refactor:`, `test:`)
  - `main` is protected; work via PR.
