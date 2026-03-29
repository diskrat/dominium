package transaction

import (
	"errors"
	"sync"
)

type Account struct {
	NFTs  map[string]bool
	Nonce uint64
}

func (a *Account) Clone() *Account {
	clonedNFTs := make(map[string]bool)
	for k, v := range a.NFTs {
		clonedNFTs[k] = v
	}
	return &Account{
		NFTs:  clonedNFTs,
		Nonce: a.Nonce,
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

func (s *AccountState) getAccount(pubKey string) *Account {
	acc, exists := s.Accounts[pubKey]
	if !exists {
		acc = &Account{
			NFTs:  make(map[string]bool),
			Nonce: 0,
		}
		s.Accounts[pubKey] = acc
	}
	return acc
}

func (s *AccountState) ValidateNFT(pubKey string, nftID string, nonce uint64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	acc := s.getAccount(pubKey)
	if !acc.NFTs[nftID] {
		return errors.New("remetente nao possui este NFT")
	}
	if nonce != acc.Nonce+1 {
		return errors.New("nonce invalido")
	}
	return nil
}
func (s *AccountState) ValidateMintNFT(adminKey string, nftID string, nonce uint64) error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.ExistingNFTs[nftID] {
		return errors.New("este NFT já foi mintado")
	}

	acc := s.getAccount(adminKey)
	if nonce != acc.Nonce+1 {
		return errors.New("nonce invalido")
	}

	return nil
}

func (s *AccountState) ApplyTransferNFT(senderPubKey, recipientPubKey, nftID string, nonce uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	senderAcc := s.getAccount(senderPubKey)
	if !senderAcc.NFTs[nftID] {
		return errors.New("remetente nao possui este NFT")
	}
	if nonce != senderAcc.Nonce+1 {
		return errors.New("nonce invalido")
	}

	recipientAcc := s.getAccount(recipientPubKey)

	delete(senderAcc.NFTs, nftID)
	senderAcc.Nonce++

	recipientAcc.NFTs[nftID] = true

	return nil
}

func (s *AccountState) ApplyMintNFT(adminKey, recipientPubKey, nftID string, nonce uint64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.ExistingNFTs[nftID] {
		return errors.New("este NFT já foi mintado")
	}

	adminAcc := s.getAccount(adminKey)
	if nonce != adminAcc.Nonce+1 {
		return errors.New("nonce invalido")
	}

	adminAcc.Nonce++
	s.ExistingNFTs[nftID] = true

	recipientAcc := s.getAccount(recipientPubKey)
	recipientAcc.NFTs[nftID] = true

	return nil
}
