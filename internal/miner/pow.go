package miner

import (
	"context"
	"crypto/sha256"
	"runtime"
	"sync"
	"sync/atomic"
)

// Mine realiza o Proof of Work para o bloco de forma concorrente.
// Ele distribui o espaço de busca do Nonce entre as CPUs disponíveis.
func Mine(block *Block) {
	numWorkers := runtime.NumCPU()
	var wg sync.WaitGroup
	var found int32 // flag atômica para encerrar as goroutines

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	for w := 0; w < numWorkers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			
			// Cópia local do header para que cada goroutine altere seu próprio nonce livre de mutexes
			headerCopy := block.Header
			
			// Dividindo o trabalho: cada worker pula a contagem de acordo com numWorkers
			for nonce := int32(workerID); ; nonce += int32(numWorkers) {
				// Verifica se algum outro worker já encontrou
				if atomic.LoadInt32(&found) == 1 {
					return
				}

				select {
				case <-ctx.Done():
					return
				default:
					headerCopy.Nonce = nonce
					serialized, _ := headerCopy.Serialize()
					hash := sha256.Sum256(serialized)
					headerCopy.Hash = hash[:]

					if ValidateNonce(headerCopy) {
						// Se achou, tenta ser o primeiro a setar a flag
						if atomic.CompareAndSwapInt32(&found, 0, 1) {
							block.Nonce = nonce
							block.Hash = hash[:]
							cancel() // Avisa as demais goroutines para parar
						}
						return
					}
				}
			}
		}(w)
	}

	wg.Wait()
}