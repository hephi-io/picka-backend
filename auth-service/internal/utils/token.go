package utils

import (
	"crypto/rand"
	"encoding/hex"
)

// GenerateRandomToken returns a securely generated random token of n bytes, hex-encoded.
func GenerateRandomToken(n int) (string, error) {
	b := make([]byte, n)
	_, err := rand.Read(b)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
