# Front & P2P - Componentes de Teste

Esta subpasta contém os componentes de **Frontend** e **P2P (Mensageria)** separados do core principal da blockchain, permitindo testes isolados com mocks.

## 📁 Estrutura

```
front-p2p/
├── frontend/           # Interface web (HTML, CSS, JS)
├── p2p/               # Camada de mensageria (RabbitMQ)
├── mocks/             # Implementações mock para testes
├── tests/             # Testes unitários
├── docker/            # Configurações Docker isoladas
└── README.md          # Este arquivo
```

## 🎯 Objetivos

1. **Isolamento**: Testar componentes de frontend e P2P sem dependências do core blockchain
2. **Mocks**: Simulação de blockchain e mensageria para testes rápidos
3. **Integração**: Validar comunicação entre frontend e backend mockado

## 🧪 Componentes Mock

### `mocks/mock_rabbitmq.go`

Mock completo do RabbitMQ com:

- Simulação de publish/subscribe
- Controle de conexão/desconexão
- Histórico de mensagens publicadas
- Simulação de falhas
- Thread-safe

### `mocks/mock_blockchain.go`

Mock da blockchain com:

- Adição/consulta de blocos
- Gerenciamento de mempool
- Validação básica de integridade
- Transações mockadas

## 🧪 Testes

### Executar todos os testes

```bash
cd front-p2p/tests
go test -v
```

### Executar testes específicos

```bash
# Testes P2P
go test -v -run TestMockRabbitMQ

# Testes Blockchain
go test -v -run TestMockBlockchain
```

### Executar benchmarks

```bash
go test -bench=. -benchmem
```

## 📊 Cobertura de Testes

### P2P (mock_rabbitmq_test.go)

- ✅ Conexão/Desconexão
- ✅ Publicação de mensagens
- ✅ Subscrição e handlers
- ✅ Filtro por tipo de mensagem
- ✅ Simulação de falhas
- ✅ Limpeza de histórico

### Blockchain (mock_blockchain_test.go)

- ✅ Criação de genesis block
- ✅ Adição de blocos válidos
- ✅ Validação de hash anterior
- ✅ Validação de índice
- ✅ Gerenciamento de mempool
- ✅ Consulta de blocos

## 🌐 Frontend

### Arquivos

- `index.html` - Interface visual da blockchain
- `app.js` - Lógica de visualização e polling
- `styles.css` - Estilos da interface
- `nginx.conf` - Configuração do servidor web
- `Dockerfile` - Container do frontend

### Executar Frontend Localmente

```bash
cd frontend
# Opção 1: Servidor Python
python3 -m http.server 8080

# Opção 2: Servidor Node
npx http-server -p 8080

# Opção 3: Nginx (Docker)
docker build -t blockchain-frontend .
docker run -p 8080:80 blockchain-frontend
```

### Conectar ao Mock

Altere a URL no `app.js` para apontar ao servidor mock:

```javascript
this.nodeURL = "http://localhost:9000"; // Servidor mock
```

## 🔧 Servidor Mock Standalone

Crie um servidor HTTP simples para testar o frontend:

```go
// Arquivo: cmd/mock-server/main.go
package main

import (
    "encoding/json"
    "net/http"
    "../mocks"
)

func main() {
    bc := mocks.NewMockBlockchain()
    mq := mocks.NewMockRabbitMQ("mock-server")

    http.HandleFunc("/blocks", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(bc.GetBlocks())
    })

    http.HandleFunc("/mempool", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]int{
            "count": len(bc.GetMempool()),
        })
    })

    http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Content-Type", "application/json")
        json.NewEncoder(w).Encode(map[string]interface{}{
            "status": "healthy",
            "blocks": bc.GetHeight(),
        })
    })

    http.ListenAndServe(":9000", nil)
}
```

Executar:

```bash
cd tests
go run mock_server.go
```

## 🐳 Docker Standalone

### Build

```bash
docker build -t blockchain-frontend -f frontend/Dockerfile frontend/
```

### Run

```bash
docker run -p 8080:80 blockchain-frontend
```

## 🔄 Fluxo de Testes

### 1. Testes Unitários (Mocks)

```bash
cd tests
go test -v
```

### 2. Testes de Integração (Mock Server + Frontend)

```bash
# Terminal 1: Servidor mock
go run cmd/mock-server/main.go

# Terminal 2: Frontend
cd frontend && python3 -m http.server 8080

# Browser: http://localhost:8080
```

### 3. Testes E2E (Com core real)

```bash
# Voltar ao diretório raiz
cd ..
docker-compose up
```

## 📝 Casos de Uso

### Teste 1: Validar Frontend sem Backend Real

1. Inicie o mock server
2. Abra o frontend
3. Valide renderização de blocos mockados

### Teste 2: Validar Mensageria P2P

1. Execute testes unitários do mock_rabbitmq
2. Valide publish/subscribe
3. Teste cenários de falha

### Teste 3: Performance de Mensageria

```bash
go test -bench=BenchmarkMockRabbitMQ -benchmem
```

## 🔗 Integração com Core Principal

Os componentes aqui são **cópias** dos originais:

- **Original**: `blockchain2/internal/p2p/` e `blockchain2/frontend/`
- **Mock**: `blockchain2/front-p2p/`

Modificações aqui **não afetam** o core principal.

## 📦 Dependências

### Go (testes)

```bash
go mod init front-p2p
go mod tidy
```

### Frontend (nenhuma)

- HTML5/CSS3/JavaScript puro
- Sem frameworks necessários

## 🚀 Próximos Passos

1. [ ] Criar mock server HTTP completo
2. [ ] Adicionar testes de integração frontend + mock
3. [ ] Implementar simulação de rede P2P multi-nó
4. [ ] Criar dashboard de métricas de mensageria
5. [ ] Adicionar testes de carga para P2P

## 📚 Referências

- [Go Testing Package](https://pkg.go.dev/testing)
- [Table-Driven Tests](https://dave.cheney.net/2019/05/07/prefer-table-driven-tests)
- [Mock Best Practices](https://martinfowler.com/articles/mocksArentStubs.html)
