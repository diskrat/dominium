#!/bin/bash

# Quick Start - Front & P2P Testing Environment
# =============================================

set -e

GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
RED='\033[0;31m'
NC='\033[0m'

echo -e "${BLUE}"
cat << "EOF"
  ______ _____   ____  _   _ _______   ___     _____  ___  _____  
 |  ____|  __ \ / __ \| \ | |__   __| |__ \   |  __ \|__ \|  __ \ 
 | |__  | |__) | |  | |  \| |  | |       ) |  | |__) |  ) | |__) |
 |  __| |  _  /| |  | | . ` |  | |      / /   |  ___/  / /|  ___/ 
 | |    | | \ \| |__| | |\  |  | |     / /_   | |     / /_| |     
 |_|    |_|  \_\\____/|_| \_|  |_|    |____|  |_|    |____|_|     
                                                                   
EOF
echo -e "${NC}"

echo -e "${GREEN}🚀 Quick Start - Testing Environment${NC}"
echo ""

# Menu
echo "Choose an option:"
echo ""
echo -e "  ${YELLOW}1${NC} - Run unit tests (mocks)"
echo -e "  ${YELLOW}2${NC} - Start mock server only"
echo -e "  ${YELLOW}3${NC} - Start frontend only"
echo -e "  ${YELLOW}4${NC} - Start BOTH (mock server + frontend)"
echo -e "  ${YELLOW}5${NC} - Run tests with coverage"
echo -e "  ${YELLOW}6${NC} - Docker environment (mock backend + frontend)"
echo -e "  ${YELLOW}7${NC} - Show available commands (make help)"
echo -e "  ${YELLOW}0${NC} - Exit"
echo ""
read -p "Enter option: " option

case $option in
    1)
        echo -e "${GREEN}Running unit tests...${NC}"
        cd tests
        go test -v
        ;;
    2)
        echo -e "${GREEN}Starting mock server on :9000...${NC}"
        echo -e "${BLUE}Access: http://localhost:9000${NC}"
        echo -e "${YELLOW}Endpoints:${NC}"
        echo "  - GET  /health"
        echo "  - GET  /blocks"
        echo "  - GET  /mempool"
        echo "  - GET  /stats"
        echo ""
        go run cmd/mock-server/main.go
        ;;
    3)
        echo -e "${GREEN}Starting frontend on :8080...${NC}"
        echo -e "${BLUE}Access: http://localhost:8080${NC}"
        echo ""
        echo -e "${YELLOW}⚠️  Make sure mock server is running on :9000${NC}"
        echo -e "${YELLOW}   Or update app.js with correct backend URL${NC}"
        echo ""
        cd frontend
        python3 -m http.server 8080
        ;;
    4)
        echo -e "${GREEN}Starting mock server + frontend...${NC}"
        echo ""
        echo -e "${BLUE}Starting mock server on :9000 (background)...${NC}"
        go run cmd/mock-server/main.go &
        MOCK_PID=$!
        
        echo -e "${BLUE}Waiting for mock server to start...${NC}"
        sleep 3
        
        echo -e "${BLUE}Starting frontend on :8080...${NC}"
        echo ""
        echo -e "${GREEN}✓ Environment ready!${NC}"
        echo -e "${BLUE}  - Mock Backend: http://localhost:9000${NC}"
        echo -e "${BLUE}  - Frontend: http://localhost:8080${NC}"
        echo ""
        echo -e "${YELLOW}Press Ctrl+C to stop${NC}"
        
        cd frontend
        python3 -m http.server 8080
        
        # Cleanup
        kill $MOCK_PID 2>/dev/null || true
        ;;
    5)
        echo -e "${GREEN}Running tests with coverage...${NC}"
        cd tests
        go test -coverprofile=coverage.out ./...
        go tool cover -html=coverage.out -o coverage.html
        echo ""
        echo -e "${GREEN}✓ Coverage report: tests/coverage.html${NC}"
        
        # Try to open in browser
        if command -v xdg-open &> /dev/null; then
            xdg-open coverage.html
        elif command -v open &> /dev/null; then
            open coverage.html
        fi
        ;;
    6)
        echo -e "${GREEN}Starting Docker environment...${NC}"
        docker-compose -f docker-compose.test.yml up --build
        ;;
    7)
        echo -e "${GREEN}Available Make commands:${NC}"
        echo ""
        make help
        ;;
    0)
        echo -e "${GREEN}Goodbye!${NC}"
        exit 0
        ;;
    *)
        echo -e "${RED}Invalid option${NC}"
        exit 1
        ;;
esac
