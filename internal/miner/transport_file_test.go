package miner

import "testing"

type mockTransport struct {
	published *Block
	handler   func(*Block) error
}

func (m *mockTransport) Publish(block *Block) error {
	m.published = block
	return nil
}

func (m *mockTransport) Subscribe(handler func(*Block) error) error {
	m.handler = handler
	return nil
}

func TestTransport_ImplementsBlockTransport(t *testing.T) {
	var _ BlockTransport = (*mockTransport)(nil)
}

func TestTransport_PublishAndSubscribeFlow(t *testing.T) {
	mt := &mockTransport{}
	b := &Block{Miner: "NODE-T"}

	if err := mt.Publish(b); err != nil {
		t.Fatalf("erro inesperado em Publish: %v", err)
	}
	if mt.published != b {
		t.Fatal("bloco publicado nao foi armazenado no mock")
	}

	called := false
	if err := mt.Subscribe(func(in *Block) error {
		called = true
		if in != b {
			t.Fatal("handler recebeu bloco incorreto")
		}
		return nil
	}); err != nil {
		t.Fatalf("erro inesperado em Subscribe: %v", err)
	}

	if mt.handler == nil {
		t.Fatal("handler de subscribe nao foi registrado")
	}
	if err := mt.handler(b); err != nil {
		t.Fatalf("erro inesperado no handler: %v", err)
	}
	if !called {
		t.Fatal("handler deveria ter sido chamado")
	}
}
