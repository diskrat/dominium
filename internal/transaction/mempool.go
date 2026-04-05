package transaction

import (
	"errors"
	"sync"
)

type Tx interface {
	GetID() string
	Validate(state *AccountState) error
	Execute(state *AccountState) error
}

// Mempool armazena transacoes pendentes.
type Mempool struct {
	mu           sync.RWMutex
	transactions map[string]Tx
	state        *AccountState
}

func NewMempool(state *AccountState) *Mempool {
	return &Mempool{
		transactions: make(map[string]Tx),
		state:        state,
	}
}

func (m *Mempool) Add(tx Tx) error {
	if tx == nil {
		return errors.New("transacao nula")
	}
	if tx.GetID() == "" {
		return errors.New("ID da transacao ausente")
	}

	m.mu.RLock()
	state := m.state
	m.mu.RUnlock()
	if state == nil {
		return errors.New("estado da mempool nao configurado")
	}

	if err := tx.Validate(state); err != nil {
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.transactions[tx.GetID()]; exists {
		return errors.New("transacao ja existe na mempool")
	}

	m.transactions[tx.GetID()] = tx
	return nil
}

func (m *Mempool) GetPending(limit int) []Tx {

	m.mu.RLock()
	defer m.mu.RUnlock()

	var pending []Tx
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

func (m *Mempool) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return len(m.transactions)
}
