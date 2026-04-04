package node

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"dominium/internal/miner"
	"dominium/internal/network"
	"dominium/internal/transaction"

	"github.com/segmentio/kafka-go"
)

const (
	defaultBroker                 = "localhost:9092"
	networkSilenceTimeout         = 10 * time.Second
	bootstrapCommitInterval       = 1 * time.Second
	kafkaReconnectDelay           = 2 * time.Second
	genesisTimestamp        int64 = 1672531200000000000 // 2023-01-01T00:00:00Z
)

type NodeServer struct {
	ctx        context.Context
	cancel     context.CancelFunc
	id         string
	brokers    []string
	mine       bool
	difficulty int32
	dataDir    string

	state          *transaction.AccountState
	mempool        *transaction.Mempool
	blockchain     *miner.Blockchain
	txTransport    *network.KafkaTransactionTransport
	blockTransport *network.KafkaBlockTransport

	synced        atomic.Bool
	healthyKafka  atomic.Bool
	bootstrapOnce sync.Once
}

func NewNodeServer(parent context.Context, id string, brokers []string, mine bool, difficulty int32, dataDir string) *NodeServer {
	ctx, cancel := context.WithCancel(parent)
	if len(brokers) == 0 {
		brokers = []string{defaultBroker}
	}
	state := transaction.NewAccountState()
	return &NodeServer{
		ctx:        ctx,
		cancel:     cancel,
		id:         strings.TrimSpace(id),
		brokers:    brokers,
		mine:       mine,
		difficulty: difficulty,
		dataDir:    dataDir,
		state:      state,
		mempool:    transaction.NewMempool(state),
		blockchain: miner.NewBlockchain(),
	}
}

func (n *NodeServer) Run() error {
	if n.id == "" {
		return errors.New("-id é obrigatório")
	}
	if n.difficulty <= 0 {
		n.difficulty = 4
	}

	n.txTransport = network.NewKafkaTransactionTransport(n.ctx, n.brokers, n.id)
	n.blockTransport = network.NewKafkaBlockTransport(n.ctx, n.brokers, n.id)
	n.healthyKafka.Store(true)

	if err := n.subscribeTransactions(); err != nil {
		return err
	}

	if err := n.bootstrapBlocks(); err != nil {
		return err
	}

	if n.mine {
		go n.miningLoop()
	}

	log.Printf("[%s] node iniciado (mine=%v difficulty=%d brokers=%v)", n.id, n.mine, n.difficulty, n.brokers)
	<-n.ctx.Done()
	return n.shutdown()
}

func (n *NodeServer) subscribeTransactions() error {
	return n.txTransport.Subscribe(n.handleIncomingTransaction)
}

func (n *NodeServer) handleIncomingTransaction(tx *transaction.Transaction) error {
	if tx == nil {
		return errors.New("transacao nula")
	}

	if err := n.ensureAccount(tx.PublKey); err != nil {
		return err
	}
	if err := n.ensureAccount(tx.Recipient); err != nil {
		return err
	}

	if err := n.mempool.Add(tx); err != nil {
		log.Printf("[%s] transacao descartada na mempool: %v", n.id, err)
		return err
	}
	log.Printf("[%s] transacao recebida: %s", n.id, tx.GetID())
	return nil
}

func (n *NodeServer) bootstrapBlocks() error {
	blockArrived := make(chan struct{}, 1)

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        n.brokers,
		Topic:          network.TopicBlocks,
		GroupID:        n.id + "-block-sync",
		StartOffset:    kafka.FirstOffset,
		MinBytes:       1,
		MaxBytes:       10e6,
		CommitInterval: bootstrapCommitInterval,
	})

	go func() {
		defer func() {
			_ = reader.Close()
		}()

		for {
			m, err := reader.ReadMessage(n.ctx)
			if err != nil {
				if n.ctx.Err() != nil {
					return
				}
				n.pauseMining()
				log.Printf("[%s] erro de leitura do kafka de blocos: %v", n.id, err)
				time.Sleep(kafkaReconnectDelay)
				continue
			}

			n.resumeMining()
			select {
			case blockArrived <- struct{}{}:
			default:
			}

			var block miner.Block
			if err := json.Unmarshal(m.Value, &block); err != nil {
				log.Printf("[%s] bloco kafka invalido: %v", n.id, err)
				continue
			}

			if err := n.processBlock(&block); err != nil {
				log.Printf("[%s] falha ao processar bloco recebido: %v", n.id, err)
			}

			if err := reader.CommitMessages(n.ctx, m); err != nil {
				log.Printf("[%s] falha ao commitar offset do bloco: %v", n.id, err)
			}
		}
	}()

	select {
	case <-blockArrived:
		n.synced.Store(true)
		log.Printf("[%s] sincronizado com blockchain existente", n.id)
	case <-time.After(networkSilenceTimeout):
		if n.blockchain.IsEmpty() && n.mine {
			log.Printf("[%s] nenhuma blockchain encontrada, gerando bloco genesis", n.id)
			if err := n.createGenesisBlock(); err != nil {
				return fmt.Errorf("erro ao criar genesis: %w", err)
			}
			n.synced.Store(true)
		}
	}

	return nil
}

