// Package network handles external communication, including the HTTP API and P2P messaging.
// This file implements a Peer-to-Peer (P2P) communication layer using Apache Kafka as a message broker.
// It allows multiple blockchain nodes to stay synchronized by broadcasting blocks and transactions.
package network

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/diskrat/dominium/internal/core"
	"github.com/diskrat/dominium/internal/transaction"
	"github.com/segmentio/kafka-go"
)

// getKafkaURL determines the Kafka broker address.
// It prioritizes the environment variable set by Docker Compose.
func getKafkaURL() string {
	url := os.Getenv("KAFKA_URL")
	if url == "" {
		return "localhost:9092" // Fallback for local execution without Docker
	}
	return url
}

const (
	// Topics used for network-wide broadcasting
	TopicBlocks = "dominium-blocks"
	TopicTxs    = "dominium-txs"
)

// P2PBroker manages Kafka communication for a specific node instance.
// It acts as the bridge between the local data (Blockchain/Mempool) and the network.
type P2PBroker struct {
	NodeID  string               // Unique identifier for this node (e.g., node-alpha)
	BC      *core.Blockchain     // Reference to the local confirmed ledger
	Mempool *transaction.Mempool // Reference to the local pending transaction pool
}

// PublishBlock broadcasts a newly mined block to the entire network.
// This fulfills the requirement for block dissemination in a P2P environment.
func (p *P2PBroker) PublishBlock(block *core.Block) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{getKafkaURL()},
		Topic:   TopicBlocks,
	})
	defer writer.Close()

	// Convert block to JSON for transmission
	blockBytes, _ := json.Marshal(block)
	
	// Tag message with NodeID to prevent the sender from re-processing its own block
	err := writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(p.NodeID), 
		Value: blockBytes,
	})

	if err != nil {
		log.Printf("[KAFKA] Publish Error: %v", err)
	} else {
		log.Printf("[KAFKA] Block #%d broadcasted to peers!", len(p.BC.GetBlocks())-1)
	}
}

// PublishTransaction sends a new transaction to all nodes' mempools.
// This ensures all miners in the network are aware of the pending task.
func (p *P2PBroker) PublishTransaction(tx *transaction.Transaction) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{getKafkaURL()},
		Topic:   TopicTxs,
	})
	defer writer.Close()

	txBytes, _ := json.Marshal(tx)
	
	err := writer.WriteMessages(context.Background(), kafka.Message{
		Value: txBytes,
	})
	
	if err != nil {
		log.Printf("[KAFKA] Tx Broadcast Error: %v", err)
	}
}

// StartConsumers launches background goroutines to listen for network events.
func (p *P2PBroker) StartConsumers() {
	go p.consumeBlocks()
	go p.consumeTxs()
}

// consumeBlocks monitors the block topic and updates the local chain upon receiving peer blocks.
func (p *P2PBroker) consumeBlocks() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{getKafkaURL()},
		Topic:   TopicBlocks,
		GroupID: p.NodeID,
	})
	defer reader.Close()

	log.Printf("[KAFKA] Node [%s] listening for peer blocks...", p.NodeID)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil { continue }

		// Filter out self-broadcasted blocks
		if string(m.Key) == p.NodeID { continue }

		var receivedBlock core.Block
		json.Unmarshal(m.Value, &receivedBlock)

		log.Printf("[KAFKA] Block received from peer [%s]", string(m.Key))

		// Strict validation of the incoming block
		err = p.BC.AddBlock(&receivedBlock)
		
		if err != nil {
			// ==========================================
			// FORK RESOLUTION (Tie-Breaker)
			// ==========================================
			resolved := p.BC.ResolveTie(&receivedBlock)
			if resolved {
				log.Printf("[CONSENSUS] Tie-Breaker lost! Replaced our block with peer's block: %x", receivedBlock.Header.Hash[:10])
			} else {
				log.Printf("[CONSENSUS] Block rejected: %v", err)
			}
		} else {
			log.Printf("[CONSENSUS] Peer block accepted. State synchronized.")
		}
	}
}

// consumeTxs monitors the transaction topic and populates the local mempool.
func (p *P2PBroker) consumeTxs() {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers: []string{getKafkaURL()},
		Topic:   TopicTxs,
		GroupID: p.NodeID,
	})
	defer reader.Close()

	log.Printf("[KAFKA] Node [%s] listening for network transactions...", p.NodeID)

	for {
		m, err := reader.ReadMessage(context.Background())
		if err != nil { continue }

		var tx transaction.Transaction
		json.Unmarshal(m.Value, &tx)

		// Validate incoming P2P transaction against local state before adding to pool
		validTxs, _ := p.BC.ValidateTransactionsAgainstState([]transaction.Transaction{tx})
		
		if len(validTxs) > 0 {
			err := p.Mempool.AddTransactionToMempool(validTxs[0])
			if err == nil {
				log.Printf("[KAFKA] Remote Transaction [%s] added to mempool", tx.AssetID)
			}
		}
	}
}