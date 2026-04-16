package agent

import (
	"bytes"
	"compress/gzip"
	"crypto/rand"
	"crypto/rsa"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/practicum-metrics/internal/encryption"
	"github.com/user/practicum-metrics/internal/model"
)

func TestSendMetricsBatch_Encrypted(t *testing.T) {
	privateKey, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatalf("generate private key: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get(encryption.HeaderEncrypted) != encryption.HeaderEncryptedValue {
			t.Fatalf("expected encrypted header to be set")
		}

		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Fatalf("read request body: %v", err)
		}

		decrypted, err := encryption.Decrypt(body, privateKey)
		if err != nil {
			t.Fatalf("decrypt body: %v", err)
		}

		gzReader, err := gzip.NewReader(bytes.NewReader(decrypted))
		if err != nil {
			t.Fatalf("create gzip reader: %v", err)
		}
		defer gzReader.Close()

		payload, err := io.ReadAll(gzReader)
		if err != nil {
			t.Fatalf("read decrypted payload: %v", err)
		}

		var metrics []model.Metrics
		if err := json.Unmarshal(payload, &metrics); err != nil {
			t.Fatalf("unmarshal metrics: %v", err)
		}

		if len(metrics) != 1 || metrics[0].ID != "Alloc" {
			t.Fatalf("unexpected metrics payload: %+v", metrics)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")
	sender.SetPublicKey(&privateKey.PublicKey)

	value := 42.5
	err = sender.SendMetricsBatch([]model.Metrics{{
		ID:    "Alloc",
		MType: "gauge",
		Value: &value,
	}})
	if err != nil {
		t.Fatalf("send metrics batch: %v", err)
	}
}
