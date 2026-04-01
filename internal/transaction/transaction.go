package transaction

import (
	"crypto/ecdsa"
	"dominium/pkg/crypto"
	"errors"
	"time"
)

type Transaction struct {
	ID        string
	Timestamp int64
	PublKey   string
	Recipient string
	NFTID     string
	Sig       []byte
	Type      byte
}

// Tx define o contrato minimo para validacao e execucao de transacoes.
type Tx interface {
	IDValue() string
	Validate(state *AccountState) error
	Execute(state *AccountState) error
}

func NewTransaction(publKey string, recipient string, nftID string, transactionType byte) (*Transaction, error) {
	tx := &Transaction{
		Timestamp: time.Now().UnixNano(),
		PublKey:   publKey,
		Recipient: recipient,
		NFTID:     nftID,
		Type:      transactionType,
	}
	return tx, nil
}

func (tx *Transaction) ComputeID() (string, error) {
	if tx == nil {
		return "", errors.New("transacao nula")
	}

	txCopy := *tx
	txCopy.ID = ""
	txCopy.Sig = nil

	hash, err := crypto.HashObject(txCopy)
	if err != nil {
		return "", err
	}

	return hash, nil
}

func (tx *Transaction) IdCalc() error {
	hash, err := tx.ComputeID()
	if err != nil {
		return err
	}
	tx.ID = hash
	return nil
}

func (tx *Transaction) Sign(sk *ecdsa.PrivateKey) error {
	if tx == nil {
		return errors.New("transacao nula")
	}
	if sk == nil {
		return errors.New("chave privada nula")
	}

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

func (tx *Transaction) IDValue() string {
	if tx == nil {
		return ""
	}
	return tx.ID
}
