package transaction

import (
	"sync"
)

type Mempool struct {
	mu           sync.RWMutex
	transactions map[string]*Transaction
}

func NewMempool() *Mempool {
	return &Mempool{
		transactions: make(map[string]*Transaction),
	}
}

func (m *Mempool) Add(tx *Transaction) error {
	return nil
}

func (m *Mempool) GetPending(limit int) []*Transaction {

	m.mu.RLock()
	defer m.mu.RUnlock()

	var pending []*Transaction
	count := 0

	for _, tx := range m.transactions {
		if count >= limit {
			break
		}
		pending = append(pending, tx)
		count++
	}

	return pending
}

func (m *Mempool) Remove(txIDs []string) {

	m.mu.Lock()
	defer m.mu.Unlock()

	for _, id := range txIDs {
		delete(m.transactions, id)
	}
}