func (n *NodeServer) processBlock(block *miner.Block) error {
	if block == nil {
		return errors.New("bloco nulo")
	}

	if len(block.Hash) == 0 {
		return errors.New("bloco sem hash")
	}

	if !n.validateBlockTransactions(block) {
		return errors.New("transacoes do bloco invalidas")
	}

	// Verifica se este bloco causará uma reorganização
	oldBestHash := n.blockchain.GetLatestHash()
	oldChainLength := n.blockchain.GetChainLength()

	if err := n.blockchain.AddBlock(*block); err != nil {
		if strings.Contains(err.Error(), "ja existe") {
			return nil
		}
		return err
	}

	newChainLength := n.blockchain.GetChainLength()

	// Se houve reorganização (cadeia ficou maior), processa reorg
	if newChainLength > oldChainLength {
		log.Printf("[Consenso] Reorganização detectada - Altura Local: %d, Altura Recebida: %d", oldChainLength, newChainLength)
		log.Printf("[Consenso] Reorganização detectada - Altura Local: %d, Altura Recebida: %d", oldChainLength, newChainLength)
		if err := n.handleReorganization(oldBestHash, block.Hash); err != nil {
			log.Printf("[%s] erro na reorganizacao: %v", n.id, err)
			// Em caso de erro, tenta reverter para o estado anterior
			return err
		}
	}

	if err := n.rebuildStateFromCanonicalChain(); err != nil {
		return err
	}

	n.removeBlockTransactionsFromMempool(block)
	return nil
}

func (n *NodeServer) handleReorganization(oldTipHash, newTipHash []byte) error {
	// A reorganização já foi feita pelo AddBlock, agora precisamos
	// devolver as transações dos blocos desconectados para a mempool

	// Para simplificar, vamos reconstruir a mempool baseada na corrente atual
	// removendo transações que já estão na blockchain
	currentChain := n.blockchain.GetCanonicalChain()

	// Coleta todas as transações que estão na corrente atual
	blockTxs := make(map[string]bool)
	for _, block := range currentChain {
		for _, tx := range block.Transactions {
			blockTxs[tx.GetID()] = true
		}
	}

	// Recria a mempool apenas com transações pendentes
	pendingTxs := n.mempool.GetPending(10000)
	n.mempool = transaction.NewMempool(n.state)

	for _, tx := range pendingTxs {
		if !blockTxs[tx.GetID()] {
			if err := n.mempool.Add(tx); err != nil {
				log.Printf("[%s] falha ao devolver transacao para mempool apos reorg: %v", n.id, err)
			}
		}
	}

	log.Printf("[%s] reorganizacao concluida: tip mudou de %x para %x", n.id, oldTipHash, newTipHash)
	return nil
}

