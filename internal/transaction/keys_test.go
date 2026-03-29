package transaction

import (
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"testing"
)

func TestGenerateKeyPair(t *testing.T) {
	_, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair returned error: %v", err)
	}
}

func TestEncodeDecodePublicKey(t *testing.T) {
	_, publicKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair returned error: %v", err)
	}

	encoded := EncodePublicKey(publicKey)

	decoded, err := DecodePublicKey(encoded)
	if err != nil {
		t.Fatalf("DecodePublicKey returned error: %v", err)
	}

	decodedPKIX, err := x509.MarshalPKIXPublicKey(decoded)
	if err != nil {
		t.Fatalf("MarshalPKIXPublicKey returned error: %v", err)
	}
	decodedHex := hex.EncodeToString(decodedPKIX)
	if decodedHex != encoded {
		t.Fatalf("public key mismatch after encode/decode round-trip: before=%s after=%s", encoded, decodedHex)
	}
}

func TestDecodePublicKeyInvalidInput(t *testing.T) {
	if _, err := DecodePublicKey("invalid-hex"); err == nil {
		t.Fatal("expected error for invalid input")
	}
}

func TestPrivateKeyPrintHexAndBase64(t *testing.T) {
	privateKey, _, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair returned error: %v", err)
	}

	der, err := x509.MarshalECPrivateKey(privateKey)
	if err != nil {
		t.Fatalf("MarshalECPrivateKey returned error: %v", err)
	}

	t.Logf("private key (hex DER): %s", hex.EncodeToString(der))
	t.Logf("private key (base64 DER): %s", base64.StdEncoding.EncodeToString(der))
}
