# 📖 Guia Rápido - front-p2p

## 🎯 Comandos Essenciais

### Navegação
```bash
cd front-p2p                    # Entra na pasta
```

### Menu Interativo (Recomendado)
```bash
./quickstart.sh                 # Menu com todas as opções
```

### Testes
```bash
make test                       # Todos os testes
make test-p2p                   # Só P2P
make test-blockchain            # Só blockchain
make test-coverage              # Com relatório HTML
```

### Execução
```bash
make mock-server                # Backend mock :9000
make frontend                   # Frontend :8080
make demo                       # Ambos juntos
```

### Docker
```bash
make docker-up                  # Sobe ambiente teste
make docker-down                # Para containers
make docker-logs                # Ver logs
```

### Utilidades
```bash
make help                       # Lista todos comandos
make clean                      # Limpa temporários
make stats                      # Estatísticas código
```

---

## 📂 Estrutura de Arquivos

```
front-p2p/
├── 📁 frontend/               # Interface web
│   ├── index.html
│   ├── app.js
│   ├── styles.css
│   └── Dockerfile
│
├── 📁 p2p/                    # Mensageria
│   ├── rabbitmq.go
│   └── messages.go
│
├── 📁 mocks/                  # Implementações mock
│   ├── mock_rabbitmq.go       # Mock RabbitMQ (175 linhas)
│   └── mock_blockchain.go     # Mock Blockchain (140 linhas)
│
├── 📁 tests/                  # Testes unitários
│   ├── p2p_test.go           # 10+ testes P2P
│   ├── blockchain_test.go    # 10+ testes blockchain
│   └── mock_server.go        # Servidor HTTP mock
│
├── 📁 docker/                 # Configs Docker
│   └── Dockerfile.mock
│
├── 📄 README.md              # Documentação completa
├── 📄 SUMMARY.md             # Resumo executivo
├── 📄 RENAME_CONFIRMATION.md # Confirmação renomeação
├── 🛠️  Makefile               # Automação (15+ comandos)
├── 🚀 quickstart.sh          # Menu interativo
├── 🧪 run_tests.sh           # Script de testes
├── 📦 go.mod                 # Módulo Go
└── 🐳 docker-compose.test.yml # Orquestração teste
```

---

## 🧪 Exemplos de Uso

### 1. Executar todos os testes
```bash
cd front-p2p
make test
```

### 2. Iniciar ambiente completo
```bash
cd front-p2p
make demo

# Acesse:
# - Backend: http://localhost:9000
# - Frontend: http://localhost:8080
```

### 3. Testar apenas P2P
```bash
cd front-p2p
make test-p2p
```

### 4. Ver cobertura de testes
```bash
cd front-p2p
make test-coverage
# Abre: tests/coverage.html
```

### 5. Ambiente Docker
```bash
cd front-p2p
make docker-up
```

---

## 🔧 Troubleshooting

### Problema: "Permission denied"
```bash
chmod +x quickstart.sh run_tests.sh
```

### Problema: "Go not found"
```bash
# Instale Go 1.25+
# Ubuntu/Debian:
sudo apt install golang

# Ou baixe de: https://go.dev/dl/
```

### Problema: "Port already in use"
```bash
# Mude a porta no mock_server.go
# Ou mate o processo:
lsof -ti:9000 | xargs kill
```

---

## 📊 Endpoints do Mock Server

Quando executar `make mock-server`, os seguintes endpoints estarão disponíveis:

| Endpoint | Método | Descrição |
|----------|--------|-----------|
| `/health` | GET | Status do servidor |
| `/blocks` | GET | Lista todos os blocos |
| `/mempool` | GET | Transações pendentes |
| `/stats` | GET | Estatísticas gerais |
| `/api/blocks` | POST | Adicionar bloco |
| `/api/transactions` | POST | Adicionar transação |

### Exemplo de uso:
```bash
# Health check
curl http://localhost:9000/health

# Listar blocos
curl http://localhost:9000/blocks

# Ver mempool
curl http://localhost:9000/mempool
```

---

## 🎓 Próximos Passos

1. **Explore o código**:
   ```bash
   cd front-p2p
   cat mocks/mock_rabbitmq.go
   cat tests/p2p_test.go
   ```

2. **Execute o menu**:
   ```bash
   ./quickstart.sh
   ```

3. **Rode os testes**:
   ```bash
   make test
   ```

4. **Veja a demo**:
   ```bash
   make demo
   ```

---

## 📚 Links Úteis

- **Documentação completa**: `README.md`
- **Resumo executivo**: `SUMMARY.md`
- **Confirmação renomeação**: `RENAME_CONFIRMATION.md`
- **Documentação raiz**: `../COMPONENT_SEPARATION.md`

---

## ⚡ Dicas

- Use `make help` para ver todos os comandos
- Execute `./quickstart.sh` para menu interativo
- Testes são rápidos (~2s vs 30s do sistema completo)
- Mock server gera blocos automaticamente a cada 10s
- Frontend faz polling a cada 2s

---

**Última atualização**: 2026-04-05  
**Versão**: 1.0 (renomeado front&p2p → front-p2p)
