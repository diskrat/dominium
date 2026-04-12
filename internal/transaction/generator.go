package transaction

import (
	"encoding/hex"
	"errors"
	"math/rand"
	"time"
)

// Generator cria transacoes assinadas usando uma fonte deterministica.
type Generator struct {
	rng      *rand.Rand
	accounts *AccountList
	admin    *WalletIdentity
}

// NewGenerator cria um gerador deterministico com a seed informada.
func NewGenerator(seed int64) *Generator {
	if seed == 0 {
		seed = time.Now().UnixNano()
	}
	return &Generator{
		rng:      rand.New(rand.NewSource(seed)),
		accounts: NewAccountList(),
	}
}

func (g *Generator) Accounts() *AccountList {
	return g.accounts
}

func (g *Generator) CreateIdentity(name string) (*WalletIdentity, error) {
	return g.accounts.CreateIdentity(name)
}

// SetAdmin registra a identidade admin usada para mint.
func (g *Generator) SetAdmin(identity *WalletIdentity) error {
	if identity == nil {
		return errors.New("identidade admin nula")
	}
	if identity.PrivateKey == "" || identity.PublicKey == "" {
		return errors.New("identidade admin invalida")
	}
	g.admin = identity
	return nil
}

// RandomNFTID cria um identificador deterministico de NFT.
func (g *Generator) RandomNFTID() string {
	b := make([]byte, 8)
	_, _ = g.rng.Read(b)
	return hex.EncodeToString(b)
}

// SignWithIdentity assina a transacao usando a identidade informada.
func (g *Generator) SignWithIdentity(identity *WalletIdentity, tx *Transaction) error {
	if identity == nil {
		return errors.New("identidade nula")
	}
	privKey, err := DecodePrivateKey(identity.PrivateKey)
	if err != nil {
		return err
	}
	return tx.Sign(privKey)
}

// MintTx cria uma transacao de mint assinada pelo admin.
func (g *Generator) MintTx(recipient *WalletIdentity, nftID string) (*Transaction, error) {
	if g.admin == nil {
		return nil, errors.New("admin nao configurado")
	}
	if recipient == nil {
		return nil, errors.New("recipient nulo")
	}
	if nftID == "" {
		nftID = g.RandomNFTID()
	}

	tx, err := NewTransaction(g.admin.PublicKey, recipient.PublicKey, nftID, TypeMintNFT)
	if err != nil {
		return nil, err
	}
	if err := g.SignWithIdentity(g.admin, tx); err != nil {
		return nil, err
	}
	return tx, nil
}

// TransferTx cria uma transacao de transferencia assinada pelo remetente.
func (g *Generator) TransferTx(sender *WalletIdentity, recipient *WalletIdentity, nftID string) (*Transaction, error) {
	if sender == nil || recipient == nil {
		return nil, errors.New("identidade nula")
	}
	if nftID == "" {
		return nil, errors.New("nftID vazio")
	}

	tx, err := NewTransaction(sender.PublicKey, recipient.PublicKey, nftID, TypeTransferNFT)
	if err != nil {
		return nil, err
	}
	if err := g.SignWithIdentity(sender, tx); err != nil {
		return nil, err
	}
	return tx, nil
}
