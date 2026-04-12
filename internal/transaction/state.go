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

func (s *AccountState) EnsureAccount(pubKey string) error {
	if pubKey == "" {
		return errors.New("chave publica vazia")
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.Accounts[pubKey]; exists {
		return nil
	}

	s.Accounts[pubKey] = &Account{NFTs: make(map[string]bool)}
	return nil
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

    if adminKey == "" || recipientPubKey == "" {
        return errors.New("chaves admin e destinatario sao obrigatorias")
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

    // Garante que a conta do Admin exista
    if _, exists := s.Accounts[adminKey]; !exists {
        s.Accounts[adminKey] = &Account{NFTs: make(map[string]bool)}
    }

    // Garante que a conta do Destinatário exista (Auto-registro)
    if _, exists := s.Accounts[recipientPubKey]; !exists {
        s.Accounts[recipientPubKey] = &Account{NFTs: make(map[string]bool)}
    }

    s.ExistingNFTs[nftID] = true
    
    // Agora o mapa de NFTs existe para o destinatário, podemos atribuir o NFT a ele
    s.Accounts[recipientPubKey].NFTs[nftID] = true

    return nil
}

// GetAllAccounts retorna uma cópia de segurança do mapa de contas para exportação na API.
func (s *AccountState) GetAllAccounts() map[string]*Account {
	s.mu.RLock()
	defer s.mu.RUnlock()
	
	// Criamos uma cópia para evitar problemas de concorrência no acesso direto
	cloned := make(map[string]*Account)
	for k, v := range s.Accounts {
		cloned[k] = v.Clone()
	}
	return cloned
}

// GetNFTsList retorna a lista de IDs de NFTs como um slice de strings (formato que o React espera).
func (a *Account) GetNFTsList() []string {
	list := make([]string, 0, len(a.NFTs))
	for nftID := range a.NFTs {
		list = append(list, nftID)
	}
	return list
}