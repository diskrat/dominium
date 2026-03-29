package transaction

import (
	"crypto/ecdsa"
	"dominium/pkg/crypto"
	"encoding/json"
	"time"
)

type Transaction struct {
	ID        string
	Timestamp int64
	PublKey   string
	Nonce     uint64
	Sig       []byte
	Type      byte
	Data      json.RawMessage
}

func NewTransaction(publKey string, transactionType byte, nonce uint64, data any) (*Transaction, error) {

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	tx := &Transaction{
		Timestamp: time.Now().UnixNano(),
		PublKey:   publKey,
		Nonce:     nonce,
		Type:      transactionType,
		Data:      dataBytes,
	}
	return tx, nil
}

func (tx *Transaction) IdCalc() error {
	txCopy := *tx
	txCopy.ID = ""
	txCopy.Sig = nil
	hash, err := crypto.HashObject(txCopy)
	if err != nil {
		return err
	}
	tx.ID = hash
	return nil
}

func (tx *Transaction) Sign(sk *ecdsa.PrivateKey) error {
	if tx.ID == "" {
		if err := tx.IdCalc(); err != nil {
			return err
		}
	}

	idBytes := []byte(tx.ID)

	signature, err := SignECDSA(sk, idBytes)
	if err != nil {
		return err
	}
	tx.Sig = signature
	return nil
}
