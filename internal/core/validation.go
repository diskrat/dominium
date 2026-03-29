// package core handles the consensus rules and validation logic for the blockchain.
package core

import (
	"bytes"
	"log"
)

// IsValidNewBlock checks the cryptographic and logical links between two blocks.
// It ensures that the candidate block follows all network rules before being appended.
func IsValidNewBlock(newBlock, prevBlock *Block) bool {
	
	// 1. Link Integrity: The new block must point to the correct previous block hash.
	if !bytes.Equal(newBlock.Header.HashOfPrevious, prevBlock.Header.Hash) {
		log.Printf("[VALIDATION] Failed: Previous hash mismatch.")
		return false
	}

	// 2. Counter Integrity: Transaction count in header must match the actual body size.
	if newBlock.Header.TransactionCounter != uint16(len(newBlock.Body.Transactions)) {
		log.Printf("[VALIDATION] Failed: Transaction counter mismatch.")
		return false
	}

	// 3. Content Integrity: Re-calculate Merkle Root to ensure transactions haven't been tampered with.
	merkle := GenMerkleRoot(newBlock.Body.Transactions)
	if !bytes.Equal(newBlock.Header.MerkleRootHash, merkle) {
		log.Printf("[VALIDATION] Failed: Merkle Root mismatch.")
		return false
	}

	// 4. Hash Integrity: Re-calculate the header hash to verify internal consistency.
	calculated := CalculateBlockHash(newBlock.Header)
	if !bytes.Equal(newBlock.Header.Hash, calculated) {
		log.Printf("[VALIDATION] Failed: Calculated hash does not match header hash.")
		return false
	}

	// 5. Proof of Work Integrity: Verify the hash meets the N-zero hex characters target.
	// FIXED: We now use CheckDifficulty (from pow.go) to count hex zeros instead of full bytes.
	if !CheckDifficulty(newBlock.Header.Hash, newBlock.Header.Nbits) {
		log.Printf("[VALIDATION] Failed: Hash does not meet difficulty requirements (%d hex zeros).", newBlock.Header.Nbits)
		return false
	}

	return true
}

// IsValidChain verifies the integrity of an entire chain from genesis to the tip.
// This is used during the "Longest Chain" resolution (ReplaceChain).
func IsValidChain(chain []*Block) bool {
	if len(chain) == 0 {
		return false
	}

	// Verify Genesis Block Integrity
	// Genesis has no predecessor, so we check if its PreviousHash is empty (32 zero bytes).
	genesis := chain[0]
	if !bytes.Equal(genesis.Header.HashOfPrevious, make([]byte, 32)) {
		log.Printf("[VALIDATION] Failed: Invalid Genesis block link.")
		return false
	}

	// Verify each subsequent block in the chain recursively
	for i := 1; i < len(chain); i++ {
		if !IsValidNewBlock(chain[i], chain[i-1]) {
			log.Printf("[VALIDATION] Failed: Chain broken at block #%d.", i)
			return false
		}
	}
	
	return true
}