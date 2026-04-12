package miner

import (
	"bytes"
	"encoding/binary"
	"time"

	"dominium/internal/transaction"
)

// Header contém os metadados do bloco para o processo de mineração e encadeamento.
type Header struct {
	HashOfPrevious     []byte `json:"hash_of_previous"`
	MerkleRootHash     []byte `json:"merkle_root_hash"`
	Timestamp          int64  `json:"timestamp"`
	Nbits              int32  `json:"nbits"`
	Nonce              int32  `json:"nonce"`
	TransactionCounter uint16 `json:"transaction_counter"`
	Hash               []byte `json:"hash"`
}

// Body contém as transações validadas incluídas no bloco.
type Body struct {
	Transactions []transaction.Transaction `json:"transactions"`
}

// Block representa um bloco completo na blockchain Dominium.
type Block struct {
	Header `json:"header"`
	Body   `json:"body"`
	Miner  string `json:"miner"`
}

// Serialize converte o Header em um slice de bytes para a geração do Hash.
// A propriedade Hash em si é omitida da serialização, pois é o resultado desse processo.
func (h *Header) Serialize() ([]byte, error) {
	buf := bytes.NewBuffer(nil)

	if err := binary.Write(buf, binary.BigEndian, h.HashOfPrevious); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.MerkleRootHash); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.Timestamp); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.Nbits); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.Nonce); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, h.TransactionCounter); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// NewBlock cria um novo bloco, inicializa o header e calcula a raiz de Merkle.
func NewBlock(prevHash []byte, txs []transaction.Transaction, nBits int32, minerID string) *Block {
	return &Block{
		Header: Header{
			HashOfPrevious:     prevHash,
			Timestamp:          time.Now().UnixNano(),
			Nbits:              nBits,
			TransactionCounter: uint16(len(txs)),
			MerkleRootHash:     CalculateMerkleRoot(txs),
		},
		Body: Body{
			Transactions: txs,
		},
		Miner: minerID,
	}
}