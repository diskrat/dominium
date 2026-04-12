package network

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"dominium/internal/miner"
	"dominium/internal/transaction"

	"github.com/segmentio/kafka-go"
)

const (
	TopicTransactions = "dominium-transactions"
	TopicBlocks       = "dominium-blocks"
)

// KafkaTransactionTransport implementa TransactionSink e TransactionSource usando Kafka.
type KafkaTransactionTransport struct {
	ctx     context.Context
	cancel  context.CancelFunc
	brokers []string
	nodeID  string
}

// KafkaBlockTransport implementa BlockSink e BlockSource usando Kafka.
type KafkaBlockTransport struct {
	ctx     context.Context
	cancel  context.CancelFunc
	brokers []string
	nodeID  string
}

// NewKafkaTransactionTransport cria um transportador Kafka para transações.
func NewKafkaTransactionTransport(parent context.Context, brokers []string, nodeID string) *KafkaTransactionTransport {
	ctx, cancel := context.WithCancel(parent)
	return &KafkaTransactionTransport{ctx: ctx, cancel: cancel, brokers: brokers, nodeID: nodeID}
}

// NewKafkaBlockTransport cria um transportador Kafka para blocos.
func NewKafkaBlockTransport(parent context.Context, brokers []string, nodeID string) *KafkaBlockTransport {
	ctx, cancel := context.WithCancel(parent)
	return &KafkaBlockTransport{ctx: ctx, cancel: cancel, brokers: brokers, nodeID: nodeID}
}

func (t *KafkaTransactionTransport) Publish(tx *transaction.Transaction) error {
	data, err := json.Marshal(tx)
	if err != nil {
		return err
	}

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  t.brokers,
		Topic:    TopicTransactions,
		Balancer: &kafka.LeastBytes{},
	})
	defer w.Close()

	ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	return w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(t.nodeID),
		Value: data,
	})
}

func (t *KafkaTransactionTransport) Subscribe(handler func(*transaction.Transaction) error) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  t.brokers,
		GroupID:  t.nodeID + "-tx-group",
		Topic:    TopicTransactions,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	go func() {
		defer r.Close()
		for {
			m, err := r.ReadMessage(t.ctx)
			if err != nil {
				if t.ctx.Err() != nil {
					return
				}
				log.Printf("erro ao ler transacao do kafka: %v", err)
				continue
			}

			var tx transaction.Transaction
			if err := json.Unmarshal(m.Value, &tx); err != nil {
				log.Printf("transacao kafka invalida: %v", err)
				continue
			}

			if err := handler(&tx); err != nil {
				log.Printf("erro no handler de transacao: %v", err)
			}
		}
	}()

	return nil
}

func (t *KafkaBlockTransport) Publish(block *miner.Block) error {
	data, err := json.Marshal(block)
	if err != nil {
		return err
	}

	w := kafka.NewWriter(kafka.WriterConfig{
		Brokers:  t.brokers,
		Topic:    TopicBlocks,
		Balancer: &kafka.LeastBytes{},
	})
	defer w.Close()

	ctx, cancel := context.WithTimeout(t.ctx, 5*time.Second)
	defer cancel()

	return w.WriteMessages(ctx, kafka.Message{
		Key:   []byte(t.nodeID),
		Value: data,
	})
}

func (t *KafkaBlockTransport) Subscribe(handler func(*miner.Block) error) error {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers:  t.brokers,
		GroupID:  t.nodeID + "-block-group",
		Topic:    TopicBlocks,
		MinBytes: 1,
		MaxBytes: 10e6,
	})

	go func() {
		defer r.Close()
		for {
			m, err := r.ReadMessage(t.ctx)
			if err != nil {
				if t.ctx.Err() != nil {
					return
				}
				log.Printf("erro ao ler bloco do kafka: %v", err)
				continue
			}

			var block miner.Block
			if err := json.Unmarshal(m.Value, &block); err != nil {
				log.Printf("bloco kafka invalido: %v", err)
				continue
			}

			if err := handler(&block); err != nil {
				log.Printf("erro no handler de bloco: %v", err)
			}
		}
	}()

	return nil
}

func (t *KafkaTransactionTransport) Stop() {
	t.cancel()
}

func (t *KafkaBlockTransport) Stop() {
	t.cancel()
}
