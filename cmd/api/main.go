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
	"time"

	"dominium/internal/api"
	"dominium/internal/transaction"

	"github.com/segmentio/kafka-go"
)

func main() {
	id := flag.String("id", "api-gateway", "Identificador do API Gateway")
	port := flag.Int("port", 8085, "Porta HTTP do servidor")
	p2p := flag.String("p2p", "", "Endereco do broker Kafka") // Padrão vazio
	adminPrivKey := flag.String("admin-key", "", "Chave privada do admin em hex (obrigatoria)")
	adminPubKey := flag.String("admin-pub", "", "Chave publica do admin em hex (obrigatoria)")
	flag.Parse()

	if *adminPrivKey == "" || *adminPubKey == "" {
		log.Fatal("flags -admin-key e -admin-pub sao obrigatorias")
	}

	// Lógica de Prioridade para o Kafka
	brokerStr := *p2p
	if brokerStr == "" {
		brokerStr = os.Getenv("P2P_BROKERS")
	}
	if brokerStr == "" {
		brokerStr = "localhost:9092"
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	adminIdentity := &transaction.WalletIdentity{
		Name:       "admin",
		PublicKey:  *adminPubKey,
		PrivateKey: *adminPrivKey,
	}

	brokers := strings.Split(brokerStr, ",")
	for {
		if kafkaOnline(brokers) {
			break
		}
		log.Printf("[startup] Aguardando Kafka (%s) ficar online...", brokerStr)
		time.Sleep(2 * time.Second)
	}

	gateway := api.NewGateway(ctx, *id, *port, brokers)
	if err := gateway.SetAdminIdentity(adminIdentity); err != nil {
		fmt.Fprintf(os.Stderr, "erro ao configurar admin: %v\n", err)
		os.Exit(1)
	}

	log.Printf("API Gateway iniciando (id=%s port=%d brokers=%v)", *id, *port, brokers)
	if err := gateway.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "erro ao iniciar gateway: %v\n", err)
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
