package miner

import (
	"testing"

	"dominium/internal/transaction"
)

func TestBlockchain_RejectsOrphanBlock(t *testing.T) {
	bc := NewBlockchain()
	orphan := NewBlock([]byte("unknown-parent"), []transaction.Transaction{createMockTx("o1")}, 1, "NODE")
	Mine(orphan)

	if err := bc.AddBlock(*orphan); err == nil {
		t.Fatal("esperava erro ao adicionar bloco orfao")
	}
}

func TestBlockchain_RejectsDuplicateBlock(t *testing.T) {
	bc := NewBlockchain()
	genesis := NewBlock([]byte{}, []transaction.Transaction{createMockTx("g1")}, 1, "NODE")
	Mine(genesis)

	if err := bc.AddBlock(*genesis); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}
	if err := bc.AddBlock(*genesis); err == nil {
		t.Fatal("esperava erro para bloco duplicado")
	}
}
