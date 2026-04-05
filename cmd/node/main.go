package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"math"
	"net/http"
	"os"
	"os/signal"
	"sort"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"dominium/internal/transaction"
	"dominium/pkg/crypto"
	amqp "github.com/rabbitmq/amqp091-go"
)

type BlockView struct {
	Index          int                      `json:"index"`
	Hash           string                   `json:"hash"`
	PreviousHash   string                   `json:"previous_hash"`
	Timestamp      int64                    `json:"timestamp"`
	Nonce          int64                    `json:"nonce"`
	Difficulty     int                      `json:"difficulty"`
	MinerID        string                   `json:"miner_id"`
	Transaction    *transaction.Transaction `json:"transaction,omitempty"`
	CumulativeWork int64                    `json:"cumulative_work"`
	Source         string                   `json:"source"`
}

type p2pEnvelope struct {
	Type    string          `json:"type"`
	Payload json.RawMessage `json:"payload"`
	Sender  string          `json:"sender"`
}

type txPayload struct {
	TxJSON string `json:"tx_json"`
}

type blockPayload struct {
	BlockJSON string `json:"block_json"`
}

type Node struct {
	nodeID     string
	apiAddr    string
	amqpURL    string
	difficulty int

	state   *transaction.AccountState
	mempool *transaction.Mempool
	gen     *transaction.Generator
	admin   *transaction.WalletIdentity
	alice   *transaction.WalletIdentity
	bob     *transaction.WalletIdentity

	mu           sync.RWMutex
	mineMu       sync.Mutex
	blocksByHash map[string]*BlockView
	tips         map[string]struct{}
	headHash     string
	exchange     string

	httpServer *http.Server
	amqpConn   *amqp.Connection
	amqpCh     *amqp.Channel
	queue      string

	ctx    context.Context
	cancel context.CancelFunc
}

func NewNode(nodeID, apiAddr, amqpURL string, difficulty int) *Node {
	if difficulty < 0 {
		difficulty = 0
	}
	if difficulty > 8 {
		difficulty = 8
	}

	state := transaction.NewAccountState()
	ctx, cancel := context.WithCancel(context.Background())
	return &Node{
		nodeID:       nodeID,
		apiAddr:      apiAddr,
		amqpURL:      amqpURL,
		difficulty:   difficulty,
		state:        state,
		mempool:      transaction.NewMempool(state),
		gen:          transaction.NewGenerator(42),
		blocksByHash: make(map[string]*BlockView),
		tips:         make(map[string]struct{}),
		exchange:     "blockchain",
		ctx:          ctx,
		cancel:       cancel,
	}
}

func (n *Node) Start() error {
	if err := n.bootstrapIdentities(); err != nil {
		return err
	}

	genesis := n.mineBlock("", 0, nil)
	if !n.addBlock(genesis, "genesis") {
		return errors.New("falha ao criar genesis block")
	}

	if n.amqpURL != "" {
		if err := n.startRabbit(); err != nil {
			return err
		}
	}

	n.startDemoTraffic()

	mux := http.NewServeMux()
	mux.HandleFunc("/health", n.handleHealth)
	mux.HandleFunc("/blocks", n.handleBlocks)
	mux.HandleFunc("/forks", n.handleForks)
	mux.HandleFunc("/mempool", n.handleMempool)
	mux.HandleFunc("/stats", n.handleStats)
	mux.HandleFunc("/difficulty", n.handleDifficulty)
	mux.HandleFunc("/api/difficulty", n.handleDifficulty)
	mux.HandleFunc("/api/transactions", n.handleAddTransaction)

	n.httpServer = &http.Server{Addr: n.apiAddr, Handler: withCORS(mux)}

	go func() {
		if err := n.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Printf("http server error: %v", err)
		}
	}()

	log.Printf("node %s iniciado em %s (difficulty=%d)", n.nodeID, n.apiAddr, n.getDifficulty())
	return nil
}

func (n *Node) Stop(ctx context.Context) error {
	n.cancel()
	var stopErr error
	if n.httpServer != nil {
		if err := n.httpServer.Shutdown(ctx); err != nil {
			stopErr = err
		}
	}
	if n.amqpCh != nil {
		_ = n.amqpCh.Close()
	}
	if n.amqpConn != nil {
		_ = n.amqpConn.Close()
	}
	return stopErr
}

