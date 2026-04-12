package miner

import (
	"bytes"
	"crypto/sha256"
	"testing"

	"dominium/internal/transaction"
)

func TestPow_MineFindsValidNonceAndHash(t *testing.T) {
	txs := []transaction.Transaction{createMockTx("pow-1")}
	blk := NewBlock([]byte("genesis"), txs, 8, "NODE-POW")

	Mine(blk)

	if !ValidateNonce(blk.Header) {
		t.Fatal("nonce minerado nao atende a dificuldade")
	}

	serialized, err := blk.Header.Serialize()
	if err != nil {
		t.Fatalf("erro ao serializar header: %v", err)
	}
	h := sha256.Sum256(serialized)
	if !bytes.Equal(blk.Hash, h[:]) {
		t.Fatal("hash final do bloco nao confere com header serializado")
	}
}
