package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
)

// TestKeyPair returns a new ECDSA key pair and its encoded public key.
func TestKeyPair() (*ecdsa.PrivateKey, string) {
	sk, _ := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	return sk, EncodePublicKey(&sk.PublicKey)
}
