package miner

import (
	"crypto/sha256"
	"dominium/internal/transaction"
)

// CalculateMerkleRoot calcula a raiz de Merkle (Merkle Root Hash) para uma lista de transações.
// O processo agrupa transações em pares e gera o hash concatenado recursivamente.
// Se houver um número ímpar de nós, o último nó é duplicado.
func CalculateMerkleRoot(txs []transaction.Transaction) []byte {
	if len(txs) == 0 {
		return []byte{}
	}

	var hashes [][]byte

	// Nível folha: hashes das transações
	for _, tx := range txs {
		payload, _ := tx.Serialize() // Ignorando erro para simplicidade, embora no block real seja bom tratar
		hash := sha256.Sum256(payload)
		hashes = append(hashes, hash[:])
	}

	// Redução iterativa até sobrar apenas a raiz (1 hash)
	for len(hashes) > 1 {
		// Se for ímpar, duplica o último
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		var nextLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			combined := append(hashes[i], hashes[i+1]...)
			newHash := sha256.Sum256(combined)
			nextLevel = append(nextLevel, newHash[:])
		}
		hashes = nextLevel
	}

	return hashes[0]
}