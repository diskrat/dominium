// internal\core\block.go
package core

import (
	"sort"
	"time"
)

type Header struct {
	HashOfPrevious     string
	MerkleRootHash     string
	Timestamp          int64
	Nbits              int32 //dificuldade
	Nonce              int32
	TransactionCounter uint16
	Hash               string
}

type Transaction struct {
	TXid string
	Data string
	Fee  int64
	Time int64
}

type Body struct {
	Transactions []Transaction
}

type Block struct {
	Header
	Body
}

// fabricamos a transacao e mandamos para a mempool
func NewAndPostTransaction(m *Mempool, id string, data string, fee int64) *Transaction {
	//recebemos e empacotamos
	tx := &Transaction{
		TXid: id,
		Data: data,
		Fee:  fee,
	}

	//adicionamos na mempool
	m.Adicionar(*tx)

	//retorno para quem chamou nao ficar no escuro
	return tx
}

//extrai transações da mempool e prepara o Bloco
func (m *Mempool) MontarProximoBloco(prevHash string, dificuldade int) *Block {
	m.mu.RLock() 
	defer m.mu.RUnlock()

	var selecionadas []Transaction
	
    //puxa todas as transações da mempool para uma slice
	for _, tx := range m.Transactions {
		selecionadas = append(selecionadas, tx)
	}

    // ordenamos a lista com base no fee (do maior para o menor)
	sort.Slice(selecionadas, func(i, j int) bool {
		return selecionadas[i].Fee > selecionadas[j].Fee
	})

    //cortamos a lista para ter no máximo 10 transações (valor que limitei para testes) -> as que pagam mais
    if len(selecionadas) > 10 {
        selecionadas = selecionadas[:10]
    }

	//apos preparar as transacoes escolhidas do bloco, agrupamos no body do bloco
	corpo := Body{Transactions: selecionadas}

	//montamos o restante da estrutura do bloco e mineramos com o gerarmerkleroot
	cabecalho := Header{
		HashOfPrevious:     prevHash, //se for genesis 0, se nao, colocar o ultimo da cadeia
		MerkleRootHash:     corpo.GerarMerkleRoot(),
		Timestamp:          time.Now().Unix(), //gera hora atual
		Nbits:              int32(dificuldade), //numero de zeros para minerar
		TransactionCounter: uint16(len(selecionadas)), 
		Nonce:              0,  //init
	}

	//usa-se ponteiro para n criar copias desnecessarias e passar o valor por referencia
	return &Block{
		Header: cabecalho,
		Body:   corpo,
	}
}