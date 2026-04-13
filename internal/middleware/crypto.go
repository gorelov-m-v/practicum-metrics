package middleware

import (
	"bytes"
	"crypto/rsa"
	"io"
	"net/http"

	"github.com/user/practicum-metrics/internal/encryption"
)

func CryptoDecrypt(privateKey *rsa.PrivateKey) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			value := r.Header.Get(encryption.HeaderEncrypted)
			if value == "" {
				next.ServeHTTP(w, r)
				return
			}

			if value != encryption.HeaderEncryptedValue {
				http.Error(w, "Unsupported encryption", http.StatusBadRequest)
				return
			}

			if privateKey == nil {
				http.Error(w, "Encryption is not configured", http.StatusBadRequest)
				return
			}

			body, err := io.ReadAll(r.Body)
			if err != nil {
				http.Error(w, "Failed to read request body", http.StatusBadRequest)
				return
			}
			r.Body.Close()

			decrypted, err := encryption.Decrypt(body, privateKey)
			if err != nil {
				http.Error(w, "Failed to decrypt request body", http.StatusBadRequest)
				return
			}

			r.Body = io.NopCloser(bytes.NewReader(decrypted))
			r.ContentLength = int64(len(decrypted))
			r.Header.Del(encryption.HeaderEncrypted)
			r.Header.Del("Content-Length")

			next.ServeHTTP(w, r)
		})
	}
}
