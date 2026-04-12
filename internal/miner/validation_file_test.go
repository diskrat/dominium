package miner

import (
	"testing"

	"dominium/internal/transaction"
)

func TestValidation_ValidateNonceThreshold(t *testing.T) {
	header := Header{Nbits: 12, Hash: []byte{0x00, 0x0f}}
	if !ValidateNonce(header) {
		t.Fatal("esperava validar 12 bits iniciais em zero")
	}

	header.Nbits = 13
	if ValidateNonce(header) {
		t.Fatal("nao deveria validar 13 bits para esse hash")
	}
}

func TestValidation_RejectsBlockWithWrongParentHash(t *testing.T) {
	prev := NewBlock([]byte{}, []transaction.Transaction{createMockTx("gen")}, 1, "NODE")
	Mine(prev)

	blk := NewBlock([]byte("pai-errado"), []transaction.Transaction{createMockTx("tx")}, 1, "NODE")
	Mine(blk)

	if ValidateBlock(*blk, prev) {
		t.Fatal("bloco com hash de pai incorreto deveria ser invalido")
	}
}
