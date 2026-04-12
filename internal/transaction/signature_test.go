package transaction

import (
	"testing"
)

func TestSignVerifyECDSA(t *testing.T) {
	privKey, publKey, err := GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair retornou erro: %v", err)
	}

	payload := []byte("Something about us")
	//hashed_payload := sha256.Sum256(payload)
	sig, err := SignECDSA(privKey, payload)
	if err != nil {
		t.Fatalf("SignECDSA retornou erro: %v", err)
	}

	if !VerifyECDSA(publKey, payload, sig) {
		t.Fatalf("esperava que a verificacao da assinatura fosse valida para o payload original")
	}
}
