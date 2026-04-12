// package core defines the essential data structures for the blockchain ledger.
package core

//internal\core\block.go
import (
	. "github.com/diskrat/dominium/internal/transaction"
)

// Header contains the metadata of a block.
// It is used by miners to perform Proof of Work and by nodes to verify the chain's integrity.
type Header struct {
	HashOfPrevious     []byte // the SHA-256 hash of the previous block's header.
	MerkleRootHash     []byte // the root hash of all transactions included in the block's body.
	Timestamp          int64  // the Unix time when the block was created.
	Nbits              int32  // the difficulty target for the mining process (number of required leading zeros).
	Nonce              int32  // the counter used by miners to find a valid hash for Proof of Work.
	TransactionCounter uint16 // the total number of transactions contained within the block.
	Hash               []byte // the final valid hash of this current block header.
	MinerID            string
}

// Body represents the payload of the block.
// It stores the list of validated transactions that have been confirmed.
type Body struct {
	Transactions []Transaction // slice of all transactions processed in this block.
}

// Block represents a single link in the blockchain.
// It is composed of a Header (metadata) and a Body (transaction data).
type Block struct {
	Header // embedded header structure.
	Body   // embedded body structure.
}

