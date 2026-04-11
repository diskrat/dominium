package api

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
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
	ParentHash string `json:"hash_of_previous"`
	MinerID    string `json:"miner_id"`
	Height     uint64 `json:"height"`
	TxCount    int    `json:"tx_count"`
	Timestamp  int64  `json:"timestamp"`
	Difficulty int32  `json:"difficulty"`
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

type accountResp struct {
	PublicKey string   `json:"publicKey"`
	NFTs      []string `json:"nfts"`
}

type networkStatusResponse struct {
	NetworkHeight    int                        `json:"network_height"`
	NodesActive      []string                   `json:"nodes_active"`
	CanonicalChain   []BlockMetadata            `json:"canonical_chain"`
	AllBlocks        []BlockMetadata            `json:"all_blocks"`
	Accounts         []accountResp              `json:"accounts"`
	Mempool          []*transaction.Transaction `json:"mempool"`
	CanonicalTxCount int                        `json:"canonical_tx_count"`
	AllBlocksTxCount int                        `json:"all_blocks_tx_count"`
	OrphanTxCount    int                        `json:"orphan_tx_count"`
	DiscardedTxCount int                        `json:"discarded_tx_count"`
	MempoolTxCount   int                        `json:"mempool_tx_count"`
	Timestamp        int64                      `json:"timestamp"`
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
	mempool          map[string]*transaction.Transaction
	stateMu          sync.RWMutex
	metaMu           sync.RWMutex
	genMu            sync.Mutex
	networkSyncReady bool
	adminPubKey      string
}

// NewGateway cria uma nova instância do API Gateway.
func NewGateway(parent context.Context, id string, port int, brokers []string) *Gateway {
	ctx, cancel := context.WithCancel(parent)
	if len(brokers) == 0 {
		brokers = []string{"kafka:9092"}
	}

	gen := transaction.NewGenerator(time.Now().UnixNano())
	state := transaction.NewAccountState()

	return &Gateway{
		ctx:        ctx,
		cancel:     cancel,
		id:         strings.TrimSpace(id),
		port:       port,
		brokers:    brokers,
		generator:  gen,
		state:      state,
		blockchain: miner.NewBlockchain(),
		mempool:    make(map[string]*transaction.Transaction),
	}
}

// SetAdminIdentity configura a identidade admin para assinar transações de mint.
func (g *Gateway) SetAdminIdentity(identity *transaction.WalletIdentity) error {
	if identity == nil {
		return errors.New("identidade admin nula")
	}
	g.adminPubKey = identity.PublicKey
	g.genMu.Lock()
	defer g.genMu.Unlock()
	return g.generator.SetAdmin(identity)
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
		w.Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
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
	mux.HandleFunc("POST /attacks/double-spend", g.handleDoubleSpendAttack)
	mux.HandleFunc("GET /wallet/generate", g.handleGenerateWallet)
	mux.HandleFunc("POST /network/difficulty", g.handleUpdateDifficulty)

	g.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", g.port),
		Handler: corsMiddleware(mux),
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

		if err := g.processBlock(&block); err != nil {
			log.Printf("[%s] falha ao processar bloco durante sync: %v", g.id, err)
		}
		blockCount++

		if err := reader.CommitMessages(g.ctx, m); err != nil {
			log.Printf("[%s] falha ao commitar offset: %v", g.id, err)
		}
	}

	g.metaMu.Lock()
	g.networkSyncReady = true
	g.metaMu.Unlock()

	log.Printf("[%s] sincronizacao concluida com %d blocos", g.id, blockCount)
	return nil
}

func (g *Gateway) subscribeBlocks() error {
	return g.blockTransport.Subscribe(func(block *miner.Block) error {
		return g.processBlock(block)
	})
}

