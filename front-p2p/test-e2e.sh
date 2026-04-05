#!/bin/bash

# Teste End-to-End da Opção 4: Mock Server + Frontend
# Valida que a correção está funcionando

set -e

echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║   🧪 Teste E2E - Mock Server + Frontend                      ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# Cores
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m'

success() { echo -e "${GREEN}✓${NC} $1"; }
error() { echo -e "${RED}✗${NC} $1"; }
info() { echo -e "${YELLOW}▶${NC} $1"; }

cleanup() {
    info "Limpando processos..."
    pkill -f "mock-server" 2>/dev/null || true
    pkill -f "http.server.*8080" 2>/dev/null || true
    sleep 1
    success "Cleanup concluído"
}

trap cleanup EXIT

echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
info "Passo 1: Compilando Mock Server"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if go build -o /tmp/test-mock-server cmd/mock-server/main.go 2>&1; then
    success "Mock server compilado sem erros"
else
    error "Falha na compilação"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
info "Passo 2: Iniciando Mock Server"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

nohup /tmp/test-mock-server > /tmp/mock-server.log 2>&1 &
MOCK_PID=$!
echo "PID: $MOCK_PID"

info "Aguardando inicialização (5s)..."
sleep 5

if ps -p $MOCK_PID > /dev/null 2>&1; then
    success "Mock server está rodando"
else
    error "Mock server não iniciou"
    cat /tmp/mock-server.log
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
info "Passo 3: Testando Endpoints"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Teste Health
info "Testando /health..."
HEALTH=$(curl -s http://localhost:9000/health)
if echo "$HEALTH" | jq -e '.status == "healthy"' > /dev/null 2>&1; then
    success "/health retornou status healthy"
else
    error "/health falhou"
    echo "$HEALTH"
    exit 1
fi

# Teste Blocks
info "Testando /blocks..."
BLOCKS=$(curl -s http://localhost:9000/blocks)
BLOCK_COUNT=$(echo "$BLOCKS" | jq '. | length' 2>/dev/null)
if [ "$BLOCK_COUNT" -ge 4 ]; then
    success "/blocks retornou $BLOCK_COUNT blocos"
else
    error "/blocks falhou (esperado >= 4, obteve $BLOCK_COUNT)"
    exit 1
fi

# Teste Mempool
info "Testando /mempool..."
MEMPOOL=$(curl -s http://localhost:9000/mempool)
if echo "$MEMPOOL" | jq -e 'type == "array"' > /dev/null 2>&1; then
    TX_COUNT=$(echo "$MEMPOOL" | jq '. | length')
    success "/mempool retornou $TX_COUNT transações"
else
    error "/mempool falhou"
    exit 1
fi

# Teste Stats
info "Testando /stats..."
STATS=$(curl -s http://localhost:9000/stats)
if echo "$STATS" | jq -e '.node_id' > /dev/null 2>&1; then
    success "/stats retornou dados válidos"
else
    error "/stats falhou"
    exit 1
fi

# Teste CORS
info "Testando CORS..."
CORS=$(curl -s -H "Origin: http://localhost:8080" \
           -H "Access-Control-Request-Method: GET" \
           -I http://localhost:9000/health 2>&1 | grep -i "access-control")
if [ -n "$CORS" ]; then
    success "CORS habilitado"
else
    error "CORS não configurado"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
info "Passo 4: Validando Frontend"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

# Verifica arquivos
for file in index.html app.js styles.css demo.html; do
    if [ -f "frontend/$file" ]; then
        success "frontend/$file existe"
    else
        error "frontend/$file não encontrado"
        exit 1
    fi
done

# Verifica auto-detecção no código
if grep -q "localhost:9000" frontend/app.js; then
    success "Auto-detecção de localhost:9000 presente no app.js"
else
    error "Auto-detecção não encontrada em app.js"
    exit 1
fi

echo ""
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
info "Passo 5: Testando Geração Automática de Blocos"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

BLOCKS_BEFORE=$(curl -s http://localhost:9000/blocks | jq '. | length')
info "Blocos atuais: $BLOCKS_BEFORE"
info "Aguardando 12s para novo bloco..."
sleep 12

BLOCKS_AFTER=$(curl -s http://localhost:9000/blocks | jq '. | length')
info "Blocos após espera: $BLOCKS_AFTER"

if [ "$BLOCKS_AFTER" -gt "$BLOCKS_BEFORE" ]; then
    success "Geração automática funcionando (+$(($BLOCKS_AFTER - $BLOCKS_BEFORE)) blocos)"
else
    error "Nenhum bloco novo gerado"
    exit 1
fi

echo ""
echo "╔═══════════════════════════════════════════════════════════════╗"
echo "║                                                               ║"
echo "║   ✅ TODOS OS TESTES PASSARAM!                               ║"
echo "║                                                               ║"
echo "╚═══════════════════════════════════════════════════════════════╝"
echo ""

echo "📊 Resumo:"
echo "  ✅ Mock server compila e inicia"
echo "  ✅ Todos endpoints funcionando"
echo "  ✅ CORS configurado"
echo "  ✅ Frontend preparado"
echo "  ✅ Auto-detecção implementada"
echo "  ✅ Geração automática de blocos"
echo ""

echo "🚀 Para usar:"
echo "  1. ./quickstart.sh → Opção 4"
echo "  2. Acesse http://localhost:8080"
echo "  3. Ou teste com: http://localhost:8080/demo.html"
echo ""

exit 0
