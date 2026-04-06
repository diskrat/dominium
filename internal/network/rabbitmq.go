package network

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

var (
	ErrConnectionFailed = fmt.Errorf("failed to connect to RabbitMQ")
	ErrChannelFailed    = fmt.Errorf("failed to open channel")
	ErrPublishFailed    = fmt.Errorf("failed to publish message")
	ErrConsumeFailed    = fmt.Errorf("failed to consume message")
)

type MessageHandler func(msg *Message) error

type RabbitMQ struct {
	conn         *amqp.Connection
	channel      *amqp.Channel
	peerID       string
	exchangeName string
	queueName    string
	handlers     map[MessageType]MessageHandler
	mu           sync.RWMutex
	closed       bool
}

func NewRabbitMQ(peerID, amqpURL string) (*RabbitMQ, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConnectionFailed, err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("%w: %v", ErrChannelFailed, err)
	}

	exchangeName := "blockchain"
	err = ch.ExchangeDeclare(
		exchangeName,
		"topic",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare exchange: %w", err)
	}

	queueName := fmt.Sprintf("node_%s", peerID)
	_, err = ch.QueueDeclare(
		queueName,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to declare queue: %w", err)
	}

	rabbitmq := &RabbitMQ{
		conn:         conn,
		channel:      ch,
		peerID:       peerID,
		exchangeName: exchangeName,
		queueName:    queueName,
		handlers:     make(map[MessageType]MessageHandler),
	}

	if err := rabbitmq.bindQueue(); err != nil {
		rabbitmq.Close()
		return nil, err
	}

	return rabbitmq, nil
}

func (r *RabbitMQ) bindQueue() error {
	routingKeys := []string{
		"block.*",
		"tx.*",
		"miner.*",
		"sync.*",
	}

	for _, key := range routingKeys {
		err := r.channel.QueueBind(
			r.queueName,
			key,
			r.exchangeName,
			false,
			nil,
		)
		if err != nil {
			return fmt.Errorf("failed to bind queue to %s: %w", key, err)
		}
	}

	return nil
}

func (r *RabbitMQ) RegisterHandler(msgType MessageType, handler MessageHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[msgType] = handler
}

func (r *RabbitMQ) Publish(routingKey string, msg *Message) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.closed {
		return fmt.Errorf("rabbitmq connection is closed")
	}

	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	err = r.channel.PublishWithContext(
		ctx,
		r.exchangeName,
		routingKey,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrPublishFailed, err)
	}

	return nil
}

func (r *RabbitMQ) PublishToAll(routingKey string, msg *Message) error {
	return r.Publish(routingKey, msg)
}

func (r *RabbitMQ) BroadcastBlock(blockJSON string, sender string) error {
	msg := NewMessage(MsgBlockNew, BlockMessage{BlockJSON: blockJSON}, sender)
	return r.Publish("block.new", msg)
}

func (r *RabbitMQ) BroadcastTransaction(txJSON string, sender string) error {
	msg := NewMessage(MsgTxNew, TxMessage{TxJSON: txJSON}, sender)
	return r.Publish("tx.new", msg)
}

func (r *RabbitMQ) BroadcastMinerRegister(peerInfo PeerInfo) error {
	msg := NewMessage(MsgMinerRegister, MinerRegisterMessage{PeerInfo: peerInfo}, peerInfo.ID)
	return r.Publish("miner.register", msg)
}

func (r *RabbitMQ) RequestChainSync(fromIndex int, sender string) error {
	msg := NewMessage(MsgChainSync, ChainSyncMessage{FromIndex: fromIndex}, sender)
	return r.Publish("sync.request", msg)
}

func (r *RabbitMQ) StartConsuming() error {
	msgs, err := r.channel.Consume(
		r.queueName,
		"",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("%w: %v", ErrConsumeFailed, err)
	}

	go func() {
		for {
			if r.isClosed() {
				return
			}

			select {
			case msg, ok := <-msgs:
				if !ok {
					return
				}

				r.handleMessage(msg.Body)

				if err := msg.Ack(false); err != nil {
					log.Printf("Failed to ack message: %v", err)
				}

			case <-time.After(1 * time.Second):
				continue
			}
		}
	}()

	return nil
}

func (r *RabbitMQ) handleMessage(body []byte) {
	var msg Message
	if err := json.Unmarshal(body, &msg); err != nil {
		log.Printf("Failed to unmarshal message: %v", err)
		return
	}

	if msg.Sender == r.peerID {
		return
	}

	r.mu.RLock()
	handler, exists := r.handlers[msg.Type]
	r.mu.RUnlock()

	if exists {
		if err := handler(&msg); err != nil {
			log.Printf("Error handling message type %s: %v", msg.Type, err)
		}
	}
}

func (r *RabbitMQ) isClosed() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.closed
}

func (r *RabbitMQ) Close() error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		return nil
	}

	r.closed = true

	if r.channel != nil {
		r.channel.Close()
	}
	if r.conn != nil {
		r.conn.Close()
	}

	return nil
}

func (r *RabbitMQ) GetPeerID() string {
	return r.peerID
}

func (r *RabbitMQ) IsConnected() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.conn == nil || r.closed {
		return false
	}

	return !r.conn.IsClosed()
}
