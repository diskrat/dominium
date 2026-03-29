package transaction

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
)

func SignECDSA(privKey *ecdsa.PrivateKey, data []byte) ([]byte, error) {
	hash := sha256.Sum256(data)
	signature, err := ecdsa.SignASN1(rand.Reader, privKey, hash[:])
	if err != nil {
		return nil, err
	}
	return signature, nil
}

func VerifyECDSA(publiKey *ecdsa.PublicKey, data []byte, signature []byte) bool {
	hash := sha256.Sum256(data)
	isValid := ecdsa.VerifyASN1(publiKey, hash[:], signature)
	return isValid
}
