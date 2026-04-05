package tests

import (
	"testing"

	"front-p2p/mocks"
)

func TestMockBlockchain_Genesis(t *testing.T) {
	bc := mocks.NewMockBlockchain()

	blocks := bc.GetBlocks()
	if len(blocks) != 1 {
		t.Errorf("Expected 1 genesis block, got %d", len(blocks))
	}

	genesis := blocks[0]
	if genesis.Index != 0 {
		t.Errorf("Expected genesis index 0, got %d", genesis.Index)
	}

	if genesis.PrevHash != "0" {
		t.Errorf("Expected genesis prev hash '0', got '%s'", genesis.PrevHash)
	}
}

func TestMockBlockchain_AddBlock(t *testing.T) {
	bc := mocks.NewMockBlockchain()
	genesis := bc.GetLastBlock()

	block1 := &mocks.MockBlock{
		Index:    1,
		Hash:     "hash_1",
		PrevHash: genesis.Hash,
		Data:     "Block 1",
		Nonce:    100,
	}

	err := bc.AddBlock(block1)
	if err != nil {
		t.Errorf("AddBlock failed: %v", err)
	}

	if bc.GetHeight() != 2 {
		t.Errorf("Expected height 2, got %d", bc.GetHeight())
	}
}

func TestMockBlockchain_AddBlockInvalidPrevHash(t *testing.T) {
	bc := mocks.NewMockBlockchain()

	block1 := &mocks.MockBlock{
		Index:    1,
		Hash:     "hash_1",
		PrevHash: "invalid_hash",
		Data:     "Block 1",
		Nonce:    100,
	}

	err := bc.AddBlock(block1)
	if err == nil {
		t.Error("Expected error for invalid previous hash")
	}
}

func TestMockBlockchain_AddBlockInvalidIndex(t *testing.T) {
	bc := mocks.NewMockBlockchain()
	genesis := bc.GetLastBlock()

	block1 := &mocks.MockBlock{
		Index:    5, // Wrong index
		Hash:     "hash_1",
		PrevHash: genesis.Hash,
		Data:     "Block 1",
		Nonce:    100,
	}

	err := bc.AddBlock(block1)
	if err == nil {
		t.Error("Expected error for invalid block index")
	}
}

func TestMockBlockchain_GetBlock(t *testing.T) {
	bc := mocks.NewMockBlockchain()
	genesis := bc.GetLastBlock()

	block1 := &mocks.MockBlock{
		Index:    1,
		Hash:     "hash_1",
		PrevHash: genesis.Hash,
		Data:     "Block 1",
		Nonce:    100,
	}
	bc.AddBlock(block1)

	retrieved := bc.GetBlock(1)
	if retrieved == nil {
		t.Error("Expected to retrieve block 1")
	}

	if retrieved.Hash != "hash_1" {
		t.Errorf("Expected hash 'hash_1', got '%s'", retrieved.Hash)
	}
}

func TestMockBlockchain_GetBlockInvalidIndex(t *testing.T) {
	bc := mocks.NewMockBlockchain()

	block := bc.GetBlock(100)
	if block != nil {
		t.Error("Expected nil for invalid index")
	}
}

func TestMockBlockchain_AddTransaction(t *testing.T) {
	bc := mocks.NewMockBlockchain()

	tx := &mocks.MockTransaction{
		ID:        "tx_1",
		Type:      "REGISTER",
		Sender:    "sender_1",
		Recipient: "recipient_1",
		Data:      map[string]interface{}{"property": "123"},
	}

	err := bc.AddTransaction(tx)
	if err != nil {
		t.Errorf("AddTransaction failed: %v", err)
	}

	mempool := bc.GetMempool()
	if len(mempool) != 1 {
		t.Errorf("Expected 1 transaction in mempool, got %d", len(mempool))
	}
}

func TestMockBlockchain_ClearMempool(t *testing.T) {
	bc := mocks.NewMockBlockchain()

	tx1 := &mocks.MockTransaction{ID: "tx_1", Type: "REGISTER"}
	tx2 := &mocks.MockTransaction{ID: "tx_2", Type: "TRANSFER"}

	bc.AddTransaction(tx1)
	bc.AddTransaction(tx2)

	mempool := bc.GetMempool()
	if len(mempool) != 2 {
		t.Errorf("Expected 2 transactions before clear, got %d", len(mempool))
	}

	bc.ClearMempool()

	mempool = bc.GetMempool()
	if len(mempool) != 0 {
		t.Errorf("Expected 0 transactions after clear, got %d", len(mempool))
	}
}

func TestMockBlockchain_GetHeight(t *testing.T) {
	bc := mocks.NewMockBlockchain()

	if bc.GetHeight() != 1 {
		t.Errorf("Expected initial height 1, got %d", bc.GetHeight())
	}

	for i := 1; i <= 5; i++ {
		lastBlock := bc.GetLastBlock()
		block := &mocks.MockBlock{
			Index:    i,
			Hash:     "hash_" + string(rune(i)),
			PrevHash: lastBlock.Hash,
			Data:     "Block",
			Nonce:    int64(i * 100),
		}
		bc.AddBlock(block)
	}

	if bc.GetHeight() != 6 {
		t.Errorf("Expected height 6, got %d", bc.GetHeight())
	}
}

// Benchmark tests
func BenchmarkMockBlockchain_AddBlock(b *testing.B) {
	bc := mocks.NewMockBlockchain()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		lastBlock := bc.GetLastBlock()
		block := &mocks.MockBlock{
			Index:    lastBlock.Index + 1,
			Hash:     "hash",
			PrevHash: lastBlock.Hash,
			Data:     "data",
			Nonce:    int64(i),
		}
		bc.AddBlock(block)
	}
}

func BenchmarkMockBlockchain_AddTransaction(b *testing.B) {
	bc := mocks.NewMockBlockchain()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		tx := &mocks.MockTransaction{
			ID:   "tx",
			Type: "REGISTER",
		}
		bc.AddTransaction(tx)
	}
}