func (n *NodeServer) validateBlockTransactions(block *miner.Block) bool {
	// Para validação de consenso, precisamos de uma chave admin conhecida
	// Por enquanto, assumimos que qualquer mint é válido (isso pode ser configurado)
	adminPubKey := "" // TODO: configurar chave admin conhecida

	stateCopy := n.state.Clone()
	for _, tx := range block.Transactions {
		if err := stateCopy.EnsureAccount(tx.PublKey); err != nil {
			log.Printf("[%s] falha ao garantir conta de publicador: %v", n.id, err)
			return false
		}
		if err := stateCopy.EnsureAccount(tx.Recipient); err != nil {
			log.Printf("[%s] falha ao garantir conta de destinatario: %v", n.id, err)
			return false
		}

		// Validação básica
		if err := tx.Validate(stateCopy); err != nil {
			log.Printf("[%s] bloco contem transacao invalida %s: %v", n.id, tx.ID, err)
			return false
		}

		// Validação de regras de consenso específicas
		if err := tx.ValidateConsensusRules(stateCopy, adminPubKey); err != nil {
			log.Printf("[%s] bloco contem transacao que viola consenso %s: %v", n.id, tx.ID, err)
			return false
		}

		if err := tx.Execute(stateCopy); err != nil {
			log.Printf("[%s] bloco contem transacao invalida %s: %v", n.id, tx.ID, err)
			return false
		}
	}
	return true
}

func (n *NodeServer) rebuildStateFromCanonicalChain() error {
	chain := n.blockchain.GetCanonicalChain()
	state := transaction.NewAccountState()

	for _, block := range chain {
		for _, tx := range block.Transactions {
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

	pending := n.mempool.GetPending(10000)
	n.state = state
	n.mempool = transaction.NewMempool(state)
	for _, tx := range pending {
		if err := n.mempool.Add(tx); err != nil {
			log.Printf("[%s] descarte transacao pendente apos rebuild: %v", n.id, err)
		}
	}
	return nil
}

func (n *NodeServer) removeBlockTransactionsFromMempool(block *miner.Block) {
	var ids []string
	for _, tx := range block.Transactions {
		if tx.ID != "" {
			ids = append(ids, tx.ID)
		}
	}
	if len(ids) > 0 {
		n.mempool.Remove(ids)
	}
}

func (n *NodeServer) createGenesisBlock() error {
	genesis := miner.NewBlock([]byte{}, []transaction.Transaction{}, n.difficulty)
	genesis.Timestamp = genesisTimestamp
	genesis.MerkleRootHash = miner.CalculateMerkleRoot(genesis.Transactions)
	miner.Mine(genesis)

	if err := n.processBlock(genesis); err != nil {
		return err
	}
	return n.blockTransport.Publish(genesis)
}

func (n *NodeServer) miningLoop() {
	log.Printf("[%s] motor de mineracao ativado", n.id)
	for {
		if n.ctx.Err() != nil {
			return
		}
		if !n.synced.Load() {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		if n.blockchain.IsEmpty() {
			time.Sleep(250 * time.Millisecond)
			continue
		}
		if !n.healthyKafka.Load() {
			log.Printf("[%s] kafka desconectado, mineracao em espera", n.id)
			time.Sleep(kafkaReconnectDelay)
			continue
		}

		pending := n.mempool.GetPending(5)
		if len(pending) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		var txs []transaction.Transaction
		var txIDs []string
		for _, item := range pending {
			if tx, ok := item.(*transaction.Transaction); ok {
				txs = append(txs, *tx)
				txIDs = append(txIDs, tx.GetID())
			}
		}

		if len(txs) == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		block := miner.NewBlock(n.blockchain.GetLatestHash(), txs, n.difficulty)
		miner.Mine(block)
		log.Printf("[%s] bloco minerado com hash %x", n.id, block.Hash)

		if err := n.processBlock(block); err != nil {
			log.Printf("[%s] falha ao processar bloco minerado: %v", n.id, err)
			continue
		}
		if err := n.blockTransport.Publish(block); err != nil {
			log.Printf("[%s] falha ao publicar bloco: %v", n.id, err)
		}
		n.mempool.Remove(txIDs)
	}
}

func (n *NodeServer) ensureAccount(pubKey string) error {
	if pubKey == "" {
		return errors.New("chave publica vazia")
	}
	return n.state.EnsureAccount(pubKey)
}

func (n *NodeServer) pauseMining() {
	n.healthyKafka.Store(false)
}

func (n *NodeServer) resumeMining() {
	n.healthyKafka.Store(true)
}

func (n *NodeServer) shutdown() error {
	log.Printf("[%s] encerrando node...", n.id)
	if n.txTransport != nil {
		n.txTransport.Stop()
	}
	if n.blockTransport != nil {
		n.blockTransport.Stop()
	}
	n.cancel()
	return nil
}
