package transaction

import (
	"errors"
	"sync"
)

type Account struct {
	NFTs map[string]bool
}

func (a *Account) Clone() *Account {
	clonedNFTs := make(map[string]bool)
	for k, v := range a.NFTs {
		clonedNFTs[k] = v
	}
	return &Account{
		NFTs: clonedNFTs,
	}
}

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

func (s *AccountState) getOrCreateAccount(pubKey string) *Account {
	acc, exists := s.Accounts[pubKey]
	if exists {
		return acc
	}

	acc = &Account{
		NFTs: make(map[string]bool),
	}
	s.Accounts[pubKey] = acc
	return acc
}

func (s *AccountState) ValidateNFT(pubKey string, nftID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acc, exists := s.getAccount(pubKey)
	if !exists {
		return errors.New("conta remetente inexistente")
	}
	if !acc.NFTs[nftID] {
		return errors.New("remetente nao possui este NFT")
	}
	return nil
}
func (s *AccountState) ValidateMintNFT(adminKey string, nftID string) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.ExistingNFTs[nftID] {
		return errors.New("este NFT já foi mintado")
	}

	return nil
}

func (s *AccountState) ApplyTransferNFT(senderPubKey, recipientPubKey, nftID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	senderAcc := s.getOrCreateAccount(senderPubKey)
	recipientAcc := s.getOrCreateAccount(recipientPubKey)

	delete(senderAcc.NFTs, nftID)
	recipientAcc.NFTs[nftID] = true

	return nil
}

func (s *AccountState) ApplyMintNFT(adminKey, recipientPubKey, nftID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	_ = s.getOrCreateAccount(adminKey)
	s.ExistingNFTs[nftID] = true

	recipientAcc := s.getOrCreateAccount(recipientPubKey)
	recipientAcc.NFTs[nftID] = true

	return nil
}
