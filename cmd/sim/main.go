// package main is the entry point for the Dominium Node simulation.
// It supports dynamic port assignment and peer-to-peer neighbor configuration via Kafka.
package main

import (
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/diskrat/dominium/internal/core"
	"github.com/diskrat/dominium/internal/network"
	"github.com/diskrat/dominium/internal/transaction"
)

func main() {
	// ----------------------------------------------------------------------
	// COMMAND LINE ARGUMENTS (FLAGS)
	// Usage: go run cmd/sim/main.go -port=8080 -id=node-alpha
	// ----------------------------------------------------------------------
	port := flag.String("port", "8080", "HTTP port for the node API")
	nodeID := flag.String("id", "node-alpha", "Unique ID for Kafka P2P Group")
	initialDifficulty := flag.Int("diff", 2, "Initial mining difficulty (Nbits)")
	flag.Parse()

	fmt.Printf("--- DOMINIUM NODE [%s] STARTED ON PORT %s ---\n", *nodeID, *port)

	// 1. Initialize core components
	mempool := transaction.NewMempool(1000)
	
	// --> CORREÇÃO AQUI: Passar o *nodeID para a inicialização da Blockchain <--
	blockchain := core.NewBlockchain(int32(*initialDifficulty), *nodeID)

	// 2. Initialize Kafka P2P Broker
	broker := &network.P2PBroker{
		NodeID:  *nodeID,
		BC:      blockchain,
		Mempool: mempool,
	}

	// 3. Start listening to Kafka topics in the background
	go broker.StartConsumers()

	// 4. Start the API in a background Goroutine
	// A API não tem laço infinito bloqueante se estiver rodando via ListenAndServe,
	// mas como ela bloqueia o fluxo, é correto usar a goroutine.
	// Nota: Mudei a string da porta para apenas o número se StartAPI espera apenas a porta, 
	// mas como você colocou ":" + *port, se a flag for "8080", vai virar ":8080" que é o correto pro http.
	go network.StartAPI(":"+*port, mempool, blockchain, broker)

	// 5. Main Mining Loop
	miningInterval := 5 * time.Second
	for {
		// Só tenta minerar se houver transações pendentes
		if len(mempool.GetPendingTransactions()) > 0 {
			log.Printf("[NODE] Transactions detected. Starting mining process...")

			start := time.Now()
			// MineAndAddBlock agora sabe seu próprio ID internamente
			newBlock, err := blockchain.MineAndAddBlock(mempool)

			if err != nil {
				// Este erro é normal se as transações foram validadas por outro nó e as do mempool atual tornaram-se inválidas
				log.Printf("[NODE] Mining aborted: %v", err)
			} else {
				duration := time.Since(start)
				log.Printf("[NODE] Success! Block #%d Mined | Hash: %x | Time: %v",
					len(blockchain.GetBlocks())-1,
					newBlock.Header.Hash,
					duration,
				)

				// REQUISITO 5: Propagate the new block to all connected peers via Kafka
				broker.PublishBlock(newBlock)
			}
		} else {
			log.Printf("[NODE] Mempool idle... waiting for transactions.")
		}

		time.Sleep(miningInterval)
	}
}