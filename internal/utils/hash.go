package utils

import (
	"crypto/rand"
	"encoding/base64"

	"golang.org/x/crypto/bcrypt"
)

// GenerateHashedAPIKey creates a secure API key and its bcrypt hash
func GenerateHashedAPIKey() (string, error) {
	// 32 random bytes
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}

	// URL-safe key
	apiKey := "sk_" + base64.RawURLEncoding.EncodeToString(bytes)

	// Hash it
	hashed, err := bcrypt.GenerateFromPassword([]byte(apiKey), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	return string(hashed), nil
}
