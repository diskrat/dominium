package mocks

import (
	"fmt"
	"sync"
)

// MockBlock representa um bloco mockado
type MockBlock struct {
	Index    int                    `json:"index"`
	Hash     string                 `json:"hash"`
	PrevHash string                 `json:"previous_hash"`
	Data     string                 `json:"data"`
	Nonce    int64                  `json:"nonce"`
	Extra    map[string]interface{} `json:"extra,omitempty"`
}

// MockTransaction representa uma transação mockada
type MockTransaction struct {
	ID        string                 `json:"id"`
	Type      string                 `json:"type"`
	Sender    string                 `json:"sender"`
	Recipient string                 `json:"recipient"`
	Data      map[string]interface{} `json:"data"`
}

// MockBlockchain simula uma blockchain para testes
type MockBlockchain struct {
	blocks       []*MockBlock
	transactions []*MockTransaction
	mempool      []*MockTransaction
	mu           sync.RWMutex
}

// NewMockBlockchain cria uma nova blockchain mockada
func NewMockBlockchain() *MockBlockchain {
	bc := &MockBlockchain{
		blocks:       make([]*MockBlock, 0),
		transactions: make([]*MockTransaction, 0),
		mempool:      make([]*MockTransaction, 0),
	}

	// Adiciona bloco genesis
	genesis := &MockBlock{
		Index:    0,
		Hash:     "genesis_hash_0000",
		PrevHash: "0",
		Data:     "Genesis Block",
		Nonce:    0,
	}
	bc.blocks = append(bc.blocks, genesis)

	return bc
}

// AddBlock adiciona um bloco à blockchain
func (bc *MockBlockchain) AddBlock(block *MockBlock) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	if len(bc.blocks) > 0 {
		lastBlock := bc.blocks[len(bc.blocks)-1]
		if block.PrevHash != lastBlock.Hash {
			return fmt.Errorf("invalid previous hash")
		}
		if block.Index != lastBlock.Index+1 {
			return fmt.Errorf("invalid block index")
		}
	}

	bc.blocks = append(bc.blocks, block)
	return nil
}

// GetBlocks retorna todos os blocos
func (bc *MockBlockchain) GetBlocks() []*MockBlock {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return append([]*MockBlock{}, bc.blocks...)
}

// GetBlock retorna um bloco específico
func (bc *MockBlockchain) GetBlock(index int) *MockBlock {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if index < 0 || index >= len(bc.blocks) {
		return nil
	}

	return bc.blocks[index]
}

// GetLastBlock retorna o último bloco
func (bc *MockBlockchain) GetLastBlock() *MockBlock {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	if len(bc.blocks) == 0 {
		return nil
	}

	return bc.blocks[len(bc.blocks)-1]
}

// AddTransaction adiciona uma transação ao mempool
func (bc *MockBlockchain) AddTransaction(tx *MockTransaction) error {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.mempool = append(bc.mempool, tx)
	return nil
}

// GetMempool retorna as transações no mempool
func (bc *MockBlockchain) GetMempool() []*MockTransaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return append([]*MockTransaction{}, bc.mempool...)
}

// ClearMempool limpa o mempool
func (bc *MockBlockchain) ClearMempool() {
	bc.mu.Lock()
	defer bc.mu.Unlock()

	bc.mempool = make([]*MockTransaction, 0)
}

// GetTransactions retorna todas as transações processadas
func (bc *MockBlockchain) GetTransactions() []*MockTransaction {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return append([]*MockTransaction{}, bc.transactions...)
}

// GetHeight retorna a altura da blockchain
func (bc *MockBlockchain) GetHeight() int {
	bc.mu.RLock()
	defer bc.mu.RUnlock()

	return len(bc.blocks)
}
