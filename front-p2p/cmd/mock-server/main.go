package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"front-p2p/mocks"
)

var (
	bc *mocks.MockBlockchain
	mq *mocks.MockRabbitMQ
)

func enableCORS(w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
}

func handleBlocks(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(bc.GetBlocks())
}

func handleMempool(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	// Retorna array direto para compatibilidade com frontend
	json.NewEncoder(w).Encode(bc.GetMempool())
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"blocks":     bc.GetHeight(),
		"mempool":    len(bc.GetMempool()),
		"connected":  mq.IsConnected(),
		"timestamp":  time.Now().Unix(),
		"node_id":    "mock-server",
		"version":    "1.0.0-mock",
	})
}

func handleStats(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"node_id":            "mock-server",
		"total_blocks":       bc.GetHeight(),
		"total_transactions": len(bc.GetTransactions()),
		"mempool_size":       len(bc.GetMempool()),
		"messages_sent":      len(mq.GetPublishedMessages()),
		"p2p_connected":      mq.IsConnected(),
		"uptime":             time.Since(time.Now()).Seconds(), // placeholder
	})
}

func handleAddBlock(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var block mocks.MockBlock
	if err := json.NewDecoder(r.Body).Decode(&block); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := bc.AddBlock(&block); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"block":   block,
	})
}

func handleAddTransaction(w http.ResponseWriter, r *http.Request) {
	enableCORS(w)
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var tx mocks.MockTransaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if err := bc.AddTransaction(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":     true,
		"transaction": tx,
	})
}

func simulateBlockGeneration() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		lastBlock := bc.GetLastBlock()
		if lastBlock == nil {
			continue
		}

		newBlock := &mocks.MockBlock{
			Index:    lastBlock.Index + 1,
			Hash:     fmt.Sprintf("mock_hash_%d_%d", lastBlock.Index+1, time.Now().Unix()),
			PrevHash: lastBlock.Hash,
			Data:     fmt.Sprintf("Mock block data %d", lastBlock.Index+1),
			Nonce:    time.Now().Unix(),
		}

		if err := bc.AddBlock(newBlock); err != nil {
			log.Printf("Error adding mock block: %v", err)
		} else {
			log.Printf("✅ Generated mock block #%d", newBlock.Index)
			
			// Simula publicação via P2P
			mq.Publish("new_block", map[string]interface{}{
				"block": newBlock,
				"from":  "mock-server",
			})
		}
	}
}

func initMockData() {
	// Adiciona alguns blocos iniciais
	for i := 1; i <= 3; i++ {
		lastBlock := bc.GetLastBlock()
		block := &mocks.MockBlock{
			Index:    i,
			Hash:     fmt.Sprintf("mock_hash_%d", i),
			PrevHash: lastBlock.Hash,
			Data:     fmt.Sprintf("Initial mock block %d", i),
			Nonce:    int64(i * 1000),
		}
		bc.AddBlock(block)
	}

	// Adiciona algumas transações ao mempool
	for i := 1; i <= 2; i++ {
		tx := &mocks.MockTransaction{
			ID:        fmt.Sprintf("mock_tx_%d", i),
			Type:      "REGISTRO",
			Sender:    "mock_sender",
			Recipient: "mock_recipient",
			Data: map[string]interface{}{
				"property_id": fmt.Sprintf("PROP_%d", i),
				"area":        100.5 * float64(i),
			},
		}
		bc.AddTransaction(tx)
	}

	log.Println("✅ Mock data initialized")
	log.Printf("   - Blocks: %d", bc.GetHeight())
	log.Printf("   - Mempool: %d transactions", len(bc.GetMempool()))
}

func main() {
	// Inicializa mocks
	bc = mocks.NewMockBlockchain()
	mq = mocks.NewMockRabbitMQ("mock-server")
	mq.Connect()

	// Inicializa dados mockados
	initMockData()

	// Inicia geração automática de blocos
	go simulateBlockGeneration()

	// Configura rotas
	http.HandleFunc("/blocks", handleBlocks)
	http.HandleFunc("/mempool", handleMempool)
	http.HandleFunc("/health", handleHealth)
	http.HandleFunc("/stats", handleStats)
	http.HandleFunc("/api/blocks", handleAddBlock)
	http.HandleFunc("/api/transactions", handleAddTransaction)

	port := ":9000"
	log.Printf("🚀 Mock Server started on http://localhost%s", port)
	log.Println("📊 Endpoints:")
	log.Println("   - GET  /health")
	log.Println("   - GET  /blocks")
	log.Println("   - GET  /mempool")
	log.Println("   - GET  /stats")
	log.Println("   - POST /api/blocks")
	log.Println("   - POST /api/transactions")
	log.Println("")
	log.Println("🔄 Auto-generating blocks every 10 seconds...")

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal(err)
	}
}
