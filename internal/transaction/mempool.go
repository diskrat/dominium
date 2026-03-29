package transaction

import (
	"errors"
	"sync"
)

type Mempool struct {
	mu           sync.RWMutex
	transactions map[string]*Transaction
	state        *AccountState // <-- ADIÇÃO 1: A mempool agora guarda uma referência do estado
}

// ADIÇÃO 2: O construtor agora exige que você passe o estado global da blockchain
func NewMempool(state *AccountState) *Mempool {
	return &Mempool{
		transactions: make(map[string]*Transaction),
		state:        state,
	}
}

// ADIÇÃO 3: O método Add deixa de ser um "stub" e passa a fazer a validação real
func (m *Mempool) Add(tx *Transaction) error {
	// 1. Delega toda a validação complexa (Assinatura, Saldo, Nonce, Tipo) para os contratos
	if err := tx.Validate(m.state); err != nil {
		return err // Se a transação for inválida, ela é barrada aqui e não entra na fila
	}

	// 2. Trava o mapa para escrita
	m.mu.Lock()
	defer m.mu.Unlock()

	// 3. Garante que a transação já não foi enviada antes (evita duplicidade na fila)
	if _, exists := m.transactions[tx.ID]; exists {
		return errors.New("transacao ja existe na mempool")
	}

	// 4. Salva a transação na memória
	m.transactions[tx.ID] = tx
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