func (n *Node) bootstrapIdentities() error {
	admin, err := n.gen.CreateIdentity("admin")
	if err != nil {
		return err
	}
	alice, err := n.gen.CreateIdentity("alice")
	if err != nil {
		return err
	}
	bob, err := n.gen.CreateIdentity("bob")
	if err != nil {
		return err
	}

	for _, identity := range []*transaction.WalletIdentity{admin, alice, bob} {
		n.ensureAccount(identity.PublicKey)
	}
	if err := n.gen.SetAdmin(admin); err != nil {
		return err
	}

	n.admin = admin
	n.alice = alice
	n.bob = bob

	mintTx, err := n.gen.MintTx(alice, "nft-demo-001")
	if err != nil {
		return err
	}
	if err := n.acceptTransaction(mintTx, false); err != nil {
		return err
	}
	return nil
}

func (n *Node) ensureAccount(pubKey string) {
	if pubKey == "" {
		return
	}
	if _, err := n.state.CreateAccount(pubKey); err != nil {
		if !strings.Contains(err.Error(), "conta ja existe") {
			log.Printf("erro ao criar conta %s: %v", pubKey, err)
		}
	}
}

func (n *Node) startDemoTraffic() {
	go func() {
		ticker := time.NewTicker(12 * time.Second)
		defer ticker.Stop()
		toggle := false
		for {
			select {
			case <-n.ctx.Done():
				return
			case <-ticker.C:
				sender := n.alice
				receiver := n.bob
				if toggle {
					sender, receiver = n.bob, n.alice
				}
				toggle = !toggle
				tx, err := n.gen.TransferTx(sender, receiver, "nft-demo-001")
				if err != nil {
					log.Printf("erro criando tx demo: %v", err)
					continue
				}
				if err := n.acceptTransaction(tx, false); err != nil {
					log.Printf("erro aplicando tx demo: %v", err)
				}
			}
		}
	}()
}

func (n *Node) acceptTransaction(tx *transaction.Transaction, publishTx bool) error {
	if tx == nil {
		return errors.New("transacao nula")
	}
	if tx.ID == "" {
		if err := tx.SetID(); err != nil {
			return err
		}
	}

	n.ensureAccount(tx.PublKey)
	n.ensureAccount(tx.Recipient)

	if err := n.mempool.Add(tx); err != nil {
		return err
	}
	if publishTx {
		if err := n.publishTx(tx); err != nil {
			return err
		}
	}

	if err := n.mineFromMempoolAndBroadcast(); err != nil {
		return err
	}
	return nil
}

func (n *Node) mineFromMempoolAndBroadcast() error {
	n.mineMu.Lock()
	defer n.mineMu.Unlock()

	pending := n.mempool.GetPending(1)
	if len(pending) == 0 {
		return nil
	}

	tx, ok := pending[0].(*transaction.Transaction)
	if !ok {
		return errors.New("tipo de transacao invalido na mempool")
	}

	n.mempool.Remove([]string{tx.ID})
	if err := tx.Execute(n.state); err != nil {
		return err
	}

	head, headHash := n.getHead()
	index := 0
	prevHash := ""
	if head != nil {
		index = head.Index + 1
		prevHash = headHash
	}

	block := n.mineBlock(prevHash, index, tx)
	if !n.addBlock(block, "local") {
		return errors.New("bloco local rejeitado")
	}

	if err := n.publishBlock(block); err != nil {
		return err
	}
	return nil
}

func (n *Node) mineBlock(prevHash string, index int, tx *transaction.Transaction) *BlockView {
	difficulty := n.getDifficulty()
	timestamp := time.Now().Unix()
	nonce := int64(0)
	for {
		hash := calculateBlockHash(prevHash, index, timestamp, nonce, difficulty, n.nodeID, tx)
		if checkDifficulty(hash, difficulty) {
			return &BlockView{
				Index:        index,
				Hash:         hash,
				PreviousHash: prevHash,
				Timestamp:    timestamp,
				Nonce:        nonce,
				Difficulty:   difficulty,
				MinerID:      n.nodeID,
				Transaction:  tx,
			}
		}
		nonce++
	}
}

