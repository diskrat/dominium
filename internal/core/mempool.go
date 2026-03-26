// internal\core\mempool.go
package core

import (
	"fmt"
	"log"
	"sync"
)

type Mempool struct {
	mu sync.RWMutex //mempool pode ser chamada varias vezes pelos mineradores sem quebrar, por isso usei essa funcao
	Transactions map[string]Transaction //cria uma chave valor dentro da estrutura da mempool para otimizar
}

//inicializacao
func NewMempool() *Mempool {
    return &Mempool{
        Transactions: make(map[string]Transaction),
    }
}

//adicionar transacao a mempool
func (m *Mempool) Adicionar(tx Transaction) error {
    //protege contra vários nós enviando ao mesmo tempo
	m.mu.Lock() 
    defer m.mu.Unlock()
	//verificar se o numero de transacoes da mempool chegou no max de transacoes em espera
	if len(m.Transactions) > 2500 {
    	return fmt.Errorf("Mempool lotada! Tente novamente mais tarde.")
	}

    //recemos um true caso a transacao adicionada ja esteja na fila (esquema: gasto duplo)
	_, ok := m.Transactions[tx.TXid]

    //verificamos se já existe no mapa
    if ok {
        return fmt.Errorf("transação %s já está na fila", tx.TXid)
    }

    //Se não existe, adicionamos
    m.Transactions[tx.TXid] = tx

	log.Printf("[MEMPOOL] Transação aceita: %s | info: %s | Fee: %d | Total: %d", 
        tx.TXid, tx.Data, tx.Fee, len(m.Transactions))
    return nil
}

//limpar transacoes quando o minerador acha um bloco
func (m *Mempool) LimparConfirmadas(txids []string) {
    //trancamos a msempool para evitar que novos registros entrem enquanto deletando.
    m.mu.Lock()
    defer m.mu.Unlock()

    for _, id := range txids {
        //função delete remove a chave do mapa.
        //se o id não existir (por algum erro), o go ignora silenciosamente
        delete(m.Transactions, id)
    }
}