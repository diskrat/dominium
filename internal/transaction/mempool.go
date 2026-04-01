package transaction

import (
	"errors"
	"sort"
	"sync"
)

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
	if tx.IDValue() == "" {
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

	if _, exists := m.transactions[tx.IDValue()]; exists {
		return errors.New("transacao ja existe na mempool")
	}

	m.transactions[tx.IDValue()] = tx
	return nil
}

// RevalidateAgainst aplica suporte leve a reorg: troca o estado canônico usado
// pela mempool e remove transações que ficaram inválidas neste novo contexto.
func (m *Mempool) RevalidateAgainst(state *AccountState) (keptIDs []string, removedIDs []string, err error) {
	if state == nil {
		return nil, nil, errors.New("estado nulo")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = state

	for id, tx := range m.transactions {
		if err := tx.Validate(state); err != nil {
			delete(m.transactions, id)
			removedIDs = append(removedIDs, id)
			continue
		}
		keptIDs = append(keptIDs, id)
	}

	sort.Strings(keptIDs)
	sort.Strings(removedIDs)
	return keptIDs, removedIDs, nil
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
