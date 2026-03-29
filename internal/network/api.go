package network

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/diskrat/dominium/internal/core"
	"github.com/diskrat/dominium/internal/transaction"
)

type BlockDisplay struct {
	Index              int                       `json:"index"`
	Hash               string                    `json:"hash"`
	HashOfPrevious     string                    `json:"hash_of_previous"`
	Transactions       []transaction.Transaction `json:"transactions"`
	Nonce              uint32                    `json:"nonce"`
	TransactionCounter uint16                    `json:"transaction_counter"`
	Miner              string                    `json:"miner"`
}

type TransactionRequest struct {
	SenderID   string `json:"sender"`
	ReceiverID string `json:"receiver"`
	AssetID    string `json:"asset"`
	Action     string `json:"action"`
	Fee        int64  `json:"fee"`
	PubKeyHex  string `json:"pub_key"`
	SigHex     string `json:"signature"`
}

func randomAssetGenerator() string {
	prefixes := []string{"Farm", "Lot", "Ranch", "Building", "Apartment", "Mansion", "Estate"}
	locations := []string{"North", "South", "Coast", "Mountain", "Central", "Valley", "Palm-Springs"}
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))
	return fmt.Sprintf("%s-%s-%d", prefixes[rnd.Intn(len(prefixes))], locations[rnd.Intn(len(locations))], rnd.Intn(999))
}

func generateKeysFromSeed(seedPhrase string) (ed25519.PrivateKey, ed25519.PublicKey) {
	hash := sha256.Sum256([]byte(seedPhrase))
	privKey := ed25519.NewKeyFromSeed(hash[:])
	return privKey, privKey.Public().(ed25519.PublicKey)
}

// enableCORS is a helper function to set necessary headers for external dashboard access.
func enableCORS(w *http.ResponseWriter, r *http.Request) bool {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
	
    // Handle preflight requests
	if r.Method == "OPTIONS" {
		(*w).WriteHeader(http.StatusOK)
		return true
	}
	return false
}

func StartAPI(port string, mempool *transaction.Mempool, bc *core.Blockchain, broker *P2PBroker) {

	http.HandleFunc("/blockchain", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(&w, r) { return }
		w.Header().Set("Content-Type", "application/json")

		blocks := bc.GetBlocks()
		displayChain := make([]BlockDisplay, len(blocks))

		for i, block := range blocks {
			displayChain[i] = BlockDisplay{
				Index:              i,
				Hash:               hex.EncodeToString(block.Header.Hash),
				HashOfPrevious:     hex.EncodeToString(block.Header.HashOfPrevious),
				Transactions:       block.Body.Transactions,
				Nonce:              uint32(block.Header.Nonce),
				TransactionCounter: block.Header.TransactionCounter,
				Miner:              block.Header.MinerID,
			}
		}
		json.NewEncoder(w).Encode(displayChain)
	})

	http.HandleFunc("/transactions", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(&w, r) { return }
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req TransactionRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid JSON payload", http.StatusBadRequest)
			return
		}

		pubKey, errPub := hex.DecodeString(req.PubKeyHex)
		sig, errSig := hex.DecodeString(req.SigHex)

		if errPub != nil || errSig != nil || len(pubKey) == 0 || len(sig) == 0 {
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "Missing or invalid cryptographic signature"})
			return
		}

		tx := transaction.NewAndPostTransaction(
			mempool, req.SenderID, req.ReceiverID, req.AssetID, req.Action, req.Fee, pubKey, sig,
		)

		if tx == nil {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{"error": "Validation failed, Unauthorized sender, or Double Spend detected"})
			return
		}

		if broker != nil {
			broker.PublishTransaction(tx)
		}

		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{"message": "Transaction Confirmed & Pooled", "tx_id": tx.TXid})
	})

	http.HandleFunc("/mempool", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(&w, r) { return }
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(mempool.GetPendingTransactions())
	})

	http.HandleFunc("/state", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(&w, r) { return }
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(bc.GetState())
	})

	http.HandleFunc("/difficulty", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(&w, r) { return }
		if r.Method != http.MethodPost { return }
		
		valStr := r.URL.Query().Get("value")
		val, _ := strconv.Atoi(valStr)
		bc.SetDifficulty(int32(val))
		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/simulate-tx", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(&w, r) { return }

		currentState := bc.GetState()
		rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

		var assetName string
		var action string
		var senderPriv ed25519.PrivateKey
		var senderPub ed25519.PublicKey

		if len(currentState) > 0 && rnd.Intn(100) < 40 {
			assets := make([]string, 0, len(currentState))
			for k := range currentState {
				assets = append(assets, k)
			}
			assetName = assets[rnd.Intn(len(assets))]
			action = "transfer"
			senderPriv, senderPub = generateKeysFromSeed("simulator-master-seed")
		} else {
			assetName = randomAssetGenerator()
			action = "register"
			senderPriv, senderPub = generateKeysFromSeed(fmt.Sprintf("user-%d", time.Now().UnixNano()))
		}

		senderAddr := transaction.PublicKeyToAddress(senderPub)
		tx := transaction.Transaction{
			SenderID:   senderAddr,
			ReceiverID: "New-Owner-Address",
			AssetID:    assetName,
			Action:     action,
			Fee:        int64(10 + rnd.Intn(50)),
		}

		sig := ed25519.Sign(senderPriv, tx.Serialize())
		txObj := transaction.NewAndPostTransaction(mempool, tx.SenderID, tx.ReceiverID, tx.AssetID, tx.Action, tx.Fee, senderPub, sig)

		if txObj != nil && broker != nil {
			broker.PublishTransaction(txObj)
		}

		w.WriteHeader(http.StatusOK)
	})

	http.HandleFunc("/simulate-attack", func(w http.ResponseWriter, r *http.Request) {
		if enableCORS(&w, r) { return }
		priv, pub := generateKeysFromSeed("attacker-seed-99")
		addr := transaction.PublicKeyToAddress(pub)
		target := "Copacabana-Palace"

		tx1 := transaction.Transaction{SenderID: addr, ReceiverID: "Accomplice-A", AssetID: target, Action: "transfer", Fee: 50}
		sig1 := ed25519.Sign(priv, tx1.Serialize())
		obj1 := transaction.NewAndPostTransaction(mempool, tx1.SenderID, tx1.ReceiverID, tx1.AssetID, tx1.Action, tx1.Fee, pub, sig1)

		if obj1 != nil && broker != nil {
			broker.PublishTransaction(obj1)
		}

		tx2 := transaction.Transaction{SenderID: addr, ReceiverID: "Accomplice-B", AssetID: target, Action: "transfer", Fee: 99}
		sig2 := ed25519.Sign(priv, tx2.Serialize())
		obj2 := transaction.NewAndPostTransaction(mempool, tx2.SenderID, tx2.ReceiverID, tx2.AssetID, tx2.Action, tx2.Fee, pub, sig2)

		if obj2 != nil && broker != nil {
			broker.PublishTransaction(obj2)
		}

		w.WriteHeader(http.StatusOK)
	})

	fs := http.FileServer(http.Dir("web"))
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.ServeFile(w, r, "web/index.html")
			return
		}
		fs.ServeHTTP(w, r)
	})

	log.Printf("[API] Server listening on http://0.0.0.0%s", port)

	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatalf("[API] Server failed: %v", err)
	}
}