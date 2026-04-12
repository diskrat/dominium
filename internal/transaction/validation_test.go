package transaction

import (
	"testing"
)

func TestValidateRejectsMissingSignature(t *testing.T) {
	_, publKey := TestKeyPair()
	tx, err := NewTransaction(publKey, "recipient", "nft-001", TypeMintNFT)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}
	tx.Sig = nil

	state := NewAccountState()
	if err := tx.Validate(state); err == nil {
		t.Fatal("expected Validate to reject transaction without signature")
	}
}

func TestValidateRejectsTamperedSignature(t *testing.T) {
	priv, publKey := TestKeyPair()
	tx, err := NewTransaction(publKey, "recipient", "nft-002", TypeMintNFT)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}
	if err := tx.Sign(priv); err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}

	tx.Sig[0] ^= 0xFF

	state := NewAccountState()
	if err := tx.Validate(state); err == nil {
		t.Fatal("expected Validate to reject transaction with tampered signature")
	}
}
