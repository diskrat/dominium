// package transaction handles the generation and management of cryptographic identities.
package transaction

// internal/transaction/keys.go
import (
	"crypto/ed25519"
	"crypto/rand"
	"fmt"
	"log"
)

// GenerateKeyPair creates a new ed25519 private and public key pair.
// this is used when a new user joins the network or creates a new wallet.
func GenerateKeyPair() (ed25519.PrivateKey, ed25519.PublicKey, error) {
	// generate the key pair using a cryptographically secure random source.
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		log.Printf("[KEYS] Error: failed to generate key pair: %v", err)
		return nil, nil, err
	}

	log.Printf("[KEYS] Success: new key pair generated")
	return priv, pub, nil
}

// PublicKeyToAddress converts a public key into a human-readable string.
// in many blockchains, this would involve extra hashing (like ripemd160), 
// but for now, we can hex-encode the public key.
func PublicKeyToAddress(pub ed25519.PublicKey) string {
	return fmt.Sprintf("%x", pub)
}