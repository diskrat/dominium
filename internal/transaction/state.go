package transaction

import (
	"errors"
	"sync"
)

// Account guarda os ativos e o controle de repetição de uma chave pública
type Account struct {
	Balance float64         // O token Fungível (ex: 100.5 moedas)
	NFTs    map[string]bool // Os tokens Não-Fungíveis (Mapeia o ID do NFT -> true)
	Nonce   uint64          // Contador contra Replay Attacks
}

type AccountState struct {
	mu           sync.RWMutex
	Accounts     map[string]*Account
	ExistingNFTs map[string]bool
}

func NewAccountState() *AccountState {
	return &AccountState{
		Accounts: make(map[string]*Account),
	}
}

// getAccount garante que a conta existe antes de tentarmos acessar mapas nela
func (s *AccountState) getAccount(pubKey string) *Account {
	acc, exists := s.Accounts[pubKey]
	if !exists {
		acc = &Account{
			Balance: 0,
			NFTs:    make(map[string]bool),
			Nonce:   0,
		}
		s.Accounts[pubKey] = acc
	}
	return acc
}

// ValidateNFT verifica se o usuário realmente possui aquele item único
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

	// O NFT já existe no mundo?
	if s.ExistingNFTs[nftID] {
		return errors.New("este NFT já foi mintado")
	}

	acc := s.getAccount(adminKey)
	if nonce != acc.Nonce+1 {
		return errors.New("nonce invalido")
	}

	return nil
}
