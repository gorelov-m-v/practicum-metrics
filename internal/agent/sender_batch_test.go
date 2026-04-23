package agent

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/practicum-metrics/internal/model"
)

func TestSendMetricsBatch_Success(t *testing.T) {
	var receivedMetrics []model.Metrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/updates" {
			t.Errorf("expected path '/updates', got '%s'", r.URL.Path)
		}

		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type 'application/json', got '%s'", r.Header.Get("Content-Type"))
		}

		if r.Header.Get("Content-Encoding") != "gzip" {
			t.Errorf("expected Content-Encoding 'gzip', got '%s'", r.Header.Get("Content-Encoding"))
		}
		if ip := net.ParseIP(r.Header.Get(headerXRealIP)); ip == nil {
			t.Errorf("expected valid %s header, got %q", headerXRealIP, r.Header.Get(headerXRealIP))
		}

		gzReader, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()

		body, err := io.ReadAll(gzReader)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		if err := json.Unmarshal(body, &receivedMetrics); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")

	gaugeValue := 123.45
	counterDelta := int64(10)

	metrics := []model.Metrics{
		{
			ID:    "test_gauge",
			MType: "gauge",
			Value: &gaugeValue,
		},
		{
			ID:    "test_counter",
			MType: "counter",
			Delta: &counterDelta,
		},
	}

	err := sender.SendMetricsBatch(metrics)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(receivedMetrics) != 2 {
		t.Errorf("expected 2 metrics, got %d", len(receivedMetrics))
	}

	if receivedMetrics[0].ID != "test_gauge" {
		t.Errorf("expected metric ID 'test_gauge', got '%s'", receivedMetrics[0].ID)
	}

	if receivedMetrics[1].ID != "test_counter" {
		t.Errorf("expected metric ID 'test_counter', got '%s'", receivedMetrics[1].ID)
	}
}

func TestSendMetricsBatch_EmptyBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not receive request for empty batch")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")
	err := sender.SendMetricsBatch([]model.Metrics{})

	if err != nil {
		t.Errorf("expected no error for empty batch, got: %v", err)
	}
}

func TestSendMetricsBatch_NilBatch(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("server should not receive request for nil batch")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")
	err := sender.SendMetricsBatch(nil)

	if err != nil {
		t.Errorf("expected no error for nil batch, got: %v", err)
	}
}

func TestSendMetricsBatch_ServerError(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")

	gaugeValue := 123.45
	metrics := []model.Metrics{
		{
			ID:    "test_gauge",
			MType: "gauge",
			Value: &gaugeValue,
		},
	}

	err := sender.SendMetricsBatch(metrics)
	if err == nil {
		t.Error("expected error after all retries failed")
	}

	if attemptCount != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount)
	}
}

func TestSendMetricsBatch_WithRetries(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
		} else {
			w.WriteHeader(http.StatusOK)
		}
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")

	gaugeValue := 123.45
	metrics := []model.Metrics{
		{
			ID:    "test_gauge",
			MType: "gauge",
			Value: &gaugeValue,
		},
	}

	err := sender.SendMetricsBatch(metrics)
	if err != nil {
		t.Errorf("expected success after retries, got error: %v", err)
	}

	if attemptCount != 3 {
		t.Errorf("expected 3 attempts, got %d", attemptCount)
	}
}

func TestSendMetricsBatch_LargeBatch(t *testing.T) {
	var receivedMetrics []model.Metrics

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gzReader, err := gzip.NewReader(r.Body)
		if err != nil {
			t.Fatalf("failed to create gzip reader: %v", err)
		}
		defer gzReader.Close()

		body, err := io.ReadAll(gzReader)
		if err != nil {
			t.Fatalf("failed to read body: %v", err)
		}

		if err := json.Unmarshal(body, &receivedMetrics); err != nil {
			t.Fatalf("failed to unmarshal JSON: %v", err)
		}

		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	sender := NewMetricsSender(server.URL, "")

	metrics := make([]model.Metrics, 100)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			value := float64(i)
			metrics[i] = model.Metrics{
				ID:    "gauge_" + string(rune(i)),
				MType: "gauge",
				Value: &value,
			}
		} else {
			delta := int64(i)
			metrics[i] = model.Metrics{
				ID:    "counter_" + string(rune(i)),
				MType: "counter",
				Delta: &delta,
			}
		}
	}

	err := sender.SendMetricsBatch(metrics)
	if err != nil {
		t.Errorf("expected no error, got: %v", err)
	}

	if len(receivedMetrics) != 100 {
		t.Errorf("expected 100 metrics, got %d", len(receivedMetrics))
	}
}
