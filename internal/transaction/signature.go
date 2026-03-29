// package transaction handles cryptographic validation for the network
package transaction

// internal/transaction/signature.go
import (
	"crypto/ed25519"
	"log"
)

// VerifySignature checks if a signature is valid for a given public key and data.
// This is a stateless function used strictly by validators and miners.
// It NEVER requires the Private Key.
func VerifySignature(pubKey []byte, signature []byte, data []byte) bool {
	// 1. ed25519 public keys must be exactly 32 bytes
	if len(pubKey) != ed25519.PublicKeySize {
		log.Printf("[VALIDATION] Error: Invalid public key length (%d bytes)", len(pubKey))
		return false
	}

	// 2. ed25519 signatures must be exactly 64 bytes
	if len(signature) != ed25519.SignatureSize {
		log.Printf("[VALIDATION] Error: Invalid signature length (%d bytes)", len(signature))
		return false
	}

	// 3. The magic happens here: Verify checks if the signature matches the data and pubKey
	isValid := ed25519.Verify(pubKey, data, signature)
	
	if !isValid {
		log.Printf("[VALIDATION] Error: Signature does not match data/pubkey")
	}

	return isValid
}