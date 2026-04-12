// package core handles the fundamental structures and rules of the distributed ledger.
// It manages block linking, state-based validation (Double Spend protection), and mining orchestration.
package core

import (
	"bytes"
	"fmt"
	"log"
	"sync"

	"github.com/diskrat/dominium/internal/transaction"
)

// Blockchain represents the immutable ledger of the network.
// It maintains the block history and a State map to track current asset ownership.
type Blockchain struct {
	mu         sync.RWMutex      // Protects the chain and state from concurrent access
	Blocks     []*Block          // The linear sequence of blocks
	Difficulty int32             // Current Proof of Work target (Nbits)
	State      map[string]string // AssetID -> OwnerID: Tracks real-time ownership of property
	NodeID     string // Unique identifier for this node (used in mining rewards and logging)
}

// NewBlockchain initializes a fresh ledger with a Genesis block and pre-registered assets.
func NewBlockchain(difficulty int32, nodeID string) *Blockchain {
	// 1. Create the Genesis block structure
	genesis := createGenesisBlock(difficulty)
	genesis.Header.MinerID = "Genesis System"

	// 2. Initialize the Blockchain object
	bc := &Blockchain{
		Blocks:     []*Block{genesis},
		Difficulty: difficulty,
		State:      make(map[string]string),
		NodeID:     nodeID,
	}

	// 3. PRE-REGISTERED ASSET (The "Copacabana Palace" Setup)
	// We assign the hotel to a master address so the Double Spend attack simulation 
	// has a valid starting point without waiting for a random registration.
	// bc.State["Copacabana-Palace"] = "Registry-Office-Natal"
	// log.Printf("[STATE] Pre-registered asset: Copacabana-Palace assigned to Registry-Office")

	return bc
}

// createGenesisBlock generates the first block (Block #0) with hardcoded values.
func createGenesisBlock(difficulty int32) *Block {
	b := &Block{
		Header: Header{
			HashOfPrevious:     make([]byte, 32), // 32 zero bytes
			Timestamp:          1711670400,
			Nbits:              difficulty,
			Nonce:              0,
			TransactionCounter: 0,
		},
		Body: Body{Transactions: []transaction.Transaction{}},
	}
	
	b.Header.MerkleRootHash = GenMerkleRoot(b.Body.Transactions)
	
	// Mine the genesis block locally to satisfy the initial difficulty target
	// Note: We use CheckDifficulty here to match our new hex-zero requirement.
	for {
		h := CalculateBlockHash(b.Header)
		if CheckDifficulty(h, difficulty) {
			b.Header.Hash = h
			break
		}
		b.Header.Nonce++
	}
	
	return b
}

// SetDifficulty updates the mining difficulty globally for this node.
func (bc *Blockchain) SetDifficulty(d int32) {
	bc.mu.Lock()
	defer bc.mu.Unlock()
	bc.Difficulty = d
	log.Printf("[BLOCKCHAIN] Difficulty adjusted to: %d", d)
}

// ValidateTransactionsAgainstState filters a list of transactions based on the confirmed ledger state.
// It ensures assets aren't registered twice and only owners can transfer their assets.
func (bc *Blockchain) ValidateTransactionsAgainstState(txs []transaction.Transaction) ([]transaction.Transaction, error) {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	var validTxs []transaction.Transaction

	for _, tx := range txs {
		currentOwner, exists := bc.State[tx.AssetID]

		if tx.Action == "register" {
			// Asset cannot already exist in the global state
			if exists {
				log.Printf("[SECURITY] Rejected Tx: Asset %s already exists.", tx.AssetID)
				continue
			}
		} else if tx.Action == "transfer" {
			// Asset must exist and the sender must be the current confirmed owner
			if !exists {
				log.Printf("[SECURITY] Rejected Tx: Asset %s not found.", tx.AssetID)
				continue 
			}
			if currentOwner != tx.SenderID {
				log.Printf("[SECURITY] Rejected Tx: Unauthorized sender for Asset %s.", tx.AssetID)
				continue 
			}
		}
		
		validTxs = append(validTxs, tx)
	}

	return validTxs, nil
}

// updateState internal helper that updates the ownership map after a block is confirmed.
func (bc *Blockchain) updateState(block *Block) {
	// This is called inside AddBlock which already holds the Lock
	for _, tx := range block.Body.Transactions {
		if tx.Action == "register" || tx.Action == "transfer" {
			bc.State[tx.AssetID] = tx.ReceiverID
			log.Printf("[STATE] Asset %s moved to owner %s", tx.AssetID, tx.ReceiverID[:8])
		}
	}
}

