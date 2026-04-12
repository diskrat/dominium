// package core handles the mining process and block assembly logic.
package core

import (
	"time"

	"github.com/diskrat/dominium/internal/transaction"
)

// Miner represents the entity responsible for solving the Proof of Work.
type Miner struct {
	ID string
}

// NewMiner creates a new Miner instance.
func NewMiner(id string) *Miner {
	return &Miner{ID: id}
}

// BlockMiner performs the Proof of Work loop to find a valid hash.
// It increments the Nonce until the hex representation of the hash starts with N zeros.
func (m *Miner) BlockMiner(block *Block) []byte {
	for {
		// Calculate the raw hash
		hash := CalculateBlockHash(block.Header) 
		
		// NEW LOGIC: Check if the HEX string has the required number of zeros.
		// Difficulty 4 now correctly requires "0000", which is much faster than 4 full bytes.
		if CheckDifficulty(hash, block.Header.Nbits) {
			block.Header.Hash = hash
			return hash
		}
		
		// Increment nonce to change the hash in the next iteration
		block.Header.Nonce++
	}
}

// AssemblerNextBlockMiner prepares a new block using a pre-validated list of transactions.
func AssemblerNextBlockMiner(validatedTxs []transaction.Transaction, prevHash []byte, difficulty int32) *Block {
	body := Body{Transactions: validatedTxs}

	header := Header{
		HashOfPrevious:     prevHash,
		MerkleRootHash:     GenMerkleRoot(validatedTxs),
		Timestamp:          time.Now().Unix(),
		Nbits:              difficulty,
		TransactionCounter: uint16(len(validatedTxs)),
		Nonce:              0, 
	}

	return &Block{
		Header: header, 
		Body:   body,
	}
}