func calculateBlockHash(prevHash string, index int, timestamp int64, nonce int64, difficulty int, minerID string, tx *transaction.Transaction) string {
	txID := ""
	if tx != nil {
		txID = tx.ID
	}
	payload := fmt.Sprintf("%s|%d|%d|%d|%d|%s|%s", prevHash, index, timestamp, nonce, difficulty, minerID, txID)
	return crypto.Hash([]byte(payload))
}

func checkDifficulty(hash string, difficulty int) bool {
	if difficulty <= 0 {
		return true
	}
	prefix := strings.Repeat("0", difficulty)
	return strings.HasPrefix(hash, prefix)
}

func blockWork(difficulty int) int64 {
	if difficulty <= 0 {
		return 1
	}
	if difficulty > 60 {
		difficulty = 60
	}
	return int64(math.Pow(2, float64(difficulty)))
}

func (n *Node) validateBlock(block *BlockView) error {
	if block == nil {
		return errors.New("bloco nulo")
	}
	if block.Index < 0 {
		return errors.New("indice invalido")
	}
	if !checkDifficulty(block.Hash, block.Difficulty) {
		return errors.New("dificuldade invalida")
	}
	expected := calculateBlockHash(block.PreviousHash, block.Index, block.Timestamp, block.Nonce, block.Difficulty, block.MinerID, block.Transaction)
	if expected != block.Hash {
		return errors.New("hash do bloco invalido")
	}
	if block.Index > 0 {
		prev, ok := n.getBlock(block.PreviousHash)
		if !ok {
			return errors.New("bloco anterior ausente")
		}
		if block.Index != prev.Index+1 {
			return errors.New("indice nao sequencial")
		}
	}
	return nil
}

func (n *Node) addBlock(block *BlockView, source string) bool {
	if err := n.validateBlock(block); err != nil {
		log.Printf("bloco rejeitado: %v", err)
		return false
	}

	n.mu.Lock()
	defer n.mu.Unlock()

	if _, exists := n.blocksByHash[block.Hash]; exists {
		return false
	}

	if block.Index == 0 {
		block.CumulativeWork = blockWork(block.Difficulty)
	} else {
		prev := n.blocksByHash[block.PreviousHash]
		block.CumulativeWork = prev.CumulativeWork + blockWork(block.Difficulty)
	}
	block.Source = source

	n.blocksByHash[block.Hash] = block
	n.tips[block.Hash] = struct{}{}
	if block.PreviousHash != "" {
		delete(n.tips, block.PreviousHash)
	}

	if n.headHash == "" || block.CumulativeWork > n.blocksByHash[n.headHash].CumulativeWork {
		n.headHash = block.Hash
	}

	return true
}

func (n *Node) getBlock(hash string) (*BlockView, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	b, ok := n.blocksByHash[hash]
	return b, ok
}

func (n *Node) getHead() (*BlockView, string) {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.headHash == "" {
		return nil, ""
	}
	return n.blocksByHash[n.headHash], n.headHash
}

func (n *Node) getDifficulty() int {
	n.mu.RLock()
	defer n.mu.RUnlock()
	return n.difficulty
}

func (n *Node) setDifficulty(v int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.difficulty = v
}

func (n *Node) snapshotMainChain() []BlockView {
	n.mu.RLock()
	defer n.mu.RUnlock()
	if n.headHash == "" {
		return []BlockView{}
	}

	chain := make([]BlockView, 0)
	cur := n.blocksByHash[n.headHash]
	for cur != nil {
		chain = append(chain, *cur)
		if cur.PreviousHash == "" {
			break
		}
		cur = n.blocksByHash[cur.PreviousHash]
	}

	for i, j := 0, len(chain)-1; i < j; i, j = i+1, j-1 {
		chain[i], chain[j] = chain[j], chain[i]
	}
	return chain
}

