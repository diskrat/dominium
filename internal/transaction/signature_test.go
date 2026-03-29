package transaction

import (
	"testing"
)

func TestSignVerifyECDSA(t *testing.T) {
	privKey, publKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair returned error: %v", err)
	}

	payload := []byte("Something about us")
	//hashed_payload := sha256.Sum256(payload)
	sig, err := SignECDSA(privKey, payload)
	if err != nil {
		t.Fatalf("SignECDSA returned error: %v", err)
	}

	if !VerifyECDSA(publKey, payload, sig) {
		t.Fatalf("expected signature verification to succeed for original payload")
	}
}
