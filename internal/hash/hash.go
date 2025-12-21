package hash

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
)

// CalculateHMAC вычисляет HMAC SHA256 хеш для данных с использованием ключа
func CalculateHMAC(data []byte, key string) string {
	if key == "" {
		return ""
	}
	h := hmac.New(sha256.New, []byte(key))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// VerifyHMAC проверяет соответствие HMAC хеша данным и ключу
func VerifyHMAC(data []byte, receivedHash string, key string) bool {
	if key == "" {
		return receivedHash == ""
	}
	expectedHash := CalculateHMAC(data, key)
	return hmac.Equal([]byte(expectedHash), []byte(receivedHash))
}
