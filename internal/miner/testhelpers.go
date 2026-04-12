package miner

import (
	"dominium/internal/transaction"
)

// createMockTx creates a mock transaction for testing, signed with a generated test key pair.
func createMockTx(id string) transaction.Transaction {
	priv, pub := transaction.TestKeyPair()
	tx, _ := transaction.NewTransaction(pub, "recipient-"+id, id, transaction.TypeTransferNFT)
	_ = tx.Sign(priv)
	return *tx
}
