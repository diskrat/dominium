package miner

import (
	"bytes"
	"testing"

	"dominium/internal/transaction"
)

func TestBlock_NewBlockInitializesHeaderAndBody(t *testing.T) {
	prevHash := []byte("prev-hash")
	txs := []transaction.Transaction{createMockTx("nft-1"), createMockTx("nft-2")}

	blk := NewBlock(prevHash, txs, 8, "NODE-X")

	if !bytes.Equal(blk.HashOfPrevious, prevHash) {
		t.Fatal("hash anterior nao foi copiado corretamente")
	}
	if blk.TransactionCounter != uint16(len(txs)) {
		t.Fatalf("transaction counter invalido: %d", blk.TransactionCounter)
	}
	if !bytes.Equal(blk.MerkleRootHash, CalculateMerkleRoot(txs)) {
		t.Fatal("merkle root inconsistente")
	}
	if blk.Miner != "NODE-X" {
		t.Fatalf("miner esperado NODE-X, obtido %s", blk.Miner)
	}
}

func TestBlock_HeaderSerializeIsDeterministic(t *testing.T) {
	h := Header{
		HashOfPrevious:     []byte("a"),
		MerkleRootHash:     []byte("b"),
		Timestamp:          123,
		Nbits:              16,
		Nonce:              42,
		TransactionCounter: 2,
	}

	b1, err := h.Serialize()
	if err != nil {
		t.Fatalf("erro no serialize: %v", err)
	}
	b2, err := h.Serialize()
	if err != nil {
		t.Fatalf("erro no serialize: %v", err)
	}

	if !bytes.Equal(b1, b2) {
		t.Fatal("serializacao deveria ser deterministica")
	}
}
