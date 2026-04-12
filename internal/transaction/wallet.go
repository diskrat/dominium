package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/hex"
	"errors"
	"strings"
	"sync"
)

// WalletIdentity representa uma carteira simples associada a um nome humano.
type WalletIdentity struct {
	Name       string
	PublicKey  string
	PrivateKey string
}

// AccountList guarda identidades em memoria para facilitar visualizacao de transferencias.
type AccountList struct {
	mu          sync.RWMutex
	byName      map[string]*WalletIdentity
	byPublicKey map[string]*WalletIdentity
}

func NewAccountList() *AccountList {
	return &AccountList{
		byName:      make(map[string]*WalletIdentity),
		byPublicKey: make(map[string]*WalletIdentity),
	}
}

func (r *AccountList) CreateIdentity(name string) (*WalletIdentity, error) {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil, errors.New("nome nao pode ser vazio")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.byName[name]; exists {
		return nil, errors.New("nome ja cadastrado")
	}

	privKey, pubKey, err := GenerateKeyPair()
	if err != nil {
		return nil, err
	}

	encodedPriv, err := EncodePrivateKey(privKey)
	if err != nil {
		return nil, err
	}

	identity := &WalletIdentity{
		Name:       name,
		PublicKey:  EncodePublicKey(pubKey),
		PrivateKey: encodedPriv,
	}

	r.byName[name] = identity
	r.byPublicKey[identity.PublicKey] = identity
	return identity, nil
}

func (r *AccountList) GetByName(name string) (*WalletIdentity, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	identity, ok := r.byName[name]
	return identity, ok
}

func (r *AccountList) NameByPublicKey(publicKey string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	identity, ok := r.byPublicKey[publicKey]
	if !ok {
		return ""
	}
	return identity.Name
}

// SignTransactionAs assina uma transacao usando a identidade registrada.
func (r *AccountList) SignTransactionAs(name string, tx *Transaction) error {
	if tx == nil {
		return errors.New("transacao nula")
	}

	r.mu.RLock()
	identity, ok := r.byName[name]
	r.mu.RUnlock()
	if !ok {
		return errors.New("identidade nao encontrada")
	}

	if tx.PublKey != identity.PublicKey {
		return errors.New("transacao pertence a outra conta")
	}

	privKey, err := DecodePrivateKey(identity.PrivateKey)
	if err != nil {
		return err
	}

	return tx.Sign(privKey)
}

// EncodePrivateKey codifica a chave privada em hex para persistencia simples.
func EncodePrivateKey(priv *ecdsa.PrivateKey) (string, error) {
	if priv == nil {
		return "", errors.New("chave privada nula")
	}

	der, err := x509.MarshalECPrivateKey(priv)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(der), nil
}

// DecodePrivateKey decodifica a chave privada em hex.
func DecodePrivateKey(privHex string) (*ecdsa.PrivateKey, error) {
	privBytes, err := hex.DecodeString(privHex)
	if err != nil {
		return nil, err
	}

	parsed, err := x509.ParseECPrivateKey(privBytes)
	if err != nil {
		return nil, err
	}

	if parsed.Curve == nil || parsed.Curve.Params() == nil {
		return nil, errors.New("curva da chave privada invalida")
	}
	if parsed.Curve.Params().Name != elliptic.P256().Params().Name {
		return nil, errors.New("somente curva P-256 eh suportada")
	}

	return parsed, nil
}