// AddBlock validates and appends a block, then updates the global ownership state.
func (bc *Blockchain) AddBlock(block *Block) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	// Verify links between blocks before accepting
	if !IsValidNewBlock(block, bc.Blocks[len(bc.Blocks)-1]) {
		return fmt.Errorf("[BLOCKCHAIN] Error: block validation failed")
	}

	bc.Blocks = append(bc.Blocks, block)
	bc.updateState(block)
	
	return nil
}

// MineAndAddBlock orchestrates the mining process. It filters transactions and 
// ABORTS if no valid transactions remain, preventing the creation of empty blocks.
func (bc *Blockchain) MineAndAddBlock(mempool *transaction.Mempool) (*Block, error) {
	prevHash := bc.LastBlock().Header.Hash
	
	// 1. Fetch raw transactions from mempool (Limit to 10 per block)
	rawTxs := mempool.GetTransactionsMinerMempool(10)

	// 2. Validate against Ledger State (Double Spend Defense - Layer 2)
	validTxs, _ := bc.ValidateTransactionsAgainstState(rawTxs)

	// 3. abort if no valid transactions exist.
	if len(validTxs) == 0 {
		var ids []string
		for _, tx := range rawTxs { ids = append(ids, tx.TXid) }
		mempool.RemoveTransactionsFromMempool(ids)
		return nil, fmt.Errorf("no valid transactions to process")
	}

	// 4. Assembly - Pack only the valid transactions
	candidate := AssemblerNextBlockMiner(validTxs, prevHash, bc.Difficulty)
	candidate.Header.MinerID = bc.NodeID
	// 5. Mining (Nonce hunting)
	miner := NewMiner(bc.NodeID)
	miner.BlockMiner(candidate) 

	// 6. Final confirmation and state update
	if err := bc.AddBlock(candidate); err != nil {
		return nil, err
	}

	// 7. Cleanup Mempool
	var ids []string
	for _, tx := range rawTxs { ids = append(ids, tx.TXid) }
	mempool.RemoveTransactionsFromMempool(ids)

	return candidate, nil
}

// ReplaceChain resolves forks using the "Longest Chain" rule and rebuilds the state.
func (bc *Blockchain) ReplaceChain(newBlocks []*Block) bool {
	if len(newBlocks) <= len(bc.Blocks) || !IsValidChain(newBlocks) {
		return false
	}

	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.Blocks = newBlocks
	
	// Reset and Rebuild State from the ground up based on the new chain
	bc.State = make(map[string]string)
	bc.State["Copacabana-Palace"] = "Registry-Office-Natal" // Keep pre-mined asset
	for _, b := range bc.Blocks {
		bc.updateState(b)
	}
	
	return true
}

// GetBlocks returns a read-only copy of the block slice.
func (bc *Blockchain) GetBlocks() []*Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.Blocks
}

// LastBlock returns the most recent block.
func (bc *Blockchain) LastBlock() *Block {
	bc.mu.RLock()
	defer bc.mu.RUnlock()
	return bc.Blocks[len(bc.Blocks)-1]
}

// GetState returns a thread-safe copy of the current global state map.
func (bc *Blockchain) GetState() map[string]string {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	stateCopy := make(map[string]string)
	for k, v := range bc.State {
		stateCopy[k] = v
	}
	return stateCopy
}

// ResolveTie applies a deterministic rule to resolve forks of the same height.
// It replaces the current tip if the peer's block has a "smaller" hash.
func (bc *Blockchain) ResolveTie(peerBlock *Block) bool {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bcLen := len(bc.Blocks)
	if bcLen < 2 {
		return false // Cannot resolve tie on Genesis
	}

	ourTip := bc.Blocks[bcLen-1]
	prevBlock := bc.Blocks[bcLen-2]

	// 1. Verifica se é um empate real (ambos apontam para o mesmo bloco anterior)
	if bytes.Equal(peerBlock.Header.HashOfPrevious, prevBlock.Header.Hash) {
		
		// 2. Desempate Determinístico: O Hash "menor" vence.
		if string(peerBlock.Header.Hash) < string(ourTip.Header.Hash) {
			
			// Nós perdemos. Substitui nosso bloco pelo do vencedor (peer).
			bc.Blocks[bcLen-1] = peerBlock
			
			// Reconstrói o State Global do zero para garantir 100% de precisão
			bc.State = make(map[string]string)
			for _, b := range bc.Blocks {
				bc.updateState(b)
			}
			return true
		}
	}
	return false
}