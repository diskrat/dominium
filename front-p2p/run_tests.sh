#!/bin/bash

# Script para executar testes dos componentes Front & P2P

set -e

echo "🧪 Running Front & P2P Component Tests"
echo "======================================"

cd "$(dirname "$0")"

# Cores para output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Função para executar comandos com feedback
run_test() {
    local name=$1
    local cmd=$2
    
    echo ""
    echo -e "${YELLOW}▶ $name${NC}"
    if eval "$cmd"; then
        echo -e "${GREEN}✓ $name passed${NC}"
        return 0
    else
        echo -e "${RED}✗ $name failed${NC}"
        return 1
    fi
}

# Testes unitários
echo ""
echo "1. Unit Tests"
echo "-------------"
cd tests
run_test "P2P Mock Tests" "go test -v -run TestMockRabbitMQ"
run_test "Blockchain Mock Tests" "go test -v -run TestMockBlockchain"
cd ..

# Benchmarks
echo ""
echo "2. Benchmarks"
echo "-------------"
cd tests
run_test "P2P Benchmarks" "go test -bench=BenchmarkMockRabbitMQ -benchmem"
run_test "Blockchain Benchmarks" "go test -bench=BenchmarkMockBlockchain -benchmem"
cd ..

# Cobertura de testes
echo ""
echo "3. Test Coverage"
echo "----------------"
cd tests
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
echo -e "${GREEN}✓ Coverage report generated: tests/coverage.html${NC}"
cd ..

echo ""
echo "======================================"
echo -e "${GREEN}✓ All tests completed!${NC}"
echo ""
echo "📊 To start mock server:"
echo "   go run cmd/mock-server/main.go"
echo ""
echo "🌐 To start frontend:"
echo "   cd frontend && python3 -m http.server 8080"
