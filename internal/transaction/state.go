package transaction

import (
	"errors"
	"sync"
)

// Account representa a posse atual de NFTs por uma chave publica.
type Account struct {
	NFTs map[string]bool
}

// Clone copia a conta.
func (a *Account) Clone() *Account {
	clonedNFTs := make(map[string]bool)
	for k, v := range a.NFTs {
		clonedNFTs[k] = v
	}
	return &Account{
		NFTs: clonedNFTs,
	}
}

// AccountState mantem o estado de posse de NFTs.
type AccountState struct {
	mu           sync.RWMutex
	Accounts     map[string]*Account
	ExistingNFTs map[string]bool
}

func NewAccountState() *AccountState {
	return &AccountState{
		Accounts:     make(map[string]*Account),
		ExistingNFTs: make(map[string]bool),
	}
}

// Clone copia todo o estado.
func (s *AccountState) Clone() *AccountState {
	s.mu.RLock()
	defer s.mu.RUnlock()

	clonedAccounts := make(map[string]*Account)
	for k, v := range s.Accounts {
		clonedAccounts[k] = v.Clone()
	}

	clonedExistingNFTs := make(map[string]bool)
	for k, v := range s.ExistingNFTs {
		clonedExistingNFTs[k] = v
	}

	return &AccountState{
		Accounts:     clonedAccounts,
		ExistingNFTs: clonedExistingNFTs,
	}
}

func (s *AccountState) getAccount(pubKey string) (*Account, bool) {
	acc, exists := s.Accounts[pubKey]
	return acc, exists
}

func (s *AccountState) CreateAccount(pubKey string) (*Account, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if pubKey == "" {
		return nil, errors.New("chave publica vazia")
	}
	if _, exists := s.Accounts[pubKey]; exists {
		return nil, errors.New("conta ja existe")
	}

	acc := &Account{NFTs: make(map[string]bool)}
	s.Accounts[pubKey] = acc
	return acc, nil
}

func (s *AccountState) ValidateTransferNFT(senderPubKey, recipientPubKey, nftID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	senderAcc, exists := s.getAccount(senderPubKey)
	if !exists {
		return errors.New("conta remetente inexistente")
	}
	if _, exists := s.getAccount(recipientPubKey); !exists {
		return errors.New("conta destinatario inexistente")
	}
	if !senderAcc.NFTs[nftID] {
		return errors.New("remetente nao possui este NFT")
	}
	return nil
}

func (s *AccountState) ValidateMintNFT(adminKey, recipientPubKey, nftID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if _, exists := s.getAccount(adminKey); !exists {
		return errors.New("conta admin inexistente")
	}
	if _, exists := s.getAccount(recipientPubKey); !exists {
		return errors.New("conta destinatario inexistente")
	}
	if s.ExistingNFTs[nftID] {
		return errors.New("este NFT já foi mintado")
	}

	return nil
}

func (s *AccountState) ApplyTransferNFT(senderPubKey, recipientPubKey, nftID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	senderAcc, senderExists := s.getAccount(senderPubKey)
	if !senderExists {
		return errors.New("conta remetente inexistente")
	}
	recipientAcc, recipientExists := s.getAccount(recipientPubKey)
	if !recipientExists {
		return errors.New("conta destinatario inexistente")
	}

	delete(senderAcc.NFTs, nftID)
	recipientAcc.NFTs[nftID] = true

	return nil
}

func (s *AccountState) ApplyMintNFT(adminKey, recipientPubKey, nftID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.getAccount(adminKey); !exists {
		return errors.New("conta admin inexistente")
	}
	s.ExistingNFTs[nftID] = true

	recipientAcc, exists := s.getAccount(recipientPubKey)
	if !exists {
		return errors.New("conta destinatario inexistente")
	}
	recipientAcc.NFTs[nftID] = true

	return nil
}

func (s *AccountState) Stats() (accounts int, nfts int) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.Accounts), len(s.ExistingNFTs)
}