func (n *Node) snapshotForks() [][]BlockView {
	n.mu.RLock()
	defer n.mu.RUnlock()

	if n.headHash == "" {
		return [][]BlockView{}
	}

	mainSet := map[string]struct{}{}
	cur := n.blocksByHash[n.headHash]
	for cur != nil {
		mainSet[cur.Hash] = struct{}{}
		if cur.PreviousHash == "" {
			break
		}
		cur = n.blocksByHash[cur.PreviousHash]
	}

	forks := make([][]BlockView, 0)
	for tipHash := range n.tips {
		if tipHash == n.headHash {
			continue
		}
		branch := make([]BlockView, 0)
		cursor := n.blocksByHash[tipHash]
		for cursor != nil {
			if _, inMain := mainSet[cursor.Hash]; inMain {
				break
			}
			branch = append(branch, *cursor)
			if cursor.PreviousHash == "" {
				break
			}
			cursor = n.blocksByHash[cursor.PreviousHash]
		}
		if len(branch) > 0 {
			for i, j := 0, len(branch)-1; i < j; i, j = i+1, j-1 {
				branch[i], branch[j] = branch[j], branch[i]
			}
			forks = append(forks, branch)
		}
	}

	sort.Slice(forks, func(i, j int) bool {
		return len(forks[i]) > len(forks[j])
	})
	return forks
}

func (n *Node) startRabbit() error {
	conn, err := amqp.Dial(n.amqpURL)
	if err != nil {
		return fmt.Errorf("falha ao conectar RabbitMQ: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("falha ao abrir canal RabbitMQ: %w", err)
	}

	if err := ch.ExchangeDeclare(n.exchange, "topic", true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("falha ao declarar exchange: %w", err)
	}

	n.queue = fmt.Sprintf("node_%s", n.nodeID)
	if _, err := ch.QueueDeclare(n.queue, true, false, false, false, nil); err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("falha ao declarar fila: %w", err)
	}

	for _, key := range []string{"tx.*", "block.*"} {
		if err := ch.QueueBind(n.queue, key, n.exchange, false, nil); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return fmt.Errorf("falha ao bind da fila: %w", err)
		}
	}

	deliveries, err := ch.Consume(n.queue, "", false, false, false, false, nil)
	if err != nil {
		_ = ch.Close()
		_ = conn.Close()
		return fmt.Errorf("falha ao consumir fila: %w", err)
	}

	n.amqpConn = conn
	n.amqpCh = ch
	go n.consumeRabbit(deliveries)
	return nil
}

func (n *Node) consumeRabbit(deliveries <-chan amqp.Delivery) {
	for {
		select {
		case <-n.ctx.Done():
			return
		case d, ok := <-deliveries:
			if !ok {
				return
			}

			var env p2pEnvelope
			if err := json.Unmarshal(d.Body, &env); err != nil {
				log.Printf("mensagem invalida: %v", err)
				_ = d.Ack(false)
				continue
			}
			if env.Sender == n.nodeID {
				_ = d.Ack(false)
				continue
			}

			switch env.Type {
			case "TX_NEW":
				var p txPayload
				if err := json.Unmarshal(env.Payload, &p); err != nil {
					log.Printf("payload tx invalido: %v", err)
					_ = d.Ack(false)
					continue
				}
				var tx transaction.Transaction
				if err := json.Unmarshal([]byte(p.TxJSON), &tx); err != nil {
					log.Printf("tx json invalido: %v", err)
					_ = d.Ack(false)
					continue
				}
				if err := n.acceptTransaction(&tx, false); err != nil {
					log.Printf("tx recebida rejeitada: %v", err)
				}
			case "BLOCK_NEW":
				var p blockPayload
				if err := json.Unmarshal(env.Payload, &p); err != nil {
					log.Printf("payload block invalido: %v", err)
					_ = d.Ack(false)
					continue
				}
				var b BlockView
				if err := json.Unmarshal([]byte(p.BlockJSON), &b); err != nil {
					log.Printf("block json invalido: %v", err)
					_ = d.Ack(false)
					continue
				}
				if ok := n.addBlock(&b, "network"); !ok {
					log.Printf("bloco de rede rejeitado")
				}
			}

			_ = d.Ack(false)
		}
	}
}

