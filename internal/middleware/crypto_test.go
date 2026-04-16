package middleware

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/practicum-metrics/internal/encryption"
	"github.com/user/practicum-metrics/internal/hash"
)

func TestCryptoDecrypt(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	plaintext := []byte(`{"id":"Alloc","type":"gauge"}`)

	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	if _, err := gzWriter.Write(plaintext); err != nil {
		t.Fatalf("compress plaintext: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	encryptedBody, err := encryption.Encrypt(compressed.Bytes(), &privateKey.PublicKey)
	if err != nil {
		t.Fatalf("encrypt body: %v", err)
	}

	handler := CryptoDecrypt(privateKey)(GzipDecompress(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(encryption.HeaderEncrypted) != "" {
			t.Fatalf("expected encrypted header to be removed")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		if string(body) != string(plaintext) {
			t.Fatalf("expected %q, got %q", plaintext, body)
		}

		w.WriteHeader(http.StatusOK)
	})))

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(encryptedBody))
	req.Header.Set(encryption.HeaderEncrypted, encryption.HeaderEncryptedValue)
	req.Header.Set("Content-Encoding", "gzip")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}

func TestCryptoDecrypt_WithoutPrivateKey(t *testing.T) {
	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader([]byte(`{}`)))
	req.Header.Set(encryption.HeaderEncrypted, encryption.HeaderEncryptedValue)
	rec := httptest.NewRecorder()

	CryptoDecrypt(nil)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Fatal("handler should not be called")
	})).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
	}
}

func TestCryptoDecrypt_WithHashVerify(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	plaintext := []byte(`{"id":"Alloc","type":"gauge","value":42.5}`)
	requestHash := hash.CalculateHMAC(plaintext, "secret")

	var compressed bytes.Buffer
	gzWriter := gzip.NewWriter(&compressed)
	if _, err := gzWriter.Write(plaintext); err != nil {
		t.Fatalf("compress plaintext: %v", err)
	}
	if err := gzWriter.Close(); err != nil {
		t.Fatalf("close gzip writer: %v", err)
	}

	encryptedBody, err := encryption.Encrypt(compressed.Bytes(), &privateKey.PublicKey)
	if err != nil {
		t.Fatalf("encrypt body: %v", err)
	}

	var handler http.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		if string(body) != string(plaintext) {
			t.Fatalf("expected %q, got %q", plaintext, body)
		}

		w.WriteHeader(http.StatusOK)
	})
	handler = CryptoDecrypt(privateKey)(GzipDecompress(HashVerify("secret")(handler)))

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(encryptedBody))
	req.Header.Set(encryption.HeaderEncrypted, encryption.HeaderEncryptedValue)
	req.Header.Set("Content-Encoding", "gzip")
	req.Header.Set("HashSHA256", requestHash)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}
}
