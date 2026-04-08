package miner

import (
	"bytes"
	"testing"

	"dominium/internal/transaction"
)

func createMockTx(id string) transaction.Transaction {
	tx, _ := transaction.NewTransaction("pubKeyA", "pubKeyB", id, transaction.TypeTransferNFT)
	return *tx
}

func TestCalculateMerkleRoot(t *testing.T) {
	tx1 := createMockTx("nft1")
	tx2 := createMockTx("nft2")
	tx3 := createMockTx("nft3")

	// Teste com 1 transação
	root1 := CalculateMerkleRoot([]transaction.Transaction{tx1})
	if len(root1) != 32 { // SHA-256 size
		t.Errorf("Esperado 32 bytes para root, recebido %d", len(root1))
	}

	// Teste com 2 transações
	root2 := CalculateMerkleRoot([]transaction.Transaction{tx1, tx2})
	if bytes.Equal(root1, root2) {
		t.Error("Root de 2 txs não deve ser igual a root de 1 tx")
	}

	// Teste com N (ímpar) transações - tx3 será duplicada internamente
	root3 := CalculateMerkleRoot([]transaction.Transaction{tx1, tx2, tx3})
	if len(root3) != 32 {
		t.Error("Falha ao calcular Merkle Root com N ímpar")
	}
}

func TestPoWMining(t *testing.T) {
	txs := []transaction.Transaction{createMockTx("genesis-nft")}
	block := NewBlock([]byte("00000000"), txs, 8,"NODE-1") // Dificuldade de 8 bits (1 byte) de zeros

	Mine(block)

	if !ValidateNonce(block.Header) {
		t.Errorf("Mineração finalizada, mas o Nonce não atendeu ao critério de %d bits", block.Nbits)
	}
}

func TestBlockchainGenesisAndPersistence(t *testing.T) {
	bc := NewBlockchain()
	
	// Criando bloco Genesis
	txsGenesis := []transaction.Transaction{createMockTx("genesis")}
	genesisBlock := NewBlock([]byte{}, txsGenesis, 4, "NODE-1") // Baixa dificuldade para testes
	Mine(genesisBlock)

	err := bc.AddBlock(*genesisBlock)
	if err != nil {
		t.Fatalf("Falha ao inserir bloco genesis: %v", err)
	}

	latestHash := bc.GetLatestHash()
	if !bytes.Equal(latestHash, genesisBlock.Hash) {
		t.Errorf("GetLatestHash() = %x, esperado %x", latestHash, genesisBlock.Hash)
	}

	// Bloco 2
	txsBlock2 := []transaction.Transaction{createMockTx("nft-transfer-1")}
	block2 := NewBlock(latestHash, txsBlock2, 4, "NODE-1")
	Mine(block2)
	
	err = bc.AddBlock(*block2)
	if err != nil {
		t.Fatalf("Falha ao inserir bloco 2: %v", err)
	}

	// ADAPTAÇÃO AQUI: Como mudamos para uma árvore, não usamos mais len(bc.Blocks)
	// Avaliamos o tamanho da cadeia canônica (oficial)
	canonical := bc.GetCanonicalChain()
	if len(canonical) != 2 {
		t.Errorf("Esperado 2 blocos na cadeia canônica, encontrado %d", len(canonical))
	}
}

// NOVO TESTE: Valida o comportamento de Forks e a Regra da Corrente Mais Longa
func TestBlockchainForksAndLongestChain(t *testing.T) {
	bc := NewBlockchain()

	// 1. Bloco Gênesis
	txsGen := []transaction.Transaction{createMockTx("genesis")}
	genBlock := NewBlock([]byte{}, txsGen, 1, "NODE-1") // Dificuldade 1 bit para rapidez
	Mine(genBlock)
	bc.AddBlock(*genBlock)

	// 2. Bloco A1 (Mina no topo do Gênesis)
	txsA1 := []transaction.Transaction{createMockTx("A1")}
	blockA1 := NewBlock(genBlock.Hash, txsA1, 1, "NODE-1")
	Mine(blockA1)
	bc.AddBlock(*blockA1)

	// 3. Bloco B1 (Outro minerador, também apoia-se no Gênesis - GERANDO UM FORK)
	txsB1 := []transaction.Transaction{createMockTx("B1")}
	blockB1 := NewBlock(genBlock.Hash, txsB1, 1, "NODE-1")
	Mine(blockB1)
	bc.AddBlock(*blockB1)

	// Ambos A1 e B1 têm altura 1 na árvore.
	// Como A1 foi inserido primeiro, ele continua a ser a ponta (bestChain tip).
	if !bytes.Equal(bc.GetLatestHash(), blockA1.Hash) {
		t.Errorf("Esperava tip A1, obteve outro hash")
	}

	// 4. Bloco B2 (Apoia-se no B1 - resolve o fork a favor do Ramo B)
	txsB2 := []transaction.Transaction{createMockTx("B2")}
	blockB2 := NewBlock(blockB1.Hash, txsB2, 1, "NODE-1")
	Mine(blockB2)
	bc.AddBlock(*blockB2)

	// Agora a cadeia B tem altura 2 e a cadeia A ficou para trás com altura 1.
	// A Regra da Corrente Mais Longa deve atualizar o tip para B2.
	if !bytes.Equal(bc.GetLatestHash(), blockB2.Hash) {
		t.Errorf("Regra falhou. O Tip devia ser B2 (A ramificação B agora é a mais longa)")
	}

	// Verifica a extração da cadeia canônica oficial
	canonical := bc.GetCanonicalChain()
	if len(canonical) != 3 {
		t.Errorf("Esperava cadeia canônica com 3 blocos (Gen -> B1 -> B2), obteve %d", len(canonical))
	}
	
	// Valida se o último bloco retornado pela corrente oficial é realmente o vencedor B2
	if !bytes.Equal(canonical[2].Hash, blockB2.Hash) {
		t.Errorf("O último bloco da cadeia canônica devia ser o B2")
	}
}