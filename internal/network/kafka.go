package network

import (
	"context"
	"encoding/json"
	"log"

	"dominium/internal/miner"
	"dominium/internal/transaction"

	"github.com/segmentio/kafka-go"
)

const (
	TopicTransactions = "dominium-transactions"
	TopicBlocks       = "dominium-blocks"
)

// P2P gerencia as conexões Kafka para simular a rede.
type P2P struct {
	brokers []string
	nodeID  string
}

func NewP2P(brokers []string, nodeID string) *P2P {
	return &P2P{
		brokers: brokers,
		nodeID:  nodeID,
	}
}

// PublishTransaction envia uma transação para a rede.
func (n *P2P) PublishTransaction(ctx context.Context, tx *transaction.Transaction) error {
	return n.publish(ctx, TopicTransactions, tx)
}

// PublishBlock envia um bloco recém-minerado para a rede.
func (n *P2P) PublishBlock(ctx context.Context, block *miner.Block) error {
	return n.publish(ctx, TopicBlocks, block)
}

func (n *P2P) publish(ctx context.Context, topic string, message interface{}) error {
	w := &kafka.Writer{
		Addr:     kafka.TCP(n.brokers...),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	defer w.Close()

	msgBytes, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(n.nodeID), // A chave pode ser o ID do nó
		Value: msgBytes,
	})
}

// SubscribeTransactions escuta novas transações da rede de forma assíncrona.
func (n *P2P) SubscribeTransactions(ctx context.Context, handler func(*transaction.Transaction)) {
	n.subscribe(ctx, TopicTransactions, n.nodeID+"-tx-group", func(val []byte) {
		var tx transaction.Transaction
		if err := json.Unmarshal(val, &tx); err == nil {
			handler(&tx)
		}
	})
}

// SubscribeBlocks escuta novos blocos propagados na rede.
func (n *P2P) SubscribeBlocks(ctx context.Context, handler func(*miner.Block)) {
	n.subscribe(ctx, TopicBlocks, n.nodeID+"-block-group", func(val []byte) {
		var block miner.Block
		if err := json.Unmarshal(val, &block); err == nil {
			handler(&block)
		}
	})
}

func (n *P2P) subscribe(ctx context.Context, topic, groupID string, handler func([]byte)) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  n.brokers,
		GroupID:  groupID, // O GroupID deve ser único por nó para que todos recebam o broadcast
		Topic:    topic,
		MaxBytes: 10e6, // 10MB
	})

	go func() {
		defer r.Close()
		for {
			m, err := r.ReadMessage(ctx)
			if err != nil {
				if ctx.Err() != nil {
					return // Context cancelado
				}
				log.Printf("erro ao ler mensagem do kafka (%s): %v", topic, err)
				continue
			}
			handler(m.Value)
		}
	}()
}