func (n *Node) publishTx(tx *transaction.Transaction) error {
	if n.amqpCh == nil {
		return nil
	}
	txBytes, err := json.Marshal(tx)
	if err != nil {
		return err
	}
	payloadBytes, err := json.Marshal(txPayload{TxJSON: string(txBytes)})
	if err != nil {
		return err
	}
	envBytes, err := json.Marshal(p2pEnvelope{Type: "TX_NEW", Payload: payloadBytes, Sender: n.nodeID})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return n.amqpCh.PublishWithContext(ctx, n.exchange, "tx.new", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         envBytes,
	})
}

func (n *Node) publishBlock(block *BlockView) error {
	if n.amqpCh == nil {
		return nil
	}
	blockBytes, err := json.Marshal(block)
	if err != nil {
		return err
	}
	payloadBytes, err := json.Marshal(blockPayload{BlockJSON: string(blockBytes)})
	if err != nil {
		return err
	}
	envBytes, err := json.Marshal(p2pEnvelope{Type: "BLOCK_NEW", Payload: payloadBytes, Sender: n.nodeID})
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return n.amqpCh.PublishWithContext(ctx, n.exchange, "block.new", false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         envBytes,
	})
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func (n *Node) handleHealth(w http.ResponseWriter, _ *http.Request) {
	chain := n.snapshotMainChain()
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"status":     "healthy",
		"blocks":     len(chain),
		"mempool":    n.mempool.Count(),
		"connected":  n.amqpConn != nil && !n.amqpConn.IsClosed(),
		"node_id":    n.nodeID,
		"difficulty": n.getDifficulty(),
		"timestamp":  time.Now().Unix(),
	})
}

func (n *Node) handleBlocks(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n.snapshotMainChain())
}

func (n *Node) handleForks(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(n.snapshotForks())
}

func (n *Node) handleMempool(w http.ResponseWriter, _ *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	pending := n.mempool.GetPending(1000)
	if pending == nil {
		pending = make([]transaction.Tx, 0)
	}
	_ = json.NewEncoder(w).Encode(pending)
}

func (n *Node) handleStats(w http.ResponseWriter, _ *http.Request) {
	accounts, nftCount := n.state.Stats()
	chain := n.snapshotMainChain()
	forks := n.snapshotForks()
	totalWork := int64(0)
	if len(chain) > 0 {
		totalWork = chain[len(chain)-1].CumulativeWork
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{
		"node_id":       n.nodeID,
		"total_blocks":  len(chain),
		"forks":         len(forks),
		"total_work":    totalWork,
		"difficulty":    n.getDifficulty(),
		"mempool_size":  n.mempool.Count(),
		"accounts":      accounts,
		"tracked_nfts":  nftCount,
		"p2p_connected": n.amqpConn != nil && !n.amqpConn.IsClosed(),
	})
}

func (n *Node) handleDifficulty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	valueStr := r.URL.Query().Get("value")
	if valueStr == "" {
		http.Error(w, "value ausente", http.StatusBadRequest)
		return
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil || value < 0 || value > 8 {
		http.Error(w, "value invalido (0..8)", http.StatusBadRequest)
		return
	}
	n.setDifficulty(value)
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "difficulty": value})
}

func (n *Node) handleAddTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	var tx transaction.Transaction
	if err := json.NewDecoder(r.Body).Decode(&tx); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err := n.acceptTransaction(&tx, true); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]interface{}{"success": true, "tx_id": tx.ID})
}

func main() {
	var (
		apiAddr    = flag.String("api", ":8080", "endereco da API")
		nodeID     = flag.String("id", "node-1", "identificador do no")
		amqpURL    = flag.String("amqp", "", "URL do RabbitMQ (vazio desabilita)")
		difficulty = flag.Int("difficulty", 2, "dificuldade PoW (0..8)")
	)
	flag.Parse()

	n := NewNode(*nodeID, *apiAddr, *amqpURL, *difficulty)
	if err := n.Start(); err != nil {
		log.Fatalf("start: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := n.Stop(shutdownCtx); err != nil {
		log.Printf("stop: %v", err)
	}
}
