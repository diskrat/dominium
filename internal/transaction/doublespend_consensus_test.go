package transaction

import (
	"testing"
)

// TestDoubleSpendConsensus ensures consensus rules reject double-spend attempts.
func TestDoubleSpendConsensus(t *testing.T) {
	state := NewAccountState()
	adminPriv, adminPub := TestKeyPair()
	user1Priv, user1Pub := TestKeyPair()
	_, user2Pub := TestKeyPair()

	state.EnsureAccount(user2Pub)
	state.EnsureAccount("attacker")

	// Mint NFT to user1
	mintTx, _ := NewTransaction(adminPub, user1Pub, "nft-cs", TypeMintNFT)
	if err := mintTx.Sign(adminPriv); err != nil {
		t.Fatalf("failed to sign mint tx: %v", err)
	}
	if err := mintTx.Execute(state); err != nil {
		t.Fatalf("mint execute failed: %v", err)
	}

	// First transfer
	dsTx1, _ := NewTransaction(user1Pub, user2Pub, "nft-cs", TypeTransferNFT)
	if err := dsTx1.Sign(user1Priv); err != nil {
		t.Fatalf("failed to sign first transfer tx: %v", err)
	}
	if err := dsTx1.ValidateConsensusRules(state, adminPub); err != nil {
		t.Fatalf("first transfer should be valid: %v", err)
	}
	if err := dsTx1.Execute(state); err != nil {
		t.Fatalf("first transfer execute failed: %v", err)
	}

	// Second (conflicting) transfer
	dsTx2, _ := NewTransaction(user1Pub, "attacker", "nft-cs", TypeTransferNFT)
	if err := dsTx2.Sign(user1Priv); err != nil {
		t.Fatalf("failed to sign second transfer tx: %v", err)
	}
	if err := dsTx2.ValidateConsensusRules(state, adminPub); err == nil {
		t.Error("double-spend transfer should be rejected by consensus rules")
	}
}
