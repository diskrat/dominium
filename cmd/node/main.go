package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"dominium/internal/node"
)

func main() {
	id := flag.String("id", "", "Identificador unico deste no Kafka")
	p2p := flag.String("p2p", "localhost:9092", "Endereco do broker Kafka")
	mine := flag.Bool("mine", false, "Ativa o motor de mineracao PoW")
	difficulty := flag.Int("difficulty", 4, "Dificuldade de mineracao (Nbits)")
	dataDir := flag.String("data", "./data", "Diretorio de persistencia local (futuro)")
	flag.Parse()

	if *id == "" {
		log.Fatal("flag -id e obrigatoria")
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	brokers := strings.Split(*p2p, ",")

	nodeServer := node.NewNodeServer(ctx, *id, brokers, *mine, int32(*difficulty), *dataDir)
	if err := nodeServer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "erro ao iniciar node: %v\n", err)
		os.Exit(1)
	}
}
