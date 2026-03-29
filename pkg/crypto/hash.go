package crypto

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
)

func Hash(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

func HashObject(v any) (string, error) {
	bytes, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	return Hash(bytes), nil
}