func (g *Gateway) processBlock(block *miner.Block) error {
	if block == nil {
		return errors.New("bloco nulo")
	}

	// Adiciona o bloco à blockchain
	result, err := g.blockchain.AddBlock(*block)
	if err != nil {
		if errors.Is(err, miner.ErrDuplicateBlock) {
			return nil
		}
		if !errors.Is(err, miner.ErrOrphanBlock) {
			return err
		}
	}

	if err != nil || !result.TipUpdated {
		return nil
	}

	if err := g.rebuildStateFromCanonicalChain(); err != nil {
		return err
	}

	return nil
}

func (g *Gateway) rebuildStateFromCanonicalChain() error {
	chain := g.blockchain.GetCanonicalChain()
	state := transaction.NewAccountState()
	chainTxs := make(map[string]bool)

	for _, block := range chain {
		for _, tx := range block.Transactions {
			chainTxs[tx.ID] = true
			if err := state.EnsureAccount(tx.PublKey); err != nil {
				return err
			}
			if err := state.EnsureAccount(tx.Recipient); err != nil {
				return err
			}
			if err := tx.Execute(state); err != nil {
				return err
			}
		}
	}

	g.stateMu.RLock()
	pending := make(map[string]*transaction.Transaction, len(g.mempool))
	for id, tx := range g.mempool {
		pending[id] = tx
	}
	g.stateMu.RUnlock()

	newMempool := make(map[string]*transaction.Transaction)
	for id, tx := range pending {
		if chainTxs[id] {
			continue
		}
		if err := state.EnsureAccount(tx.PublKey); err != nil {
			continue
		}
		if err := state.EnsureAccount(tx.Recipient); err != nil {
			continue
		}
		if err := tx.Validate(state); err != nil {
			continue
		}
		if err := tx.ValidateConsensusRules(state, g.adminPubKey); err != nil {
			continue
		}
		newMempool[id] = tx
	}

	g.stateMu.Lock()
	g.state = state
	g.mempool = newMempool
	g.stateMu.Unlock()

	return nil
}

func (g *Gateway) snapshotAccounts() []accountResp {
	g.stateMu.RLock()
	allAccounts := g.state.GetAllAccounts()
	accountsResponse := make([]accountResp, 0, len(allAccounts))
	for pubKey, data := range allAccounts {
		accountsResponse = append(accountsResponse, accountResp{
			PublicKey: pubKey,
			NFTs:      data.GetNFTsList(),
		})
	}
	g.stateMu.RUnlock()

	sort.Slice(accountsResponse, func(i, j int) bool {
		return accountsResponse[i].PublicKey < accountsResponse[j].PublicKey
	})
	return accountsResponse
}

func (g *Gateway) snapshotMempool() []*transaction.Transaction {
	g.stateMu.RLock()
	mempoolResponse := make([]*transaction.Transaction, 0, len(g.mempool))
	for _, tx := range g.mempool {
		mempoolResponse = append(mempoolResponse, tx)
	}
	g.stateMu.RUnlock()

	sort.Slice(mempoolResponse, func(i, j int) bool {
		return mempoolResponse[i].Timestamp > mempoolResponse[j].Timestamp
	})
	return mempoolResponse
}

func buildBlockMetadata(hashHex string, block miner.Block, height uint64) BlockMetadata {
	return BlockMetadata{
		Hash:       hashHex,
		ParentHash: hex.EncodeToString(block.HashOfPrevious),
		MinerID:    block.Miner,
		Height:     height,
		TxCount:    len(block.Transactions),
		Timestamp:  block.Timestamp,
		Difficulty: block.Nbits,
	}
}

