package miner

import (
	"bytes"
	"testing"

	"dominium/internal/transaction"
)

func TestMerkle_EmptyTransactionsReturnsEmptyHash(t *testing.T) {
	root := CalculateMerkleRoot(nil)
	if len(root) != 0 {
		t.Fatalf("esperava raiz vazia, obteve %d bytes", len(root))
	}
}

func TestMerkle_DeterministicForSameTransactions(t *testing.T) {
	txs := []transaction.Transaction{createMockTx("m1"), createMockTx("m2"), createMockTx("m3")}

	r1 := CalculateMerkleRoot(txs)
	r2 := CalculateMerkleRoot(txs)

	if !bytes.Equal(r1, r2) {
		t.Fatal("merkle root deve ser deterministica para a mesma entrada")
	}
}
