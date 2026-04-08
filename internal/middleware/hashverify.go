package middleware

import (
	"bytes"
	"io"
	"net/http"

	"github.com/user/practicum-metrics/internal/hash"
)

type hashResponseWriter struct {
	http.ResponseWriter
	body *bytes.Buffer
	key  string
}

func (w *hashResponseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// HashVerify returns a middleware that verifies HMAC-SHA256 signatures
// on incoming requests and signs outgoing responses using the provided key.
func HashVerify(key string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if key == "" {
				next.ServeHTTP(w, r)
				return
			}

			if r.Method == http.MethodPost || r.Method == http.MethodPut {
				receivedHash := r.Header.Get("HashSHA256")

				body, err := io.ReadAll(r.Body)
				if err != nil {
					http.Error(w, "Failed to read request body", http.StatusBadRequest)
					return
				}
				r.Body.Close()

				if receivedHash != "" {
					if !hash.VerifyHMAC(body, receivedHash, key) {
						http.Error(w, "Hash verification failed", http.StatusBadRequest)
						return
					}
				}

				r.Body = io.NopCloser(bytes.NewBuffer(body))
			}

			hrw := &hashResponseWriter{
				ResponseWriter: w,
				body:           &bytes.Buffer{},
				key:            key,
			}

			next.ServeHTTP(hrw, r)

			if hrw.body.Len() > 0 {
				responseHash := hash.CalculateHMAC(hrw.body.Bytes(), key)
				w.Header().Set("HashSHA256", responseHash)
			}
		})
	}
}