func (g *Gateway) buildNetworkStatus() networkStatusResponse {
	chain := g.blockchain.GetCanonicalChain()
	allNodesSnapshot := g.blockchain.GetAllBlocks()
	orphansSnapshot := g.blockchain.GetOrphansSnapshot()
	discardedSnapshot := g.blockchain.GetDiscardedSnapshot()

	canonicalChain := make([]BlockMetadata, 0, len(chain))
	canonicalTxCount := 0
	for _, block := range chain {
		hashHex := hex.EncodeToString(block.Hash)
		height := uint64(0)
		if node, ok := allNodesSnapshot[hashHex]; ok {
			height = node.Height
		}
		canonicalTxCount += len(block.Transactions)
		canonicalChain = append(canonicalChain, buildBlockMetadata(hashHex, block, height))
	}

	allBlocks := make([]BlockMetadata, 0, len(allNodesSnapshot)+len(orphansSnapshot)+len(discardedSnapshot))
	allBlocksTxCount := 0
	orphanTxCount := 0
	discardedTxCount := 0
	for hashHex, node := range allNodesSnapshot {
		block := node.Block
		allBlocksTxCount += len(block.Transactions)
		allBlocks = append(allBlocks, buildBlockMetadata(hashHex, block, node.Height))
	}
	for hashHex, block := range orphansSnapshot {
		orphanTxCount += len(block.Transactions)
		allBlocks = append(allBlocks, buildBlockMetadata(hashHex, block, 0))
	}
	for hashHex, block := range discardedSnapshot {
		discardedTxCount += len(block.Transactions)
		allBlocks = append(allBlocks, buildBlockMetadata(hashHex, block, 0))
	}

	sort.Slice(allBlocks, func(i, j int) bool {
		return allBlocks[i].Timestamp < allBlocks[j].Timestamp
	})

	activeNodeSet := make(map[string]bool)
	for _, meta := range allBlocks {
		if meta.MinerID != "" {
			activeNodeSet[meta.MinerID] = true
		}
	}
	activeNodes := make([]string, 0, len(activeNodeSet))
	for node := range activeNodeSet {
		activeNodes = append(activeNodes, node)
	}

	accountsResponse := g.snapshotAccounts()
	mempoolResponse := g.snapshotMempool()

	return networkStatusResponse{
		NetworkHeight:    len(chain),
		NodesActive:      activeNodes,
		CanonicalChain:   canonicalChain,
		AllBlocks:        allBlocks,
		Accounts:         accountsResponse,
		Mempool:          mempoolResponse,
		CanonicalTxCount: canonicalTxCount,
		AllBlocksTxCount: allBlocksTxCount,
		OrphanTxCount:    orphanTxCount,
		DiscardedTxCount: discardedTxCount,
		MempoolTxCount:   len(mempoolResponse),
		Timestamp:        time.Now().UnixNano(),
	}
}

