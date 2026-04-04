package miner

import (
	"bytes"
	"crypto/sha256"
)

// ValidateNonce verifica se o hash atende à dificuldade definida (Nbits).
// Interpreta Nbits como a quantidade de bits iniciais que devem ser zero.
func ValidateNonce(header Header) bool {
	if len(header.Hash) == 0 {
		return false
	}

	zerosCount := 0
	for _, b := range header.Hash {
		if b == 0 {
			zerosCount += 8
		} else {
			// Conta os bits zero dentro do byte
			for i := 7; i >= 0; i-- {
				if (b>>i)&1 == 0 {
					zerosCount++
				} else {
					break
				}
			}
			break
		}
	}

	return int32(zerosCount) >= header.Nbits
}

// ValidateBlock verifica a integridade completa de um bloco que está sendo recebido.
func ValidateBlock(block Block, prevBlock *Block) bool {
	// Verifica a ligação com o bloco anterior
	if prevBlock != nil && !bytes.Equal(block.HashOfPrevious, prevBlock.Hash) {
		return false
	}

	// Verifica se a Merkle Root confere com as transações do corpo
	computedMerkleRoot := CalculateMerkleRoot(block.Transactions)
	if !bytes.Equal(block.MerkleRootHash, computedMerkleRoot) {
		return false
	}

	// Valida se o hash bate de fato com a serialização atual e com a dificuldade
	serialized, err := block.Header.Serialize()
	if err != nil {
		return false
	}
	hash := sha256.Sum256(serialized)
	
	if !bytes.Equal(hash[:], block.Hash) {
		return false
	}

	return ValidateNonce(block.Header)
}