// package main simulates a client wallet sending valid transactions to the node.
package main

//internal/client/main.go
import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io" // Importante para ler o corpo do erro
	"log"
	"math/rand"
	"net/http"
	"time"

	"github.com/diskrat/dominium/internal/transaction"
)

type TransactionRequest struct {
	SenderID   string `json:"sender"`
	ReceiverID string `json:"receiver"`
	AssetID    string `json:"asset"`
	Action     string `json:"action"`
	Fee        int64  `json:"fee"`
	PubKeyHex  string `json:"pub_key"`
	SigHex     string `json:"signature"`
}

func GenerateKeysFromSeed(seedPhrase string) (ed25519.PrivateKey, ed25519.PublicKey) {
	hash := sha256.Sum256([]byte(seedPhrase))
	privKey := ed25519.NewKeyFromSeed(hash[:])
	pubKey := privKey.Public().(ed25519.PublicKey)
	return privKey, pubKey
}

func main() {
	fmt.Println("--- DOMINIUM CLIENT: STARTING TRANSACTION GENERATOR ---")

	seedPhrase := "minha-semente-secreta-projeto-ufrn-2026"
	privKey, pubKey := GenerateKeysFromSeed(seedPhrase)
	senderAddr := transaction.PublicKeyToAddress(pubKey)

	log.Printf("[CLIENT] Identity Loaded. Address: %s", senderAddr[:16]+"...")

	apiURL := "http://localhost:8080/transactions"
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 1; i <= 5; i++ {
		txCore := transaction.Transaction{
			SenderID:   senderAddr,
			ReceiverID: "Cartorio-Central-RN",
			AssetID:    fmt.Sprintf("Matricula-Terra-%d", r.Intn(9999)),
			Action:     "register",
			Fee:        int64(10 + r.Intn(50)),
		}

		dataToSign := txCore.Serialize()
		signature := ed25519.Sign(privKey, dataToSign)

		payload := TransactionRequest{
			SenderID:   txCore.SenderID,
			ReceiverID: txCore.ReceiverID,
			AssetID:    txCore.AssetID,
			Action:     txCore.Action,
			Fee:        txCore.Fee,
			PubKeyHex:  hex.EncodeToString(pubKey),
			SigHex:     hex.EncodeToString(signature),
		}

		jsonData, _ := json.Marshal(payload)

		log.Printf("[CLIENT] Sending Tx: Asset %s (Fee: %d)", payload.AssetID, payload.Fee)
		resp, err := http.Post(apiURL, "application/json", bytes.NewBuffer(jsonData))

		if err != nil {
			log.Fatalf("[CLIENT] Error connecting to Node: %v", err)
		}

		// AQUI ESTÁ A MÁGICA PARA DESCOBRIR O ERRO:
		if resp.StatusCode == http.StatusCreated {
			log.Printf("[CLIENT] Success! Node accepted the transaction.")
		} else {
			bodyBytes, _ := io.ReadAll(resp.Body)
			log.Printf("[CLIENT] Rejected! Status: %d | Motivo: %s", resp.StatusCode, string(bodyBytes))
		}
		
		resp.Body.Close()
		time.Sleep(3 * time.Second)
	}
	
	fmt.Println("[CLIENT] Finished sending batch.")
}