func (g *Gateway) handlePostTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}

	g.metaMu.RLock()
	syncReady := g.networkSyncReady
	g.metaMu.RUnlock()

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
		g.genMu.Lock()
		tx, err = g.generator.MintTx(recIdentity, req.NFTID)
		g.genMu.Unlock()
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
		g.genMu.Lock()
		tx, err = g.generator.TransferTx(senderIdentity, recipientIdentity, req.NFTID)
		g.genMu.Unlock()
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

	g.stateMu.RLock()
	if err := tx.Validate(g.state); err != nil {
		g.stateMu.RUnlock()
		resp := TransactionResponse{TxID: tx.ID, Error: fmt.Sprintf("validacao falhou: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}

	if err := tx.ValidateConsensusRules(g.state, g.adminPubKey); err != nil {
		g.stateMu.RUnlock()
		resp := TransactionResponse{TxID: tx.ID, Error: fmt.Sprintf("consenso violado: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(resp)
		return
	}
	g.stateMu.RUnlock()

	// Publica no Kafka
	if err := g.txTransport.Publish(tx); err != nil {
		resp := TransactionResponse{TxID: tx.ID, Error: fmt.Sprintf("falha ao publicar: %v", err)}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// === NOVIDADE: Adiciona na mempool local para o Visualizer ver ===
	g.stateMu.Lock()
	g.mempool[tx.ID] = tx
	g.stateMu.Unlock()

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

	status := g.buildNetworkStatus()

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(status)
}

func (g *Gateway) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}

	g.metaMu.RLock()
	syncReady := g.networkSyncReady
	g.metaMu.RUnlock()

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

func (g *Gateway) handleDoubleSpendAttack(w http.ResponseWriter, r *http.Request) {
	g.stateMu.Lock()
	defer g.stateMu.Unlock()

	// 1. Localiza uma conta que tenha pelo menos um NFT
	allAccounts := g.state.GetAllAccounts()
	var targetPubKey string
	var nftID string

	// Busca a primeira conta que possua pelo menos 1 NFT
	for pubKey, acc := range allAccounts {
		if acc != nil && len(acc.NFTs) > 0 {
			targetPubKey = pubKey

			// Como NFTs é um map[string]bool, pegamos a primeira chave do mapa
			for id := range acc.NFTs {
				nftID = id
				break // Pegamos apenas um NFT
			}
			break
		}
	}

	if targetPubKey == "" || nftID == "" {
		http.Error(w, "Nenhuma conta com NFTs disponível para ataque. Rode o Chaos Mint primeiro.", http.StatusBadRequest)
		return
	}

	// 2. Cria duas transações conflitantes (Tipo 1 = Transferência)
	tx1 := transaction.Transaction{
		ID:        "ATTACK_A_" + nftID,
		PublKey:   targetPubKey,
		Recipient: "DESTINATARIO_LEGITIMO",
		NFTID:     nftID,
		Type:      1,
		Timestamp: time.Now().UnixNano(),
	}

	tx2 := transaction.Transaction{
		ID:        "ATTACK_B_" + nftID,
		PublKey:   targetPubKey,
		Recipient: "DESTINATARIO_HACKER",
		NFTID:     nftID,
		Type:      1,
		Timestamp: time.Now().UnixNano() + 1,
	}

	// 3. Dispara as duas para o Kafka passando os ponteiros
	g.txTransport.Publish(&tx1)
	g.txTransport.Publish(&tx2)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status": "Ataque disparado",
		"nft":    nftID,
		"tx_a":   tx1.ID,
		"tx_b":   tx2.ID,
	})
}

func (g *Gateway) handleGenerateWallet(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}

	// Usa as funções reais de criptografia da sua blockchain (ECDSA P-256)
	privKey, pubKey, err := transaction.GenerateKeyPair()
	if err != nil {
		http.Error(w, "Erro ao gerar par de chaves", http.StatusInternalServerError)
		return
	}

	encodedPriv, err := transaction.EncodePrivateKey(privKey)
	if err != nil {
		http.Error(w, "Erro ao codificar chave privada", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"public_key":  transaction.EncodePublicKey(pubKey),
		"private_key": encodedPriv,
	})
}

func (g *Gateway) handleUpdateDifficulty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "metodo nao permitido", http.StatusMethodNotAllowed)
		return
	}

	var req map[string]int32
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "payload invalido", http.StatusBadRequest)
		return
	}

	newDifficulty := req["difficulty"]
	if newDifficulty < 4 || newDifficulty > 64 {
		http.Error(w, "dificuldade deve estar entre 4 e 64 bits", http.StatusBadRequest)
		return
	}

	// 1. Cria a mensagem JSON de configuração
	msgData := map[string]interface{}{
		"type":      "DIFFICULTY_UPDATE",
		"value":     newDifficulty,
		"timestamp": time.Now().UnixNano(),
	}
	payload, _ := json.Marshal(msgData)

	// 2. Configura um Escritor rápido para o tópico "network-config"
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  g.brokers,
		Topic:    "network-config",
		Balancer: &kafka.LeastBytes{},
	})
	defer writer.Close()

	// 3. Dispara a mensagem para a rede
	err := writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte("config"),
		Value: payload,
	})

	if err != nil {
		log.Printf("[%s] ERRO ao propagar dificuldade: %v", g.id, err)
		http.Error(w, "Erro ao propagar para a rede", http.StatusInternalServerError)
		return
	}

	log.Printf("[%s] ADMIN: Comando propagado! Nova dificuldade global: %d bits", g.id, newDifficulty)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":         "success",
		"new_difficulty": newDifficulty,
	})
}
