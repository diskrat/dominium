package transaction

import (
	"testing"
)

func TestMempoolPruneInvalidAfterStateChange(t *testing.T) {
	priv, publKey := TestKeyPair()
	_, recipientPub := TestKeyPair()

	state := NewAccountState()
	if err := state.EnsureAccount(publKey); err != nil {
		t.Fatalf("failed to ensure sender account: %v", err)
	}
	if err := state.EnsureAccount(recipientPub); err != nil {
		t.Fatalf("failed to ensure recipient account: %v", err)
	}
	state.Accounts[publKey].NFTs["nft-001"] = true

	mempool := NewMempool(state)
	tx, err := NewTransaction(publKey, recipientPub, "nft-001", TypeTransferNFT)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}
	if err := tx.Sign(priv); err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}

	if err := mempool.Add(tx); err != nil {
		t.Fatalf("failed to add valid transaction to mempool: %v", err)
	}

	state.Accounts[publKey].NFTs["nft-001"] = false

	removed := mempool.PruneInvalid()
	if len(removed) != 1 {
		t.Fatalf("expected 1 invalid transaction removed, got %d", len(removed))
	}
	if _, exists := mempool.transactions[tx.GetID()]; exists {
		t.Fatal("expected invalid transaction to be removed from mempool")
	}
}
