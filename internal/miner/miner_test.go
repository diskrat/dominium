package miner

import (
	"dominium/internal/transaction"
)

func createMockTx(id string) transaction.Transaction {
	tx, _ := transaction.NewTransaction("pubKeyA", "pubKeyB", id, transaction.TypeTransferNFT)
	return *tx
}
