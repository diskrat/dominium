package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"encoding/binary"
	"errors"
	"time"

	"dominium/pkg/crypto"
)

// Transaction representa uma transacao assinada de NFT.
type Transaction struct {
	ID        string `json:"id"`
	Timestamp int64  `json:"timestamp"`
	Sig       []byte `json:"signature"`
	Type      byte   `json:"type"`
	PublKey   string `json:"publicKey"`
	Recipient string `json:"recipient"`
	NFTID     string `json:"nft_id"`
}

func NewTransaction(publKey, recipient, nftID string, transactionType byte) (*Transaction, error) {
	return NewTransactionWithTimestamp(publKey, recipient, nftID, transactionType, time.Now().UnixNano())
}

func NewTransactionWithTimestamp(publKey, recipient, nftID string, transactionType byte, timestamp int64) (*Transaction, error) {
	tx := &Transaction{
		Timestamp: timestamp,
		Type:      transactionType,
		PublKey:   publKey,
		Recipient: recipient,
		NFTID:     nftID,
	}
	return tx, nil
}

// Serialize retorna uma representacao canonica da transacao sem ID/assinatura.
func (tx *Transaction) Serialize() ([]byte, error) {
	if tx == nil {
		return nil, errors.New("transacao nula")
	}

	buf := bytes.NewBuffer(nil)
	if err := buf.WriteByte(tx.Type); err != nil {
		return nil, err
	}
	if err := binary.Write(buf, binary.BigEndian, tx.Timestamp); err != nil {
		return nil, err
	}
	if err := writeString(buf, tx.PublKey); err != nil {
		return nil, err
	}
	if err := writeString(buf, tx.Recipient); err != nil {
		return nil, err
	}
	if err := writeString(buf, tx.NFTID); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// VerifySerialized valida a assinatura para um payload ja serializado.
func VerifySerialized(publicKeyHex string, payload []byte, signature []byte) error {
	if publicKeyHex == "" || len(payload) == 0 || len(signature) == 0 {
		return errors.New("assinatura invalida")
	}
	pubKey, err := DecodePublicKey(publicKeyHex)
	if err != nil || !VerifyECDSA(pubKey, payload, signature) {
		return errors.New("assinatura invalida")
	}
	return nil
}

func writeString(buf *bytes.Buffer, value string) error {
	data := []byte(value)
	if err := binary.Write(buf, binary.BigEndian, uint32(len(data))); err != nil {
		return err
	}
	_, err := buf.Write(data)
	return err
}

// ComputeID calcula o ID canonicamente a partir do payload serializado.
func (tx *Transaction) ComputeID() (string, error) {
	if tx == nil {
		return "", errors.New("transacao nula")
	}

	payload, err := tx.Serialize()
	if err != nil {
		return "", err
	}

	return crypto.Hash(payload), nil
}

func (tx *Transaction) SetID() error {
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
		if err := tx.SetID(); err != nil {
			return err
		}
	}

	payload, err := tx.Serialize()
	if err != nil {
		return err
	}

	signature, err := SignECDSA(sk, payload)
	if err != nil {
		return err
	}
	tx.Sig = signature
	return nil
}

func (tx *Transaction) GetID() string {
	if tx == nil {
		return ""
	}
	return tx.ID
}
