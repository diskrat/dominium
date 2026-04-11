package miner

import (
	"bytes"
	"encoding/hex"
	"errors"
	"testing"

	"dominium/internal/transaction"
)

func TestBlockchain_RejectsOrphanBlock(t *testing.T) {
	bc := NewBlockchain()
	orphan := NewBlock([]byte("unknown-parent"), []transaction.Transaction{createMockTx("o1")}, 1, "NODE")
	Mine(orphan)

	if result, err := bc.AddBlock(*orphan); err == nil {
		t.Fatal("esperava erro ao adicionar bloco orfao")
	} else if !errors.Is(err, ErrOrphanBlock) {
		t.Fatalf("erro inesperado ao adicionar bloco orfao: %v", err)
	} else if !result.Orphan {
		t.Fatal("bloco orfao deveria ser marcado como orfao")
	}

	if bc.GetOrphanCount() != 1 {
		t.Fatal("esperava um bloco orfao na fila")
	}
}

func TestBlockchain_RejectsDuplicateBlock(t *testing.T) {
	bc := NewBlockchain()
	genesis := NewBlock([]byte{}, []transaction.Transaction{createMockTx("g1")}, 1, "NODE")
	Mine(genesis)

	if _, err := bc.AddBlock(*genesis); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}
	if _, err := bc.AddBlock(*genesis); err == nil {
		t.Fatal("esperava erro para bloco duplicado")
	}
}

func TestBlockchain_AttachesOrphanWhenParentArrives(t *testing.T) {
	bc := NewBlockchain()

	genesis := NewBlock([]byte{}, []transaction.Transaction{createMockTx("g1")}, 1, "NODE")
	Mine(genesis)

	block1 := NewBlock(genesis.Hash, []transaction.Transaction{createMockTx("b1")}, 1, "NODE")
	Mine(block1)
	block2 := NewBlock(block1.Hash, []transaction.Transaction{createMockTx("b2")}, 1, "NODE")
	Mine(block2)

	if _, err := bc.AddBlock(*block2); err == nil {
		t.Fatal("esperava erro ao adicionar bloco orfao")
	}
	if bc.GetOrphanCount() != 1 {
		t.Fatal("esperava bloco orfao na fila")
	}

	if _, err := bc.AddBlock(*genesis); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}
	if _, err := bc.AddBlock(*block1); err != nil {
		t.Fatalf("falha ao adicionar bloco pai: %v", err)
	}

	if bc.GetOrphanCount() != 0 {
		t.Fatal("fila de orfaos deveria estar vazia")
	}
	if _, ok := bc.GetBlockByHash(block2.Hash); !ok {
		t.Fatal("bloco orfao deveria ser anexado quando pai chega")
	}
}

func TestBlockchain_ForkConsolidationAfterTwoBlocks(t *testing.T) {
	bc := NewBlockchain()

	genesis := NewBlock([]byte{}, []transaction.Transaction{createMockTx("g1")}, 1, "NODE")
	Mine(genesis)
	if _, err := bc.AddBlock(*genesis); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}

	blockA1 := NewBlock(genesis.Hash, []transaction.Transaction{createMockTx("a1")}, 1, "NODE-A")
	Mine(blockA1)
	if _, err := bc.AddBlock(*blockA1); err != nil {
		t.Fatalf("falha ao adicionar bloco A1: %v", err)
	}

	blockB1 := NewBlock(genesis.Hash, []transaction.Transaction{createMockTx("b1")}, 1, "NODE-B")
	Mine(blockB1)
	if _, err := bc.AddBlock(*blockB1); err != nil {
		t.Fatalf("falha ao adicionar bloco B1: %v", err)
	}

	if got, want := bc.GetLatestHash(), blockA1.Hash; !bytes.Equal(got, want) {
		t.Fatal("fork com mesma altura nao deve mudar a ponta")
	}

	blockB2 := NewBlock(blockB1.Hash, []transaction.Transaction{createMockTx("b2")}, 1, "NODE-B")
	Mine(blockB2)
	if _, err := bc.AddBlock(*blockB2); err != nil {
		t.Fatalf("falha ao adicionar bloco B2: %v", err)
	}

	if got, want := bc.GetLatestHash(), blockA1.Hash; !bytes.Equal(got, want) {
		t.Fatal("fork com diferenca de 1 bloco nao deve mudar a ponta")
	}

	blockB3 := NewBlock(blockB2.Hash, []transaction.Transaction{createMockTx("b3")}, 1, "NODE-B")
	Mine(blockB3)
	if _, err := bc.AddBlock(*blockB3); err != nil {
		t.Fatalf("falha ao adicionar bloco B3: %v", err)
	}

	if got, want := bc.GetLatestHash(), blockB3.Hash; !bytes.Equal(got, want) {
		t.Fatal("fork com diferenca de 2 blocos deve consolidar")
	}
}

func TestBlockchain_PrunesStaleForkWhenBehindByTwo(t *testing.T) {
	bc := NewBlockchain()

	genesis := NewBlock([]byte{}, []transaction.Transaction{createMockTx("g1")}, 1, "NODE")
	Mine(genesis)
	if _, err := bc.AddBlock(*genesis); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}

	blockA1 := NewBlock(genesis.Hash, []transaction.Transaction{createMockTx("a1")}, 1, "NODE-A")
	Mine(blockA1)
	if _, err := bc.AddBlock(*blockA1); err != nil {
		t.Fatalf("falha ao adicionar bloco A1: %v", err)
	}

	blockB1 := NewBlock(genesis.Hash, []transaction.Transaction{createMockTx("b1")}, 1, "NODE-B")
	Mine(blockB1)
	if _, err := bc.AddBlock(*blockB1); err != nil {
		t.Fatalf("falha ao adicionar bloco B1: %v", err)
	}

	blockA2 := NewBlock(blockA1.Hash, []transaction.Transaction{createMockTx("a2")}, 1, "NODE-A")
	Mine(blockA2)
	if _, err := bc.AddBlock(*blockA2); err != nil {
		t.Fatalf("falha ao adicionar bloco A2: %v", err)
	}

	blockA3 := NewBlock(blockA2.Hash, []transaction.Transaction{createMockTx("a3")}, 1, "NODE-A")
	Mine(blockA3)
	if _, err := bc.AddBlock(*blockA3); err != nil {
		t.Fatalf("falha ao adicionar bloco A3: %v", err)
	}

	if _, ok := bc.GetBlockByHash(blockB1.Hash); ok {
		t.Fatal("fork atrasado deveria ter sido descartado")
	}
	if _, ok := bc.GetDiscardedSnapshot()[hex.EncodeToString(blockB1.Hash)]; !ok {
		t.Fatal("fork descartado deveria aparecer nos descartados")
	}
}
