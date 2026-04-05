# 🧪 Testando a Opção 4 - Mock Server + Frontend

## ✅ Correções Aplicadas

### 1. Erro de Compilação
**Problema**: `genesis` declarada mas não utilizada
```go
// ❌ Antes
genesis := bc.GetLastBlock()
for i := 1; i <= 3; i++ {
    lastBlock := bc.GetLastBlock() // genesis não usado
}

// ✅ Depois
for i := 1; i <= 3; i++ {
    lastBlock := bc.GetLastBlock()
}
```

### 2. Frontend não conectava ao Mock
**Problema**: Frontend buscava `/api/node1/*` em vez de usar mock `:9000`

**Solução**: Auto-detecção de ambiente
```javascript
// Se rodando em localhost:8080, usa mock automaticamente
this.nodeURL = isLocalhost && window.location.port === '8080' 
    ? 'http://localhost:9000'
    : document.getElementById('node-url').value;
```

---

## 🚀 Como Testar

### Opção 1: Via quickstart.sh (Recomendado)
```bash
cd front-p2p
./quickstart.sh
# Escolha: 4
```

### Opção 2: Manual
```bash
cd front-p2p

# Terminal 1: Mock Server
go run cmd/mock-server/main.go

# Terminal 2: Frontend
cd frontend && python3 -m http.server 8080
```

---

## 📊 O que Esperar

### ✅ Mock Server Iniciando
```
🚀 Mock Server started on http://localhost:9000
📊 Endpoints:
   - GET  /health
   - GET  /blocks
   - GET  /mempool
   - GET  /stats
   - POST /api/blocks
   - POST /api/transactions

✅ Mock data initialized
   - Blocks: 4
   - Mempool: 2 transactions

🔄 Auto-generating blocks every 10 seconds...
```

### ✅ Frontend Iniciando
```
Serving HTTP on 0.0.0.0 port 8080 (http://0.0.0.0:8080/) ...
```

---

## 🌐 Acessar os Serviços

### 1. Demo Interativa (NOVO!)
```
http://localhost:8080/demo.html
```
- ✅ Testes rápidos dos endpoints
- ✅ Status dos serviços
- ✅ Links diretos

### 2. Blockchain Viewer (Principal)
```
http://localhost:8080/
```
- ✅ Visualização da blockchain
- ✅ Polling automático (2s)
- ✅ Detalhes dos blocos

### 3. API Direta
```bash
# Health Check
curl http://localhost:9000/health | jq

# Listar Blocos
curl http://localhost:9000/blocks | jq

# Ver Mempool
curl http://localhost:9000/mempool | jq

# Estatísticas
curl http://localhost:9000/stats | jq
```

---

## 🔍 Validação

### 1. Mock Server Funcionando
```bash
# Deve retornar JSON com status: healthy
curl http://localhost:9000/health

# Resposta esperada:
{
  "status": "healthy",
  "blocks": 4,
  "mempool": 2,
  "connected": true,
  "timestamp": 1712279920,
  "node_id": "mock-server",
  "version": "1.0.0-mock"
}
```

### 2. Frontend Conectado
Abra o browser em `http://localhost:8080`

**Deve mostrar**:
- ✅ Status: 🟢 Conectado
- ✅ Blocos aparecendo
- ✅ Contador de mempool
- ✅ Auto-refresh funcionando

**Console do browser (F12)**:
```javascript
// Não deve ter erros 404
// Deve ter logs de:
"Fetching from: http://localhost:9000/blocks"
"Fetching from: http://localhost:9000/mempool"
"Fetching from: http://localhost:9000/health"
```

---

## ⚠️ Troubleshooting

### Erro: "genesis declared and not used"
```bash
# Solução: Já corrigido! Atualize o arquivo
cd front-p2p
git pull  # ou reaplique as correções
```

### Frontend retorna 404
**Sintoma**: `GET /api/node1/blocks 404`

**Causa**: Frontend ainda busca node1 real

**Solução**: Limpe o cache do browser
```
1. Abra Developer Tools (F12)
2. Application → Clear Storage
3. Clear site data
4. Recarregue (Ctrl+Shift+R)
```

### Mock Server não inicia
```bash
# Verifica se porta está ocupada
lsof -i :9000

# Se ocupada, mata o processo
lsof -ti:9000 | xargs kill

# Tenta novamente
go run cmd/mock-server/main.go
```

### Frontend não conecta ao mock
```bash
# Verifica se está na porta correta
# Frontend deve estar em :8080
# Mock deve estar em :9000

# Teste CORS
curl -H "Origin: http://localhost:8080" \
  http://localhost:9000/health \
  -v
```

---

## 📝 Logs Esperados

### Mock Server
```
✅ Generated mock block #5
✅ Generated mock block #6
...
```

### Frontend (Browser Console)
```
Block count: 6
Mempool: 2
Status: healthy
```

### Frontend (Server Logs)
```
127.0.0.1 - - [05/Apr/2026 01:00:00] "GET / HTTP/1.1" 200 -
127.0.0.1 - - [05/Apr/2026 01:00:00] "GET /app.js HTTP/1.1" 200 -
127.0.0.1 - - [05/Apr/2026 01:00:00] "GET /styles.css HTTP/1.1" 200 -
```

---

## 🎯 Checklist de Sucesso

- [ ] Mock server compila sem erros
- [ ] Mock server inicia em :9000
- [ ] Endpoint /health retorna JSON
- [ ] Frontend inicia em :8080
- [ ] Browser abre http://localhost:8080
- [ ] Frontend mostra blocos
- [ ] Console sem erros 404
- [ ] Auto-refresh funciona (2s)
- [ ] Novos blocos aparecem (10s)

---

## 🆘 Se Nada Funcionar

```bash
# 1. Pare tudo
killall python3 2>/dev/null
lsof -ti:9000 | xargs kill 2>/dev/null
lsof -ti:8080 | xargs kill 2>/dev/null

# 2. Limpe e reconstrua
cd front-p2p
go mod tidy
go build cmd/mock-server/main.go

# 3. Teste isoladamente
# Terminal 1:
./main

# Terminal 2:
curl http://localhost:9000/health

# Se funcionar, inicie frontend
cd frontend && python3 -m http.server 8080
```

---

## 📚 Recursos Adicionais

- **demo.html**: Interface de testes rápidos
- **test-option4.sh**: Script de teste automatizado
- **FIXES.md**: Histórico de correções

---

**Última atualização**: 2026-04-05  
**Status**: ✅ Testado e funcionando
