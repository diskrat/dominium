.PHONY: help test test-unit test-coverage test-bench mock-server frontend clean docker-up docker-down

# Variáveis
TESTS_DIR=tests
MOCKS_DIR=mocks
FRONTEND_DIR=frontend

# Cores
GREEN=\033[0;32m
YELLOW=\033[1;33m
NC=\033[0m

help: ## Mostra este help
	@echo "$(GREEN)Front & P2P - Makefile Commands$(NC)"
	@echo ""
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "  $(YELLOW)%-20s$(NC) %s\n", $$1, $$2}'
	@echo ""

test: ## Executa todos os testes
	@echo "$(GREEN)Running all tests...$(NC)"
	@cd $(TESTS_DIR) && go test -v ./...

test-unit: ## Executa testes unitários
	@echo "$(GREEN)Running unit tests...$(NC)"
	@cd $(TESTS_DIR) && go test -v -short ./...

test-coverage: ## Gera relatório de cobertura
	@echo "$(GREEN)Generating coverage report...$(NC)"
	@cd $(TESTS_DIR) && go test -coverprofile=coverage.out ./...
	@cd $(TESTS_DIR) && go tool cover -html=coverage.out -o coverage.html
	@echo "$(GREEN)✓ Coverage report: $(TESTS_DIR)/coverage.html$(NC)"

test-bench: ## Executa benchmarks
	@echo "$(GREEN)Running benchmarks...$(NC)"
	@cd $(TESTS_DIR) && go test -bench=. -benchmem ./...

test-p2p: ## Testes apenas do P2P mock
	@echo "$(GREEN)Running P2P tests...$(NC)"
	@cd $(TESTS_DIR) && go test -v -run TestMockRabbitMQ

test-blockchain: ## Testes apenas do blockchain mock
	@echo "$(GREEN)Running Blockchain tests...$(NC)"
	@cd $(TESTS_DIR) && go test -v -run TestMockBlockchain

mock-server: ## Inicia o servidor mock
	@echo "$(GREEN)Starting mock server on :9000...$(NC)"
	@go run cmd/mock-server/main.go

frontend: ## Inicia o frontend (Python HTTP server)
	@echo "$(GREEN)Starting frontend on :8080...$(NC)"
	@cd $(FRONTEND_DIR) && python3 -m http.server 8080

docker-up: ## Inicia containers Docker de teste
	@echo "$(GREEN)Starting Docker containers...$(NC)"
	@docker-compose -f docker-compose.test.yml up --build

docker-down: ## Para containers Docker
	@echo "$(GREEN)Stopping Docker containers...$(NC)"
	@docker-compose -f docker-compose.test.yml down

docker-logs: ## Mostra logs dos containers
	@docker-compose -f docker-compose.test.yml logs -f

clean: ## Remove arquivos temporários
	@echo "$(GREEN)Cleaning up...$(NC)"
	@find . -name "coverage.out" -delete
	@find . -name "coverage.html" -delete
	@echo "$(GREEN)✓ Cleanup complete$(NC)"

lint: ## Executa linter Go
	@echo "$(GREEN)Running Go linter...$(NC)"
	@cd $(MOCKS_DIR) && go fmt ./...
	@cd $(TESTS_DIR) && go fmt ./...

init: ## Inicializa módulo Go
	@echo "$(GREEN)Initializing Go module...$(NC)"
	@go mod init front-p2p || true
	@go mod tidy

structure: ## Mostra estrutura do projeto
	@echo "$(GREEN)Project structure:$(NC)"
	@find . -type f -name "*.go" -o -name "*.js" -o -name "*.html" | grep -v node_modules | sort

stats: ## Estatísticas do código
	@echo "$(GREEN)Code statistics:$(NC)"
	@echo "Go files: $$(find . -name '*.go' | wc -l)"
	@echo "JavaScript files: $$(find . -name '*.js' | wc -l)"
	@echo "HTML files: $$(find . -name '*.html' | wc -l)"
	@echo "Total lines of Go: $$(find . -name '*.go' -exec cat {} \; | wc -l)"
	@echo "Total lines of JS: $$(find . -name '*.js' -exec cat {} \; | wc -l)"

demo: ## Demonstração completa (mock server + frontend)
	@echo "$(GREEN)Starting demo environment...$(NC)"
	@echo "$(YELLOW)1. Starting mock server...$(NC)"
	@cd $(TESTS_DIR) && go run mock_server.go &
	@sleep 2
	@echo "$(YELLOW)2. Starting frontend...$(NC)"
	@echo "$(GREEN)✓ Demo ready!$(NC)"
	@echo "   - Backend: http://localhost:9000"
	@echo "   - Frontend: http://localhost:8080"
	@cd $(FRONTEND_DIR) && python3 -m http.server 8080

.DEFAULT_GOAL := help
