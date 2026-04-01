package transaction

import (
	"crypto/x509"
	"encoding/hex"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	_, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair retornou erro: %v", err)
	}
}

func TestEncodeDecodePublicKey(t *testing.T) {
	_, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair retornou erro: %v", err)
	}

	encoded := EncodePublicKey(publicKey)

	decoded, err := DecodePublicKey(encoded)
	if err != nil {
		t.Fatalf("DecodePublicKey retornou erro: %v", err)
	}

	decodedPKIX, err := x509.MarshalPKIXPublicKey(decoded)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey retornou erro: %v", err)
	}
	decodedHex := hex.EncodeToString(decodedPKIX)
	if decodedHex != encoded {
		t.Fatalf("chave publica divergente apos ida e volta encode/decode: antes=%s depois=%s", encoded, decodedHex)
	}
}

func TestDecodePublicKeyInvalidInput(t *testing.T) {
	if _, err := DecodePublicKey("invalid-hex"); err == nil {
		t.Fatal("esperava erro para entrada invalida")
	}
}
