package transaction

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"encoding/hex"
	"errors"
)

func GenerateKeyPair() (*ecdsa.PrivateKey, *ecdsa.PublicKey, error) {
	privateKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	publicKey := &privateKey.PublicKey
	return privateKey, publicKey, nil
}

func EncodePublicKey(pub *ecdsa.PublicKey) string {
	if pub == nil {
		return ""
	}
	pubBytes, err := x509.MarshalPKIXPublicKey(pub)
	if err != nil {
		return ""
	}
	return hex.EncodeToString(pubBytes)
}

func DecodePublicKey(pubHex string) (*ecdsa.PublicKey, error) {
	pubBytes, err := hex.DecodeString(pubHex)
	if err != nil {
		return nil, err
	}

	parsed, err := x509.ParsePKIXPublicKey(pubBytes)
	if err != nil {
		return nil, err
	}

	pub, ok := parsed.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New("invalid public key type")
	}
	return pub, nil
}
