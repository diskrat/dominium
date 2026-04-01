package transaction

import (
	"errors"
)

const (
	TypeTransferNFT byte = 1
	TypeMintNFT     byte = 2
)

const AdminPubKey = "somekey"

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

	pubKey, err := DecodePublicKey(tx.PublKey)
	if err != nil || !VerifyECDSA(pubKey, []byte(tx.ID), tx.Sig) {
		return errors.New("assinatura invalida")
	}
	if tx.NFTID == "" {
		return errors.New("ID do NFT nao pode ser vazio")
	}
	if tx.Recipient == "" {
		return errors.New("recipient nao pode ser vazio")
	}

	switch tx.Type {

	case TypeTransferNFT:
		return state.ValidateNFT(tx.PublKey, tx.NFTID)
	case TypeMintNFT:

		if tx.PublKey != AdminPubKey {
			return errors.New("acesso negado: apenas o admin pode fazer mint")
		}
		return state.ValidateMintNFT(tx.PublKey, tx.NFTID)
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
