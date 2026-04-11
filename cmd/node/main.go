package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv" // Necessário para converter a string do .env para número
	"strings"
	"syscall"
	"time"

	"dominium/internal/node"

	"github.com/segmentio/kafka-go"
)

func main() {
	id := flag.String("id", "", "Identificador unico deste no Kafka")
	p2p := flag.String("p2p", "", "Endereco do broker Kafka")
	mine := flag.Bool("mine", false, "Ativa o motor de mineracao PoW")
	// Alteramos o valor padrão da flag para 0 para detectar se o usuário passou algo via terminal
	difficultyFlag := flag.Int("difficulty", 0, "Dificuldade de mineracao (Nbits)")
	flag.Parse()

	if *id == "" {
		log.Fatal("flag -id e obrigatoria")
	}

	// 1. Lógica de Prioridade para o Kafka
	brokerStr := *p2p
	if brokerStr == "" {
		brokerStr = os.Getenv("P2P_BROKERS")
	}
	if brokerStr == "" {
		brokerStr = "localhost:9092"
	}

	// 2. Lógica de Prioridade para Dificuldade (Busca no .env primeiro)
	difficulty := 4 // Valor padrão de segurança caso nada seja encontrado

	// Tenta ler "DIFFICULTY" (nome que você usou no .env)
	if envDiff := os.Getenv("DIFFICULTY"); envDiff != "" {
		if d, err := strconv.Atoi(envDiff); err == nil {
			difficulty = d
		}
	} else if *difficultyFlag > 0 {
		// Se não houver no ENV, mas o usuário passou via flag "-difficulty X"
		difficulty = *difficultyFlag
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	brokers := strings.Split(brokerStr, ",")

	// 3. Espera Kafka ficar online antes de iniciar o node
	for {
		if kafkaOnline(brokers) {
			break
		}
		log.Printf("[startup] Aguardando Kafka (%s) ficar online...", brokerStr)
		time.Sleep(2 * time.Second)
	}

	nodeServer := node.NewNodeServer(ctx, *id, brokers, *mine, int32(difficulty))
	log.Printf("[%s] Iniciando com Dificuldade: %d", *id, difficulty)
	if err := nodeServer.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "erro ao iniciar node: %v\n", err)
		os.Exit(1)
	}
}

// kafkaOnline faz um healthcheck simples tentando conectar no broker
func kafkaOnline(brokers []string) bool {
	for _, addr := range brokers {
		conn, err := kafka.Dial("tcp", addr)
		if err == nil {
			conn.Close()
			return true
		}
	}
	return false
}
