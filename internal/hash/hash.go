// Package hash provides HMAC-SHA256 signing and verification utilities.
package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// CalculateHMAC computes HMAC-SHA256 for data using the given key.
// Returns empty string if key is empty.
func CalculateHMAC(data []byte, key string) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC checks that receivedHash matches the HMAC-SHA256 of data.
func VerifyHMAC(data []byte, receivedHash string, key string) bool {
	if key == "" {
		return receivedHash == ""
	}
	expectedHash := CalculateHMAC(data, key)
	return hmac.Equal([]byte(expectedHash), []byte(receivedHash))
}
