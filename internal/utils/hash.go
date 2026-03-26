package utils

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
)

// HashKey returns a deterministic SHA-256 hex digest of the key, suitable for DB lookup.
func HashKey(key string) (string, error) {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:]), nil
}

// GenerateHashedAPIKey creates a secure API key and its bcrypt hash
func GenerateHashedAPIKey() (string, string, error) {
	// 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", "", err
	}

	// URL-safe key
	apiKey := "sk_" + base64.RawURLEncoding.EncodeToString(bytes)

	// Hash it
	hashed, err := HashKey(apiKey)
	if err != nil {
		return "", "", err
	}

	return apiKey, string(hashed), nil
}
