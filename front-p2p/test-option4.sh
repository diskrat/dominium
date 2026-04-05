#!/bin/bash

# Teste isolado da opção 4 do quickstart
# Para debug e validação

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}Testing Option 4: Mock Server + Frontend${NC}"
echo ""

# Função de limpeza
cleanup() {
    echo ""
    echo -e "${YELLOW}Cleaning up...${NC}"
    if [ ! -z "$MOCK_PID" ]; then
        kill $MOCK_PID 2>/dev/null && echo "  ✓ Mock server stopped" || echo "  - Mock server already stopped"
    fi
    if [ ! -z "$FRONTEND_PID" ]; then
        kill $FRONTEND_PID 2>/dev/null && echo "  ✓ Frontend stopped" || echo "  - Frontend already stopped"
    fi
    echo -e "${GREEN}Cleanup complete${NC}"
    exit 0
}

# Registra cleanup para Ctrl+C
trap cleanup INT TERM EXIT

# Verifica dependências
echo -e "${BLUE}Checking dependencies...${NC}"

if ! command -v go &> /dev/null; then
    echo -e "${RED}✗ Go not found${NC}"
    echo "  Install Go: https://go.dev/dl/"
    exit 1
fi
echo "  ✓ Go found: $(go version | awk '{print $3}')"

if ! command -v python3 &> /dev/null; then
    echo -e "${RED}✗ Python3 not found${NC}"
    echo "  Install Python3: sudo apt install python3"
    exit 1
fi
echo "  ✓ Python3 found: $(python3 --version)"

# Verifica portas
echo ""
echo -e "${BLUE}Checking ports...${NC}"

if lsof -Pi :9000 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "${RED}✗ Port 9000 is already in use${NC}"
    echo "  Run: lsof -ti:9000 | xargs kill"
    exit 1
fi
echo "  ✓ Port 9000 available"

if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
    echo -e "${RED}✗ Port 8080 is already in use${NC}"
    echo "  Run: lsof -ti:8080 | xargs kill"
    exit 1
fi
echo "  ✓ Port 8080 available"

# Inicia mock server
echo ""
echo -e "${GREEN}Starting mock server on :9000...${NC}"
go run cmd/mock-server/main.go > /tmp/mock-server.log 2>&1 &
MOCK_PID=$!

sleep 2

# Verifica se o mock server está rodando
if ! kill -0 $MOCK_PID 2>/dev/null; then
    echo -e "${RED}✗ Mock server failed to start${NC}"
    echo "  Check logs: cat /tmp/mock-server.log"
    cat /tmp/mock-server.log
    exit 1
fi

# Testa o health endpoint
if curl -s http://localhost:9000/health > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Mock server is running${NC}"
    echo "  Test: curl http://localhost:9000/health"
else
    echo -e "${YELLOW}⚠ Mock server started but health check failed${NC}"
    echo "  Waiting a bit more..."
    sleep 2
    if curl -s http://localhost:9000/health > /dev/null 2>&1; then
        echo -e "${GREEN}✓ Mock server is now healthy${NC}"
    else
        echo -e "${RED}✗ Health check failed${NC}"
        echo "  Check logs: cat /tmp/mock-server.log"
        exit 1
    fi
fi

# Inicia frontend
echo ""
echo -e "${GREEN}Starting frontend on :8080...${NC}"
cd frontend
python3 -m http.server 8080 > /tmp/frontend.log 2>&1 &
FRONTEND_PID=$!
cd ..

sleep 2

# Verifica se o frontend está rodando
if ! kill -0 $FRONTEND_PID 2>/dev/null; then
    echo -e "${RED}✗ Frontend failed to start${NC}"
    echo "  Check logs: cat /tmp/frontend.log"
    cat /tmp/frontend.log
    exit 1
fi

# Testa o frontend
if curl -s http://localhost:8080 > /dev/null 2>&1; then
    echo -e "${GREEN}✓ Frontend is running${NC}"
else
    echo -e "${YELLOW}⚠ Frontend started but test failed${NC}"
fi

# Mostra status
echo ""
echo -e "${BLUE}╔════════════════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║                                                    ║${NC}"
echo -e "${BLUE}║   ${GREEN}✓ Environment Ready!${BLUE}                           ║${NC}"
echo -e "${BLUE}║                                                    ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════════════════╝${NC}"
echo ""
echo -e "${GREEN}Services:${NC}"
echo -e "  ${BLUE}→${NC} Mock Backend:  http://localhost:9000"
echo -e "  ${BLUE}→${NC} Frontend:      http://localhost:8080"
echo ""
echo -e "${GREEN}Endpoints:${NC}"
echo -e "  ${BLUE}→${NC} Health:        http://localhost:9000/health"
echo -e "  ${BLUE}→${NC} Blocks:        http://localhost:9000/blocks"
echo -e "  ${BLUE}→${NC} Mempool:       http://localhost:9000/mempool"
echo -e "  ${BLUE}→${NC} Stats:         http://localhost:9000/stats"
echo ""
echo -e "${GREEN}Test Commands:${NC}"
echo "  curl http://localhost:9000/health | jq"
echo "  curl http://localhost:9000/blocks | jq"
echo ""
echo -e "${YELLOW}Press Ctrl+C to stop both services${NC}"
echo ""

# Monitora logs
tail -f /tmp/mock-server.log /tmp/frontend.log
