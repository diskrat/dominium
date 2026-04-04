package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"dominium/internal/miner"
	"dominium/internal/network"
	"dominium/internal/transaction"

	"github.com/segmentio/kafka-go"
)

// BlockMetadata armazena informações sobre um bloco para exibição.
type BlockMetadata struct {
	Hash       string `json:"hash"`
	MinerID    string `json:"miner_id"`
	Height     uint64 `json:"height"`
	TxCount    int    `json:"tx_count"`
	Timestamp  int64  `json:"timestamp"`
	Difficulty int32  `json:"difficulty"`
}

// NetworkStatus representa o estado da rede Dominium.
type NetworkStatus struct {
	NetworkHeight  int             `json:"network_height"`
	NodesActive    []string        `json:"nodes_active"`
	CanonicalChain []BlockMetadata `json:"canonical_chain"`
	Timestamp      int64           `json:"timestamp"`
}

// TransactionRequest é o payload de entrada para POST /transactions.
type TransactionRequest struct {
	Type      string `json:"type"`      // "mint" ou "transfer"
	Recipient string `json:"recipient"` // chave publica do destinatario
	NFTID     string `json:"nft_id"`    // ID do NFT (opcional para mint)
	Sender    string `json:"sender"`    // chave publica do remetente (requerido para transfer)
}

// TransactionResponse é a resposta de POST /transactions.
type TransactionResponse struct {
	TxID   string `json:"tx_id"`
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
}

type Gateway struct {
	ctx              context.Context
	cancel           context.CancelFunc
	id               string
	port             int
	brokers          []string
	server           *http.Server
	txTransport      *network.KafkaTransactionTransport
	blockTransport   *network.KafkaBlockTransport
	generator        *transaction.Generator
	state            *transaction.AccountState
	blockchain       *miner.Blockchain
	blockMetadata    map[string]*BlockMetadata
	activeNodes      map[string]bool
	mu               sync.RWMutex
	networkSyncReady bool
	adminPubKey      string
}

// NewGateway cria uma nova instância do API Gateway.
func NewGateway(parent context.Context, id string, port int, brokers []string) *Gateway {
	ctx, cancel := context.WithCancel(parent)
	if len(brokers) == 0 {
		brokers = []string{"localhost:9092"}
	}

	gen := transaction.NewGenerator(time.Now().UnixNano())
	state := transaction.NewAccountState()

	return &Gateway{
		ctx:           ctx,
		cancel:        cancel,
		id:            strings.TrimSpace(id),
		port:          port,
		brokers:       brokers,
		generator:     gen,
		state:         state,
		blockchain:    miner.NewBlockchain(),
		blockMetadata: make(map[string]*BlockMetadata),
		activeNodes:   make(map[string]bool),
	}
}

// SetAdminIdentity configura a identidade admin para assinar transações de mint.
func (g *Gateway) SetAdminIdentity(identity *transaction.WalletIdentity) error {
	if identity == nil {
		return errors.New("identidade admin nula")
	}
	g.adminPubKey = identity.PublicKey
	return g.generator.SetAdmin(identity)
}

// Run inicia o servidor HTTP e a sincronização de rede.
func (g *Gateway) Run() error {
	if g.id == "" {
		return errors.New("-id é obrigatório")
	}
	if g.port <= 0 || g.port > 65535 {
		return errors.New("porta invalida")
	}

	// Inicializa os transportes
	g.txTransport = network.NewKafkaTransactionTransport(g.ctx, g.brokers, g.id)
	g.blockTransport = network.NewKafkaBlockTransport(g.ctx, g.brokers, g.id)

	// Sincronização retroativa de blocos
	if err := g.syncBlockchainHistory(); err != nil {
		return err
	}

	// Escuta por novos blocos
	if err := g.subscribeBlocks(); err != nil {
		return err
	}

	// Configure o servidor HTTP
	mux := http.NewServeMux()
	mux.HandleFunc("POST /transactions", g.handlePostTransaction)
	mux.HandleFunc("GET /network/status", g.handleGetNetworkStatus)
	mux.HandleFunc("GET /health", g.handleHealth)

	g.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", g.port),
		Handler: mux,
	}

	log.Printf("[%s] API Gateway iniciado em http://0.0.0.0:%d", g.id, g.port)

	// Inicia o servidor em uma goroutine
	go func() {
		if err := g.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("[%s] erro ao servir HTTP: %v", g.id, err)
		}
	}()

	<-g.ctx.Done()
	return g.shutdown()
}

