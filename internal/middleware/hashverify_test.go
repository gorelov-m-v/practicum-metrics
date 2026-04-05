package middleware

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/practicum-metrics/internal/hash"
)

func TestHashVerify_WithKey(t *testing.T) {
	key := "test-secret-key"
	body := []byte(`{"test":"data"}`)
	validHash := hash.CalculateHMAC(body, key)

	tests := []struct {
		name           string
		requestBody    []byte
		requestHash    string
		expectedStatus int
	}{
		{
			name:           "valid hash",
			requestBody:    body,
			requestHash:    validHash,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid hash",
			requestBody:    body,
			requestHash:    "invalid-hash",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "no hash header",
			requestBody:    body,
			requestHash:    "",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "wrong body with valid format hash",
			requestBody:    []byte(`{"test":"wrong"}`),
			requestHash:    validHash,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
				w.Write([]byte("OK"))
			})

			middleware := HashVerify(key)
			wrappedHandler := middleware(handler)

			req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(tt.requestBody))
			if tt.requestHash != "" {
				req.Header.Set("HashSHA256", tt.requestHash)
			}

			rec := httptest.NewRecorder()
			wrappedHandler.ServeHTTP(rec, req)

			if rec.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
			}

			if rec.Code == http.StatusOK {
				responseHash := rec.Header().Get("HashSHA256")
				if responseHash == "" {
					t.Error("Expected HashSHA256 header in response, got empty")
				}
			}
		})
	}
}

func TestHashVerify_WithoutKey(t *testing.T) {
	body := []byte(`{"test":"data"}`)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := HashVerify("")
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("HashSHA256", "some-hash")

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	responseHash := rec.Header().Get("HashSHA256")
	if responseHash != "" {
		t.Error("Expected no HashSHA256 header in response without key")
	}
}

func TestHashVerify_ResponseHash(t *testing.T) {
	key := "test-secret-key"
	requestBody := []byte(`{"test":"data"}`)
	responseBody := []byte(`{"result":"success"}`)
	validHash := hash.CalculateHMAC(requestBody, key)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(responseBody)
	})

	middleware := HashVerify(key)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(requestBody))
	req.Header.Set("HashSHA256", validHash)

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	responseHash := rec.Header().Get("HashSHA256")
	expectedResponseHash := hash.CalculateHMAC(responseBody, key)

	if responseHash != expectedResponseHash {
		t.Errorf("Expected response hash %s, got %s", expectedResponseHash, responseHash)
	}
}

func TestHashVerify_GetRequest(t *testing.T) {
	key := "test-secret-key"

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := HashVerify(key)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodGet, "/test", nil)

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestHashVerify_RequestBodyPreserved(t *testing.T) {
	key := "test-secret-key"
	body := []byte(`{"test":"data"}`)
	validHash := hash.CalculateHMAC(body, key)

	var receivedBody []byte
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		receivedBody, err = io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("Failed to read body in handler: %v", err)
		}
		w.WriteHeader(http.StatusOK)
	})

	middleware := HashVerify(key)
	wrappedHandler := middleware(handler)

	req := httptest.NewRequest(http.MethodPost, "/test", bytes.NewBuffer(body))
	req.Header.Set("HashSHA256", validHash)

	rec := httptest.NewRecorder()
	wrappedHandler.ServeHTTP(rec, req)

	if !bytes.Equal(receivedBody, body) {
		t.Errorf("Request body not preserved. Expected %s, got %s", string(body), string(receivedBody))
	}
}
