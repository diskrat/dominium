package transaction

import (
	"errors"
)

const (
	TypeTransferNFT byte = 1
	TypeMintNFT     byte = 2
)

// ValidateConsensusRules valida regras específicas de consenso além da validação básica
func (tx *Transaction) ValidateConsensusRules(state *AccountState, adminPubKey string) error {
	if tx == nil {
		return errors.New("transacao nula")
	}
	if state == nil {
		return errors.New("estado nulo")
	}

	switch tx.Type {
	case TypeMintNFT:
		// Regra de autoridade: apenas admin pode mintar
		if tx.PublKey != adminPubKey {
			return errors.New("apenas admin pode executar mint")
		}
		// Regra de unicidade: NFT não pode já existir
		if state.ExistingNFTs[tx.NFTID] {
			return errors.New("NFT ja foi mintado")
		}

	case TypeTransferNFT:
		// Regra de posse: sender deve ter o NFT
		senderAcc, exists := state.Accounts[tx.PublKey]
		if !exists {
			return errors.New("conta do sender inexistente")
		}
		if !senderAcc.NFTs[tx.NFTID] {
			return errors.New("sender nao possui este NFT")
		}

	default:
		return errors.New("tipo de transacao desconhecido")
	}

	return nil
}

func (tx *Transaction) Validate(state *AccountState) error {
	if tx == nil {
		return errors.New("transacao nula")
	}
	if state == nil {
		return errors.New("estado nulo")
	}
	if tx.ID == "" {
		return errors.New("ID da transacao ausente")
	}
	if len(tx.Sig) == 0 {
		return errors.New("assinatura ausente")
	}

	expectedID, err := tx.ComputeID()
	if err != nil {
		return err
	}
	if expectedID != tx.ID {
		return errors.New("integridade invalida: ID nao corresponde ao conteudo")
	}

	payload, err := tx.Serialize()
	if err != nil {
		return errors.New("assinatura invalida")
	}
	if err := VerifySerialized(tx.PublKey, payload, tx.Sig); err != nil {
		return err
	}
	if tx.NFTID == "" {
		return errors.New("ID do NFT nao pode ser vazio")
	}
	if tx.Recipient == "" {
		return errors.New("recipient nao pode ser vazio")
	}

	switch tx.Type {

	case TypeTransferNFT:
		return state.ValidateTransferNFT(tx.PublKey, tx.Recipient, tx.NFTID)
	case TypeMintNFT:
		return state.ValidateMintNFT(tx.PublKey, tx.Recipient, tx.NFTID)
	default:
		return errors.New("tipo de transacao desconhecido")
	}
}

func (tx *Transaction) Execute(state *AccountState) error {

	if err := tx.Validate(state); err != nil {
		return err
	}

	switch tx.Type {

	case TypeTransferNFT:
		return state.ApplyTransferNFT(tx.PublKey, tx.Recipient, tx.NFTID)

	case TypeMintNFT:
		return state.ApplyMintNFT(tx.PublKey, tx.Recipient, tx.NFTID)

	default:
		return errors.New("tipo de transacao desconhecido")
	}
}
