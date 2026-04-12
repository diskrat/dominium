package miner

import (
	"dominium/internal/transaction"
	"testing"
)

// TestDoubleSpendInBlock ensures that two conflicting transactions for the same NFT are not both accepted in the same block.
func TestDoubleSpendInBlock(t *testing.T) {
	bc := NewBlockchain()
	adminPriv, adminPub := transaction.TestKeyPair()
	user1Priv, user1Pub := transaction.TestKeyPair()
	_, user2Pub := transaction.TestKeyPair()

	// Mint NFT to user1
	mintTx, _ := transaction.NewTransaction(adminPub, user1Pub, "nft-ds", transaction.TypeMintNFT)
	if err := mintTx.Sign(adminPriv); err != nil {
		t.Fatalf("failed to sign mint tx: %v", err)
	}
	block := NewBlock([]byte{}, []transaction.Transaction{*mintTx}, 1, "NODE")
	Mine(block)
	if _, err := bc.AddBlock(*block, adminPub); err != nil {
		t.Fatalf("mint block failed: %v", err)
	}

	// Create two conflicting transfers for the same NFT
	dsTx1, _ := transaction.NewTransaction(user1Pub, user2Pub, "nft-ds", transaction.TypeTransferNFT)
	if err := dsTx1.Sign(user1Priv); err != nil {
		t.Fatalf("failed to sign first transfer tx: %v", err)
	}
	dsTx2, _ := transaction.NewTransaction(user1Pub, "attacker", "nft-ds", transaction.TypeTransferNFT)
	if err := dsTx2.Sign(user1Priv); err != nil {
		t.Fatalf("failed to sign second transfer tx: %v", err)
	}

	dsBlock := NewBlock(block.Hash, []transaction.Transaction{*dsTx1, *dsTx2}, 1, "NODE")
	Mine(dsBlock)
	_, err := bc.AddBlock(*dsBlock, adminPub)
	if err == nil {
		t.Error("block with double-spend should be rejected or only one tx applied")
	}
}

// TestDoubleSpendAcrossBlocks ensures that a double-spend attempt across blocks is detected.
func TestDoubleSpendAcrossBlocks(t *testing.T) {
	bc := NewBlockchain()
	adminPriv, adminPub := transaction.TestKeyPair()
	user1Priv, user1Pub := transaction.TestKeyPair()
	_, user2Pub := transaction.TestKeyPair()

	// Mint NFT to user1
	mintTx, _ := transaction.NewTransaction(adminPub, user1Pub, "nft-ds2", transaction.TypeMintNFT)
	if err := mintTx.Sign(adminPriv); err != nil {
		t.Fatalf("failed to sign mint tx: %v", err)
	}
	block := NewBlock([]byte{}, []transaction.Transaction{*mintTx}, 1, "NODE")
	Mine(block)
	if _, err := bc.AddBlock(*block, adminPub); err != nil {
		t.Fatalf("mint block failed: %v", err)
	}

	// First transfer
	dsTx1, _ := transaction.NewTransaction(user1Pub, user2Pub, "nft-ds2", transaction.TypeTransferNFT)
	if err := dsTx1.Sign(user1Priv); err != nil {
		t.Fatalf("failed to sign first transfer tx: %v", err)
	}
	block1 := NewBlock(block.Hash, []transaction.Transaction{*dsTx1}, 1, "NODE")
	Mine(block1)
	if _, err := bc.AddBlock(*block1, adminPub); err != nil {
		t.Fatalf("first transfer block failed: %v", err)
	}

	// Second (conflicting) transfer
	dsTx2, _ := transaction.NewTransaction(user1Pub, "attacker", "nft-ds2", transaction.TypeTransferNFT)
	if err := dsTx2.Sign(user1Priv); err != nil {
		t.Fatalf("failed to sign second transfer tx: %v", err)
	}
	block2 := NewBlock(block1.Hash, []transaction.Transaction{*dsTx2}, 1, "NODE")
	Mine(block2)
	_, err := bc.AddBlock(*block2, adminPub)
	if err == nil {
		t.Error("block with double-spend across blocks should be rejected")
	}
}
