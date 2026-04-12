// package core provides the cryptographic math for the Proof of Work and Merkle Trees.
package core

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/diskrat/dominium/internal/transaction"
)

// CalculateBlockHash computes the SHA-256 hash of a block header.
// It unifies the hashing logic so both the Miner and the Validator use the same rules.
func CalculateBlockHash(h Header) []byte {
	// We serialize the header fields into a string before hashing.
	data := fmt.Sprintf("%x%x%d%d%d%s",
		h.HashOfPrevious, h.MerkleRootHash, h.Timestamp, h.Nbits, h.Nonce, h.MinerID)

	hash := sha256.Sum256([]byte(data))
	return hash[:]
}

// CheckDifficulty verifies if a hash string starts with N hexadecimal zeros.
// This ensures that Difficulty 4 means "0000...", not 4 full bytes.
func CheckDifficulty(hash []byte, difficulty int32) bool {
	hashHex := hex.EncodeToString(hash)
	prefix := strings.Repeat("0", int(difficulty))
	return strings.HasPrefix(hashHex, prefix)
}

// GenMerkleRoot calculates the Merkle Root (hash of hashes) for a transaction list.
func GenMerkleRoot(transactions []transaction.Transaction) []byte {
	var hashes [][]byte

	for _, tx := range transactions {
		txData := tx.Serialize()
		hashTx := sha256.Sum256(txData)
		hashes = append(hashes, hashTx[:])
	}

	if len(hashes) == 0 {
		return make([]byte, 32) // Return empty hash for empty blocks
	}

	// Iteratively pair and hash until only one root remains
	for len(hashes) > 1 {
		if len(hashes)%2 != 0 {
			hashes = append(hashes, hashes[len(hashes)-1])
		}

		var upperLevel [][]byte
		for i := 0; i < len(hashes); i += 2 {
			combined := append(hashes[i], hashes[i+1]...)
			parentHash := sha256.Sum256(combined)
			upperLevel = append(upperLevel, parentHash[:])
		}
		hashes = upperLevel
	}

	return hashes[0]
}