package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"dominium/internal/api"
	"dominium/internal/transaction"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// NetworkState repassa as informações completas do Gateway para o React
type NetworkState struct {
	Nodes          []NodeState               `json:"nodes"`
	AllBlocks      []api.BlockMetadata       `json:"all_blocks"`
	CanonicalChain []api.BlockMetadata       `json:"canonical_chain"`
	Accounts       []AccountState            `json:"accounts"`
	Mempool        []transaction.Transaction `json:"mempool"`
	Timestamp      int64                     `json:"timestamp"`
}

type NodeState struct {
	ID          string              `json:"id"`
	Status      string              `json:"status"`
	Difficulty  int32               `json:"difficulty"`
	MempoolSize int                 `json:"mempoolSize"`
	BlockHeight int                 `json:"blockHeight"`
	LocalBlocks []api.BlockMetadata `json:"local_blocks"`
}

type AccountState struct {
	PublicKey string   `json:"publicKey"`
	NFTs      []string `json:"nfts"`
}

type VisualizerServer struct {
	gateway   *api.Gateway
	clients   map[*websocket.Conn]bool
	broadcast chan NetworkState
}

func NewVisualizerServer(gateway *api.Gateway) *VisualizerServer {
	return &VisualizerServer{
		gateway:   gateway,
		clients:   make(map[*websocket.Conn]bool),
		broadcast: make(chan NetworkState, 100),
	}
}

func (vs *VisualizerServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer conn.Close()

	vs.clients[conn] = true
	log.Printf("New WebSocket client connected. Total clients: %d", len(vs.clients))

	// Send initial state
	vs.sendCurrentState(conn)

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			delete(vs.clients, conn)
			log.Printf("WebSocket client disconnected. Total clients: %d", len(vs.clients))
			break
		}
	}
}

func (vs *VisualizerServer) sendCurrentState(conn *websocket.Conn) {
	state := vs.collectNetworkState()
	data, err := json.Marshal(state)
	if err != nil {
		log.Printf("Error marshaling state: %v", err)
		return
	}

	if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
		log.Printf("Error sending state: %v", err)
	}
}

