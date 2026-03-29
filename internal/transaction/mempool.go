// package transaction handles the staging area for unconfirmed transactions.
package transaction

import (
	"fmt"
	"log"
	"sort"
	"sync"
)

// Mempool represents a thread-safe in-memory pool of transactions
// waiting to be included in a block by a miner.
type Mempool struct {
	mu              sync.RWMutex           // ensures safe concurrent access.
	Transactions    map[string]Transaction // unique key-value store for pending txs.
	MaxTransactions int                    // pool capacity limit.
}

// NewMempool initializes and returns a new Mempool instance.
func NewMempool(MaxTransactions int) *Mempool {
	return &Mempool{
		Transactions:    make(map[string]Transaction),
		MaxTransactions: MaxTransactions,
	}
}

// AddTransactionToMempool performs safety checks and adds a valid transaction to the pool.
func (m *Mempool) AddTransactionToMempool(tx Transaction) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// check if the pool has reached its defined capacity.
	if len(m.Transactions) >= m.MaxTransactions {
		err := fmt.Errorf("mempool is full")
		log.Printf("[MEMPOOL] Error: %v", err)
		return err
	}

	// verify if the transaction ID already exists to prevent duplicates.
	_, ok := m.Transactions[tx.TXid]
	if ok {
		err := fmt.Errorf("transaction %s already in queue", tx.TXid)
		log.Printf("[MEMPOOL] Error: %v", err)
		return err
	}

	// ==========================================
	// DOUBLE SPEND DEFENSE (Requirement 6)
	// ==========================================
	// Scan the mempool to ensure no one is trying to transfer or register 
	// the same asset ID in another pending transaction.
	for _, pendingTx := range m.Transactions {
		if pendingTx.AssetID == tx.AssetID {
			err := fmt.Errorf("DOUBLE SPEND ATTEMPT DETECTED: Asset %s is already pending", tx.AssetID)
			log.Printf("[SECURITY] Alert: %v", err)
			return err
		}
	}

	// store the transaction.
	m.Transactions[tx.TXid] = tx

	// log successful acceptance.
	log.Printf("[MEMPOOL] Accept: %s | Action: %s | Asset: %s | Queue: %d/%d",
		tx.TXid, tx.Action, tx.AssetID, len(m.Transactions), m.MaxTransactions)

	return nil
}

// RemoveTransactionsFromMempool removes mined transactions from the pool.
func (m *Mempool) RemoveTransactionsFromMempool(txids []string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range txids {
		delete(m.Transactions, id)
	}
	log.Printf("[MEMPOOL] Cleared: %d transactions removed after mining", len(txids))
}

// GetTransactionsMinerMempool selects the top transactions by fee to be included in a block.
func (m *Mempool) GetTransactionsMinerMempool(limit int) []Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var selected []Transaction
	for _, tx := range m.Transactions {
		selected = append(selected, tx)
	}

	// sort by fee: highest first
	sort.Slice(selected, func(i, j int) bool {
		return selected[i].Fee > selected[j].Fee
	})

	if len(selected) > limit {
		selected = selected[:limit]
	}

	return selected
}

// GetPendingTransactions returns a thread-safe copy of all transactions currently in the mempool.
// Used by the API to serve the frontend Dashboard.
func (m *Mempool) GetPendingTransactions() []Transaction {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var pending []Transaction
	for _, tx := range m.Transactions {
		pending = append(pending, tx)
	}

	return pending
}