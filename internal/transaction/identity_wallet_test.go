package transaction

import "testing"

func TestIdentityRegistrySimpleTransferFlow(t *testing.T) {
	registry := NewIdentityRegistry()

	alice, err := registry.CreateIdentity("alice")
	if err != nil {
		t.Fatalf("falha ao criar identidade alice: %v", err)
	}
	bob, err := registry.CreateIdentity("bob")
	if err != nil {
		t.Fatalf("falha ao criar identidade bob: %v", err)
	}

	if got := registry.NameByPublicKey(alice.PublicKey); got != "alice" {
		t.Fatalf("esperava proprietario da chave publica alice, obtido: %s", got)
	}
	if got := registry.NameByPublicKey(bob.PublicKey); got != "bob" {
		t.Fatalf("esperava proprietario da chave publica bob, obtido: %s", got)
	}

	state := NewAccountState()
	state.Accounts[alice.PublicKey] = &Account{NFTs: map[string]bool{"nft-demo": true}}

	tx, err := NewTransaction(alice.PublicKey, bob.PublicKey, "nft-demo", TypeTransferNFT)
	if err != nil {
		t.Fatalf("falha ao criar transacao: %v", err)
	}

	if err := registry.SignTransactionAs("alice", tx); err != nil {
		t.Fatalf("falha ao assinar como alice: %v", err)
	}
	if err := tx.Validate(state); err != nil {
		t.Fatalf("esperava validacao da transacao assinada: %v", err)
	}

	txWrong, err := NewTransaction(bob.PublicKey, alice.PublicKey, "nft-demo", TypeTransferNFT)
	if err != nil {
		t.Fatalf("falha ao criar segunda transacao: %v", err)
	}

	if err := registry.SignTransactionAs("alice", txWrong); err == nil {
		t.Fatalf("esperava falha na assinatura quando identidade nao possui remetente da transacao")
	}
}
