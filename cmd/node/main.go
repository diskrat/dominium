package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"dominium/internal/transaction"
)

// Node encapsula o que ja existe (estado/mempool) e abstrai rede, mineracao e API.
type Node struct {
	state   *transaction.AccountState
	mempool *transaction.Mempool

	// TODO: adicionar rede P2P, mineracao e API.
}

func NewNode() *Node {
	state := transaction.NewAccountState()
	return &Node{
		state:   state,
		mempool: transaction.NewMempool(state),
	}
}

func (n *Node) Start() error {
	// TODO: iniciar rede P2P, mineracao e API.
	return nil
}

func (n *Node) Stop(ctx context.Context) error {
	// TODO: encerrar servicos de rede/mineracao/API.
	_ = ctx
	return nil
}

func main() {
	var (
		apiAddr = flag.String("api", ":8080", "endereco da API")
		p2pAddr = flag.String("p2p", ":9000", "endereco P2P")
		dataDir = flag.String("data", "./data", "diretorio de dados")
		mine    = flag.Bool("mine", false, "habilita mineracao")
	)
	flag.Parse()

	_ = apiAddr
	_ = p2pAddr
	_ = dataDir
	_ = mine

	n := NewNode()
	if err := n.Start(); err != nil {
		log.Fatalf("start: %v", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	<-ctx.Done()
	stop()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := n.Stop(shutdownCtx); err != nil {
		log.Printf("stop: %v", err)
	}
}
