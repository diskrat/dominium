package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"
)

func createSignedTransaction(from, to, assetID string, txType byte, timestamp int64, privateKey *ecdsa.PrivateKey) (*Transaction, error) {
	tx, err := NewTransactionWithTimestamp(from, to, assetID, txType, timestamp)
	if err != nil {
		return nil, err
	}

	if err := tx.Sign(privateKey); err != nil {
		return nil, err
	}
	return tx, nil
}

func TestConsensusRules(t *testing.T) {
	state := NewAccountState()

	// Gera chave privada para admin
	adminPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("erro ao gerar chave privada admin: %v", err)
	}
	adminPubKey := EncodePublicKey(&adminPrivateKey.PublicKey)

	// Gera chave privada para usuário normal
	userPrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("erro ao gerar chave privada usuario: %v", err)
	}
	userPubKey := EncodePublicKey(&userPrivateKey.PublicKey)

	// Gera chave privada para recipient1
	recipient1PrivateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("erro ao gerar chave privada recipient1: %v", err)
	}

	// Cria contas no estado
	_, err = state.CreateAccount(adminPubKey)
	if err != nil {
		t.Fatalf("erro ao criar conta admin: %v", err)
	}
	_, err = state.CreateAccount("recipient1")
	if err != nil {
		t.Fatalf("erro ao criar conta recipient1: %v", err)
	}
	_, err = state.CreateAccount("recipient2")
	if err != nil {
		t.Fatalf("erro ao criar conta recipient2: %v", err)
	}
	_, err = state.CreateAccount("recipient3")
	if err != nil {
		t.Fatalf("erro ao criar conta recipient3: %v", err)
	}

	// Testa mint válido
	mintTx, err := createSignedTransaction(adminPubKey, "recipient1", "nft-001", TypeMintNFT, time.Now().UnixNano(), adminPrivateKey)
	if err != nil {
		t.Fatalf("erro ao criar mint tx: %v", err)
	}

	if err := mintTx.ValidateConsensusRules(state, adminPubKey); err != nil {
		t.Errorf("mint valido deveria passar: %v", err)
	}

	// Executa o mint
	if err := mintTx.Execute(state); err != nil {
		t.Fatalf("erro ao executar mint: %v", err)
	}

	// Testa mint duplicado (deve falhar)
	mintTx2, err := createSignedTransaction(adminPubKey, "recipient2", "nft-001", TypeMintNFT, time.Now().UnixNano(), adminPrivateKey)
	if err != nil {
		t.Fatalf("erro ao criar segundo mint tx: %v", err)
	}

	if err := mintTx2.ValidateConsensusRules(state, adminPubKey); err == nil {
		t.Error("mint duplicado deveria falhar")
	}

	// Testa mint por não-admin (deve falhar)
	mintTx3, err := createSignedTransaction(userPubKey, "recipient3", "nft-002", TypeMintNFT, time.Now().UnixNano(), userPrivateKey)
	if err != nil {
		t.Fatalf("erro ao criar mint tx por nao-admin: %v", err)
	}

	if err := mintTx3.ValidateConsensusRules(state, adminPubKey); err == nil {
		t.Error("mint por nao-admin deveria falhar")
	}

	// Testa transfer válido
	transferTx, err := createSignedTransaction("recipient1", "recipient2", "nft-001", TypeTransferNFT, time.Now().UnixNano(), recipient1PrivateKey)
	if err != nil {
		t.Fatalf("erro ao criar transfer tx: %v", err)
	}

	if err := transferTx.ValidateConsensusRules(state, adminPubKey); err != nil {
		t.Errorf("transfer valido deveria passar: %v", err)
	}

	// Testa transfer de NFT que não possui (deve falhar)
	transferTx2, err := createSignedTransaction("recipient2", "recipient1", "nft-999", TypeTransferNFT, time.Now().UnixNano(), userPrivateKey)
	if err != nil {
		t.Fatalf("erro ao criar transfer tx invalido: %v", err)
	}

	if err := transferTx2.ValidateConsensusRules(state, adminPubKey); err == nil {
		t.Error("transfer de NFT nao possuido deveria falhar")
	}
}