package miner

import (
	"bytes"
	"errors"
	"fmt"
	"testing"

	"dominium/internal/transaction"
)

func TestBlockchain_RejectsOrphanBlock(t *testing.T) {
	bc := NewBlockchain()
	orphan := NewBlock([]byte("unknown-parent"), []transaction.Transaction{}, 1, "NODE")
	Mine(orphan)

	if result, err := bc.AddBlock(*orphan, ""); err == nil {
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
	genesis := NewBlock([]byte{}, []transaction.Transaction{}, 1, "NODE")
	Mine(genesis)

	if _, err := bc.AddBlock(*genesis, ""); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}
	if _, err := bc.AddBlock(*genesis, ""); err == nil {
		t.Fatal("esperava erro para bloco duplicado")
	}
}

func TestBlockchain_AttachesOrphanWhenParentArrives(t *testing.T) {
	bc := NewBlockchain()

	genesis := NewBlock([]byte{}, []transaction.Transaction{}, 1, "NODE")
	Mine(genesis)

	block1 := NewBlock(genesis.Hash, []transaction.Transaction{}, 1, "NODE")
	Mine(block1)
	block2 := NewBlock(block1.Hash, []transaction.Transaction{}, 1, "NODE")
	Mine(block2)

	if _, err := bc.AddBlock(*block2, ""); err == nil {
		t.Fatal("esperava erro ao adicionar bloco orfao")
	}
	if bc.GetOrphanCount() != 1 {
		t.Fatal("esperava bloco orfao na fila")
	}

	if _, err := bc.AddBlock(*genesis, ""); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}
	if _, err := bc.AddBlock(*block1, ""); err != nil {
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

	genesis := NewBlock([]byte{}, []transaction.Transaction{}, 1, "NODE")
	Mine(genesis)
	if _, err := bc.AddBlock(*genesis, ""); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}

	blockA1 := NewBlock(genesis.Hash, []transaction.Transaction{}, 1, "NODE-A")
	Mine(blockA1)
	if _, err := bc.AddBlock(*blockA1, ""); err != nil {
		t.Fatalf("falha ao adicionar bloco A1: %v", err)
	}

	blockB1 := NewBlock(genesis.Hash, []transaction.Transaction{}, 1, "NODE-B")
	Mine(blockB1)
	if _, err := bc.AddBlock(*blockB1, ""); err != nil {
		t.Fatalf("falha ao adicionar bloco B1: %v", err)
	}

	if got, want := bc.GetLatestHash(), blockA1.Hash; !bytes.Equal(got, want) {
		t.Fatal("fork com mesma altura nao deve mudar a ponta")
	}

	blockB2 := NewBlock(blockB1.Hash, []transaction.Transaction{}, 1, "NODE-B")
	Mine(blockB2)
	if _, err := bc.AddBlock(*blockB2, ""); err != nil {
		t.Fatalf("falha ao adicionar bloco B2: %v", err)
	}

	if got, want := bc.GetLatestHash(), blockA1.Hash; !bytes.Equal(got, want) {
		t.Fatal("fork com diferenca de 1 bloco nao deve mudar a ponta")
	}

	blockB3 := NewBlock(blockB2.Hash, []transaction.Transaction{}, 1, "NODE-B")
	Mine(blockB3)
	if _, err := bc.AddBlock(*blockB3, ""); err != nil {
		t.Fatalf("falha ao adicionar bloco B3: %v", err)
	}

	if got, want := bc.GetLatestHash(), blockB3.Hash; !bytes.Equal(got, want) {
		t.Fatal("fork com diferenca de 2 blocos deve consolidar")
	}
}

func TestBlockchain_LongForkDelayedAdvantage(t *testing.T) {
	bc := NewBlockchain()

	genesis := NewBlock([]byte{}, []transaction.Transaction{}, 1, "NODE")
	Mine(genesis)
	if _, err := bc.AddBlock(*genesis, ""); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}

	mainBlocks := make([]*Block, 3)
	parent := genesis
	for i := 0; i < 3; i++ {
		mainBlocks[i] = NewBlock(parent.Hash, []transaction.Transaction{}, 1, fmt.Sprintf("MAIN-%d", i+1))
		Mine(mainBlocks[i])
		if _, err := bc.AddBlock(*mainBlocks[i], ""); err != nil {
			t.Fatalf("falha ao adicionar main block %d: %v", i+1, err)
		}
		parent = mainBlocks[i]
	}

	forkBase := mainBlocks[2]
	chainA := make([]*Block, 5)
	chainB := make([]*Block, 7)
	parentA := forkBase
	parentB := forkBase

	for i := 0; i < len(chainB); i++ {
		if i < len(chainA) {
			chainA[i] = NewBlock(parentA.Hash, []transaction.Transaction{}, 1, fmt.Sprintf("A-%d", i+4))
			Mine(chainA[i])
			if _, err := bc.AddBlock(*chainA[i], ""); err != nil {
				t.Fatalf("falha ao adicionar A-%d: %v", i+4, err)
			}
			parentA = chainA[i]
		}

		chainB[i] = NewBlock(parentB.Hash, []transaction.Transaction{}, 1, fmt.Sprintf("B-%d", i+4))
		Mine(chainB[i])
		if _, err := bc.AddBlock(*chainB[i], ""); err != nil {
			t.Fatalf("falha ao adicionar B-%d: %v", i+4, err)
		}
		parentB = chainB[i]

		if i == 4 {
			if got, want := bc.GetLatestHash(), chainA[4].Hash; !bytes.Equal(got, want) {
				t.Fatalf("chain B com mesma altura nao deve mudar a ponta no B-%d", i+4)
			}
		}
	}

	if got, want := bc.GetLatestHash(), chainB[6].Hash; !bytes.Equal(got, want) {
		t.Fatal("chain B com vantagem de 2 blocos deveria se tornar a nova ponta")
	}

	for i := 0; i < len(chainA); i++ {
		if _, ok := bc.GetBlockByHash(chainA[i].Hash); ok {
			t.Fatalf("bloco A-%d deveria ter sido apagado apos reorganizacao", i+4)
		}
	}
}

func TestBlockchainRejectsBlockWithInvalidTransactionSignature(t *testing.T) {
	bc := NewBlockchain()
	adminPriv, adminPub := transaction.TestKeyPair()

	genesis := NewBlock([]byte{}, []transaction.Transaction{}, 1, "NODE")
	Mine(genesis)
	if _, err := bc.AddBlock(*genesis, ""); err != nil {
		t.Fatalf("falha ao adicionar genesis: %v", err)
	}

	tx, err := transaction.NewTransaction(adminPub, "recipient", "nft-invalid-sig", transaction.TypeMintNFT)
	if err != nil {
		t.Fatalf("failed to create transaction: %v", err)
	}
	if err := tx.Sign(adminPriv); err != nil {
		t.Fatalf("failed to sign transaction: %v", err)
	}
	tx.Sig[0] ^= 0xFF

	block := NewBlock(genesis.Hash, []transaction.Transaction{*tx}, 1, "NODE")
	Mine(block)
	if _, err := bc.AddBlock(*block, adminPub); err == nil {
		t.Fatal("expected block with invalid transaction signature to be rejected")
	}
}