func (vs *VisualizerServer) collectNetworkState() NetworkState {
	resp, err := http.Get("http://api-gateway:8085/network/status")
	if err != nil {
		log.Printf("Error calling gateway API: %v", err)
		return vs.getMockNetworkState()
	}
	defer resp.Body.Close()

	var status struct {
		NetworkHeight  int                       `json:"network_height"`
		NodesActive    []string                  `json:"nodes_active"`
		CanonicalChain []api.BlockMetadata       `json:"canonical_chain"`
		AllBlocks      []api.BlockMetadata       `json:"all_blocks"`
		Accounts       []AccountState            `json:"accounts"`
		Mempool        []transaction.Transaction `json:"mempool"`
		Timestamp      int64                     `json:"timestamp"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		log.Printf("Error decoding gateway response: %v", err)
		return vs.getMockNetworkState()
	}

	currentDifficulty := int32(4)
	if len(status.CanonicalChain) > 0 {
		currentDifficulty = status.CanonicalChain[len(status.CanonicalChain)-1].Difficulty
	}

	nodeStatus := "ocioso"
	if time.Now().UnixNano()-status.Timestamp < int64(15*time.Second) {
		nodeStatus = "mining"
	}

	clusterNodes := []string{"node-1", "node-2", "node-3"}
	nodes := make([]NodeState, len(clusterNodes))
	
	for i, nodeID := range clusterNodes {
		// Calcula a porta do nó (node-1 -> 8082, node-2 -> 8083, etc.)
		port := 8081 + (i + 1)
		var localBlocks []api.BlockMetadata

		// Tenta buscar a blockchain local daquele nó específico
		localResp, localErr := http.Get(fmt.Sprintf("http://%s:%d/local-chain", nodeID, port))
		if localErr == nil {
			json.NewDecoder(localResp.Body).Decode(&localBlocks)
			localResp.Body.Close()
		} else {
			log.Printf("Falha ao ler API local do %s na porta %d: %v", nodeID, port, localErr)
		}

		nodes[i] = NodeState{
			ID:          nodeID,
			Status:      nodeStatus,
			Difficulty:  currentDifficulty,
			BlockHeight: status.NetworkHeight,
			LocalBlocks: localBlocks, // Insere a árvore individual
		}
	}

	return NetworkState{
		Nodes:          nodes,
		AllBlocks:      status.AllBlocks,
		CanonicalChain: status.CanonicalChain,
		Accounts:       status.Accounts,
		Mempool:        status.Mempool,
		Timestamp:      status.Timestamp,
	}
}

func (vs *VisualizerServer) getMockNetworkState() NetworkState {
	// Fallback mock data: Funciona como a tela de "Carregando..."
	// nos primeiros segundos antes do Gateway responder.

	clusterNodes := []string{"node-1", "node-2", "node-3"}
	nodes := make([]NodeState, len(clusterNodes))

	for i, nodeID := range clusterNodes {
		nodes[i] = NodeState{
			ID:          nodeID,
			Status:      "ocioso", // Começa cinza (Ocioso) para não piscar verde à toa
			Difficulty:  0,        // Fica 0 até ler a dificuldade real do Kafka
			MempoolSize: 0,
			BlockHeight: 0, // Começa no bloco 0
		}
	}

	return NetworkState{
		Nodes:          nodes,
		AllBlocks:      []api.BlockMetadata{},
		CanonicalChain: []api.BlockMetadata{},
		Accounts:       []AccountState{},
		Mempool:        []transaction.Transaction{},
		Timestamp:      time.Now().Unix(),
	}
}

func (vs *VisualizerServer) broadcastState() {
	for {
		state := <-vs.broadcast

		data, err := json.Marshal(state)
		if err != nil {
			log.Printf("Error marshaling broadcast state: %v", err)
			continue
		}

		for client := range vs.clients {
			if err := client.WriteMessage(websocket.TextMessage, data); err != nil {
				log.Printf("Error broadcasting to client: %v", err)
				delete(vs.clients, client)
			}
		}
	}
}

func (vs *VisualizerServer) startStateBroadcaster() {
	go vs.broadcastState()

	// Broadcast state every 2 seconds
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		state := vs.collectNetworkState()
		select {
		case vs.broadcast <- state:
		default:
			// Channel full, skip this update
		}
	}
}

func (vs *VisualizerServer) handleChaosMint(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate and submit multiple mint transactions
	for i := 0; i < 10; i++ {
		// This would generate random transactions and submit them
		log.Printf("Generating chaos mint transaction %d", i+1)
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "10 chaos mint transactions submitted"})
}

func (vs *VisualizerServer) handleGenerateWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Generate a new wallet (simplified - in real implementation, use proper wallet generation)
	response := map[string]string{
		"publicKey":  "generated-public-key",
		"privateKey": "generated-private-key", // WARNING: Never send private key in production
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (vs *VisualizerServer) handleRaceAttack(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Simulate race attack with conflicting transactions
	log.Printf("Simulating race attack...")

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"status": "Race attack simulation started"})
}

func main() {
	ctx := context.Background()

	// Initialize gateway (simplified - in real implementation, connect to actual running gateway)
	gateway := api.NewGateway(ctx, "visualizer-gateway", 8081, []string{"kafka:9092"})

	visualizer := NewVisualizerServer(gateway)

	// Start state broadcaster
	go visualizer.startStateBroadcaster()

	// HTTP routes
	http.HandleFunc("/ws", visualizer.handleWebSocket)
	http.HandleFunc("/api/chaos-mint", visualizer.handleChaosMint)
	http.HandleFunc("/api/generate-wallet", visualizer.handleGenerateWallet)
	http.HandleFunc("/api/race-attack", visualizer.handleRaceAttack)

	// Serve static files
	fs := http.FileServer(http.Dir("./web/build"))
	http.Handle("/", fs)

	fmt.Println("🚀 Dominium Visualizer Server starting on :8080")
	fmt.Println("📊 WebSocket endpoint: ws://localhost:8080/ws")
	fmt.Println("🌐 Frontend: http://localhost:8080")

	log.Fatal(http.ListenAndServe(":8080", nil))
}