func (g *Gateway) syncBlockchainHistory() error {
	log.Printf("[%s] iniciando sincronizacao retroativa de blocos", g.id)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        g.brokers,
		Topic:          network.TopicBlocks,
		GroupID:        g.id + "-sync",
		StartOffset:    kafka.FirstOffset,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: 1 * time.Second,
	})
	defer reader.Close()

	ctx, cancel := context.WithTimeout(g.ctx, 30*time.Second)
	defer cancel()

	blockCount := 0
	for {
		m, err := reader.FetchMessage(ctx)
		if err != nil {
			if errors.Is(err, context.DeadlineExceeded) {
				break
			}
			log.Printf("[%s] erro ao sincronizar blocos: %v", g.id, err)
			break
		}

		var block miner.Block
		if err := json.Unmarshal(m.Value, &block); err != nil {
			log.Printf("[%s] bloco invalido durante sync: %v", g.id, err)
			continue
		}

		if err := g.processBlock(&block, m.Key); err != nil {
			log.Printf("[%s] falha ao processar bloco durante sync: %v", g.id, err)
		}
		blockCount++

		if err := reader.CommitMessages(g.ctx, m); err != nil {
			log.Printf("[%s] falha ao commitar offset: %v", g.id, err)
		}
	}

	g.mu.Lock()
	g.networkSyncReady = true
	g.mu.Unlock()

	log.Printf("[%s] sincronizacao concluida com %d blocos", g.id, blockCount)
	return nil
}

func (g *Gateway) subscribeBlocks() error {
	return g.blockTransport.Subscribe(func(block *miner.Block) error {
		// Extrai miner ID da chave da mensagem (será necessário melhorar isso)
		return g.processBlock(block, nil)
	})
}

func (g *Gateway) processBlock(block *miner.Block, minerKeyBytes []byte) error {
	if block == nil {
		return errors.New("bloco nulo")
	}

	minerID := ""
	if minerKeyBytes != nil {
		minerID = string(minerKeyBytes)
	}

	// Adiciona o bloco à blockchain
	if err := g.blockchain.AddBlock(*block); err != nil {
		if strings.Contains(err.Error(), "ja existe") {
			return nil
		}
		return err
	}

	// Atualiza metadados
	g.mu.Lock()
	defer g.mu.Unlock()

	hashHex := hex.EncodeToString(block.Hash)
	g.blockMetadata[hashHex] = &BlockMetadata{
		Hash:       hashHex,
		MinerID:    minerID,
		TxCount:    len(block.Transactions),
		Timestamp:  block.Timestamp,
		Difficulty: block.Nbits,
	}

	if minerID != "" {
		g.activeNodes[minerID] = true
	}

	// Atualiza AccountState após processar transações
	for _, tx := range block.Transactions {
		if err := g.state.EnsureAccount(tx.PublKey); err != nil {
			log.Printf("[%s] erro ao garantir conta: %v", g.id, err)
		}
		if err := g.state.EnsureAccount(tx.Recipient); err != nil {
			log.Printf("[%s] erro ao garantir recipient: %v", g.id, err)
		}

		// Valida regras de consenso antes de executar
		if err := tx.ValidateConsensusRules(g.state, g.adminPubKey); err != nil {
			log.Printf("[%s] bloco contem transacao que viola consenso: %v", g.id, err)
			continue // Pula transação inválida mas continua processando o bloco
		}

		if err := tx.Execute(g.state); err != nil {
			log.Printf("[%s] erro ao executar transacao: %v", g.id, err)
		}
	}

	return nil
}

