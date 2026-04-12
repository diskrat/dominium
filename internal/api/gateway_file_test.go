package api

import (
	"context"
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"dominium/internal/miner"
	"dominium/internal/transaction"
)

func newAdminIdentity(t *testing.T) (*transaction.WalletIdentity, *ecdsa.PrivateKey) {
	t.Helper()

	priv, pub, err := transaction.GenerateKeyPair()
	if err != nil {
		t.Fatalf("erro ao gerar chave admin: %v", err)
	}
	privHex, err := transaction.EncodePrivateKey(priv)
	if err != nil {
		t.Fatalf("erro ao codificar chave admin: %v", err)
	}
	pubHex := transaction.EncodePublicKey(pub)

	identity := &transaction.WalletIdentity{
		Name:       "admin",
		PublicKey:  pubHex,
		PrivateKey: privHex,
	}
	return identity, priv
}

func createSignedTx(t *testing.T, publKey, recipient, nftID string, txType byte, priv *ecdsa.PrivateKey) *transaction.Transaction {
	t.Helper()

	tx, err := transaction.NewTransaction(publKey, recipient, nftID, txType)
	if err != nil {
		t.Fatalf("erro ao criar transacao: %v", err)
	}
	if err := tx.Sign(priv); err != nil {
		t.Fatalf("erro ao assinar transacao: %v", err)
	}
	return tx
}

func TestGateway_ProcessBlockUpdatesStateAndMetadata(t *testing.T) {
	g := NewGateway(context.Background(), "gateway-test", 0, nil)
	adminIdentity, adminPriv := newAdminIdentity(t)
	if err := g.SetAdminIdentity(adminIdentity); err != nil {
		t.Fatalf("erro ao configurar admin: %v", err)
	}

	recipient := "recipient-1"
	tx := createSignedTx(t, adminIdentity.PublicKey, recipient, "nft-1", transaction.TypeMintNFT, adminPriv)
	block := miner.NewBlock([]byte{}, []transaction.Transaction{*tx}, 1, "NODE")
	miner.Mine(block)

	if err := g.processBlock(block); err != nil {
		t.Fatalf("erro ao processar bloco: %v", err)
	}

	g.stateMu.RLock()
	accounts := g.state.GetAllAccounts()
	g.stateMu.RUnlock()

	acc, ok := accounts[recipient]
	if !ok || acc == nil || !acc.NFTs["nft-1"] {
		t.Fatal("estado nao foi atualizado com NFT do bloco")
	}

	req := httptest.NewRequest(http.MethodGet, "/network/status", nil)
	rec := httptest.NewRecorder()
	g.handleGetNetworkStatus(rec, req)

	var payload struct {
		AllBlocks []BlockMetadata `json:"all_blocks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}

	hashHex := hex.EncodeToString(block.Hash)
	found := false
	for _, meta := range payload.AllBlocks {
		if meta.Hash == hashHex && meta.MinerID == "NODE" {
			found = true
			break
		}
	}
	if !found {
		t.Fatal("bloco nao apareceu no status com miner correto")
	}
}

func TestGateway_HandleGetNetworkStatusShowsForkHeights(t *testing.T) {
	g := NewGateway(context.Background(), "gateway-test", 0, nil)
	adminIdentity, adminPriv := newAdminIdentity(t)
	if err := g.SetAdminIdentity(adminIdentity); err != nil {
		t.Fatalf("erro ao configurar admin: %v", err)
	}

	genesisTx := createSignedTx(t, adminIdentity.PublicKey, "r1", "nft-gen", transaction.TypeMintNFT, adminPriv)
	genesis := miner.NewBlock([]byte{}, []transaction.Transaction{*genesisTx}, 1, "NODE")
	miner.Mine(genesis)
	if err := g.processBlock(genesis); err != nil {
		t.Fatalf("erro ao processar genesis: %v", err)
	}

	blockATx := createSignedTx(t, adminIdentity.PublicKey, "r2", "nft-a", transaction.TypeMintNFT, adminPriv)
	blockA := miner.NewBlock(genesis.Hash, []transaction.Transaction{*blockATx}, 1, "NODE-A")
	miner.Mine(blockA)
	blockBTx := createSignedTx(t, adminIdentity.PublicKey, "r3", "nft-b", transaction.TypeMintNFT, adminPriv)
	blockB := miner.NewBlock(genesis.Hash, []transaction.Transaction{*blockBTx}, 1, "NODE-B")
	miner.Mine(blockB)

	if err := g.processBlock(blockA); err != nil {
		t.Fatalf("erro ao processar bloco A: %v", err)
	}
	if err := g.processBlock(blockB); err != nil {
		t.Fatalf("erro ao processar bloco B: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/network/status", nil)
	rec := httptest.NewRecorder()
	g.handleGetNetworkStatus(rec, req)

	var payload struct {
		AllBlocks []BlockMetadata `json:"all_blocks"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("erro ao decodificar resposta: %v", err)
	}

	heights := make(map[string]uint64)
	for _, b := range payload.AllBlocks {
		heights[b.Hash] = b.Height
	}

	if heights[hex.EncodeToString(blockA.Hash)] != 1 || heights[hex.EncodeToString(blockB.Hash)] != 1 {
		t.Fatal("forks deveriam aparecer na altura real 1")
	}
}

func TestGateway_RebuildStateRemovesConfirmedMempool(t *testing.T) {
	g := NewGateway(context.Background(), "gateway-test", 0, nil)
	adminIdentity, adminPriv := newAdminIdentity(t)
	if err := g.SetAdminIdentity(adminIdentity); err != nil {
		t.Fatalf("erro ao configurar admin: %v", err)
	}

	recipient := "recipient-1"
	tx := createSignedTx(t, adminIdentity.PublicKey, recipient, "nft-2", transaction.TypeMintNFT, adminPriv)

	g.stateMu.Lock()
	g.mempool[tx.ID] = tx
	g.stateMu.Unlock()

	block := miner.NewBlock([]byte{}, []transaction.Transaction{*tx}, 1, "NODE")
	miner.Mine(block)
	if err := g.processBlock(block); err != nil {
		t.Fatalf("erro ao processar bloco: %v", err)
	}

	g.stateMu.RLock()
	_, exists := g.mempool[tx.ID]
	g.stateMu.RUnlock()
	if exists {
		t.Fatal("transacao confirmada deveria sair da mempool")
	}
}
