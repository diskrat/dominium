package transaction

import (
	"encoding/json"
	"errors"
)

const (
	TypeTransferNFT byte = 1
	TypeMintNFT     byte = 2
)
const AdminPubKey = "somekey"

// TransferNFTPayload transfere a posse de um ID único
type TransferNFTPayload struct {
	Recipient string `json:"recipient"`
	NFTID     string `json:"nft_id"`
}

type MintNFTPayload struct {
	Recipient string `json:"recipient"` // Para quem o NFT vai nascer
	NFTID     string `json:"nft_id"`    // O ID único do novo NFT
}

// Validate é a porta de entrada chamada pela Mempool
func (tx *Transaction) Validate(state *AccountState) error {
	pubKey, err := DecodePublicKey(tx.PublKey)
	if err != nil || !VerifyECDSA(pubKey, []byte(tx.ID), tx.Sig) {
		return errors.New("assinatura invalida")
	}

	switch tx.Type {

	case TypeTransferNFT:
		var payload TransferNFTPayload
		if err := json.Unmarshal(tx.Data, &payload); err != nil {
			return errors.New("formato de payload NFT invalido")
		}
		if payload.NFTID == "" {
			return errors.New("ID do NFT nao pode ser vazio")
		}
		return state.ValidateNFT(tx.PublKey, payload.NFTID, tx.Nonce)
	case TypeMintNFT:
		// 1. REGRA DE AUTORIZAÇÃO: Quem assinou a transação é o Admin?
		if tx.PublKey != AdminPubKey {
			return errors.New("acesso negado: apenas o admin pode fazer mint")
		}

		var payload MintNFTPayload
		if err := json.Unmarshal(tx.Data, &payload); err != nil {
			return errors.New("formato de payload Mint NFT invalido")
		}
		if payload.NFTID == "" {
			return errors.New("ID do NFT nao pode ser vazio")
		}

		// 2. Chama o estado para garantir que o NFT já não existe e checar o nonce
		return state.ValidateMintNFT(tx.PublKey, payload.NFTID, tx.Nonce)
	default:
		return errors.New("tipo de transacao desconhecido")
	}
}