func (g *Gateway) handlePostTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}

	g.mu.RLock()
	syncReady := g.networkSyncReady
	g.mu.RUnlock()

	if !syncReady {
		resp := TransactionResponse{Error: "rede ainda nao sincronizada"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(resp)
		return
	}

	var req TransactionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		resp := TransactionResponse{Error: fmt.Sprintf("corpo invalido: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	var tx *transaction.Transaction
	var err error

	switch strings.ToLower(req.Type) {
	case "mint":
		if req.Recipient == "" {
			resp := TransactionResponse{Error: "recipient obrigatorio para mint"}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}

		recIdentity := &transaction.WalletIdentity{PublicKey: req.Recipient}
		tx, err = g.generator.MintTx(recIdentity, req.NFTID)

	case "transfer":
		if req.Sender == "" || req.Recipient == "" || req.NFTID == "" {
			resp := TransactionResponse{Error: "sender, recipient e nft_id obrigatorios para transfer"}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(resp)
			return
		}

		senderIdentity := &transaction.WalletIdentity{PublicKey: req.Sender}
		recipientIdentity := &transaction.WalletIdentity{PublicKey: req.Recipient}
		tx, err = g.generator.TransferTx(senderIdentity, recipientIdentity, req.NFTID)

	default:
		resp := TransactionResponse{Error: "tipo de transacao invalido: mint ou transfer"}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err != nil {
		resp := TransactionResponse{Error: fmt.Sprintf("erro ao criar transacao: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := tx.SetID(); err != nil {
		resp := TransactionResponse{Error: fmt.Sprintf("erro ao calcular ID: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Valida antes de publicar
	if err := tx.Validate(g.state); err != nil {
		resp := TransactionResponse{TxID: tx.ID, Error: fmt.Sprintf("validacao falhou: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Valida regras de consenso específicas
	if err := tx.ValidateConsensusRules(g.state, g.adminPubKey); err != nil {
		resp := TransactionResponse{TxID: tx.ID, Error: fmt.Sprintf("consenso violado: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Publica a transação
	if err := g.txTransport.Publish(tx); err != nil {
		resp := TransactionResponse{TxID: tx.ID, Error: fmt.Sprintf("falha ao publicar: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	resp := TransactionResponse{
		TxID:   tx.ID,
		Status: "Published",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (g *Gateway) handleGetNetworkStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}

	g.mu.RLock()
	defer g.mu.RUnlock()

	chain := g.blockchain.GetCanonicalChain()
	canonicalChain := make([]BlockMetadata, 0, len(chain))

	for i, block := range chain {
		hashHex := hex.EncodeToString(block.Hash)
		metadata := BlockMetadata{
			Hash:       hashHex,
			Height:     uint64(i),
			TxCount:    len(block.Transactions),
			Timestamp:  block.Timestamp,
			Difficulty: block.Nbits,
		}

		if stored, ok := g.blockMetadata[hashHex]; ok {
			metadata.MinerID = stored.MinerID
		}

		canonicalChain = append(canonicalChain, metadata)
	}

	activeNodes := make([]string, 0, len(g.activeNodes))
	for node := range g.activeNodes {
		activeNodes = append(activeNodes, node)
	}

	status := NetworkStatus{
		NetworkHeight:  len(chain),
		NodesActive:    activeNodes,
		CanonicalChain: canonicalChain,
		Timestamp:      time.Now().UnixNano(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}

	g.mu.RLock()
	syncReady := g.networkSyncReady
	g.mu.RUnlock()

	health := map[string]interface{}{
		"status": "ok",
		"synced": syncReady,
	}

	w.Header().Set("Content-Type", "application/json")
	if syncReady {
		w.WriteHeader(http.StatusOK)
	} else {
		w.WriteHeader(http.StatusServiceUnavailable)
	}
	json.NewEncoder(w).Encode(health)
}

func (g *Gateway) shutdown() error {
	log.Printf("[%s] encerrando API Gateway...", g.id)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if g.server != nil {
		if err := g.server.Shutdown(ctx); err != nil {
			log.Printf("[%s] erro ao desligar servidor HTTP: %v", g.id, err)
		}
	}

	if g.txTransport != nil {
		g.txTransport.Stop()
	}
	if g.blockTransport != nil {
		g.blockTransport.Stop()
	}

	g.cancel()
	return nil
}
