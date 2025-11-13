package agent

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/practicum-metrics/internal/model"
)

func TestNewMetricsSender(t *testing.T) {
	tests := []struct {
		name          string
		serverAddress string
	}{
		{"create new sender", "http://localhost:8080"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sender := NewMetricsSender(tt.serverAddress)
			if sender == nil {
				t.Fatal("NewMetricsSender returned nil")
			}

			if sender.serverAddress != tt.serverAddress {
				t.Errorf("expected serverAddress to be '%s', got '%s'", tt.serverAddress, sender.serverAddress)
			}

			if sender.client == nil {
				t.Error("resty client not initialized")
			}
		})
	}
}

func TestSendGauge(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		metricValue    float64
		expectedStatus int
		expectedPath   string
		wantError      bool
	}{
		{
			name:           "successful gauge send",
			metricName:     "Alloc",
			metricValue:    123.456,
			expectedStatus: http.StatusOK,
			expectedPath:   "/update/gauge/Alloc/123.456",
			wantError:      false,
		},
		{
			name:           "gauge with zero value",
			metricName:     "TestMetric",
			metricValue:    0.0,
			expectedStatus: http.StatusOK,
			expectedPath:   "/update/gauge/TestMetric/0",
			wantError:      false,
		},
		{
			name:           "server returns non-200 status",
			metricName:     "FailMetric",
			metricValue:    100.0,
			expectedStatus: http.StatusInternalServerError,
			expectedPath:   "/update/gauge/FailMetric/100",
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				contentType := r.Header.Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("expected Content-Type 'text/plain', got '%s'", contentType)
				}

				if r.URL.Path != tt.expectedPath {
					t.Errorf("expected path '%s', got '%s'", tt.expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.expectedStatus)
			}))
			defer server.Close()

			sender := NewMetricsSender(server.URL)
			err := sender.SendGauge(tt.metricName, tt.metricValue)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSendCounter(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		metricValue    int64
		expectedStatus int
		expectedPath   string
		wantError      bool
	}{
		{
			name:           "successful counter send",
			metricName:     "PollCount",
			metricValue:    5,
			expectedStatus: http.StatusOK,
			expectedPath:   "/update/counter/PollCount/5",
			wantError:      false,
		},
		{
			name:           "counter with zero value",
			metricName:     "TestCounter",
			metricValue:    0,
			expectedStatus: http.StatusOK,
			expectedPath:   "/update/counter/TestCounter/0",
			wantError:      false,
		},
		{
			name:           "server returns non-200 status",
			metricName:     "FailCounter",
			metricValue:    50,
			expectedStatus: http.StatusBadRequest,
			expectedPath:   "/update/counter/FailCounter/50",
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				contentType := r.Header.Get("Content-Type")
				if contentType != "text/plain" {
					t.Errorf("expected Content-Type 'text/plain', got '%s'", contentType)
				}

				if r.URL.Path != tt.expectedPath {
					t.Errorf("expected path '%s', got '%s'", tt.expectedPath, r.URL.Path)
				}

				w.WriteHeader(tt.expectedStatus)
			}))
			defer server.Close()

			sender := NewMetricsSender(server.URL)
			err := sender.SendCounter(tt.metricName, tt.metricValue)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSendGaugeJSON(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		metricValue    float64
		expectedStatus int
		wantError      bool
	}{
		{
			name:           "successful gauge send",
			metricName:     "Alloc",
			metricValue:    123.456,
			expectedStatus: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "gauge with zero value",
			metricName:     "TestMetric",
			metricValue:    0.0,
			expectedStatus: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "server returns non-200 status",
			metricName:     "FailMetric",
			metricValue:    100.0,
			expectedStatus: http.StatusInternalServerError,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				if r.URL.Path != "/update" {
					t.Errorf("expected path '/update', got '%s'", r.URL.Path)
				}

				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("failed to read body: %v", err)
				}

				var metric model.Metrics
				if err := json.Unmarshal(body, &metric); err != nil {
					t.Fatalf("failed to unmarshal JSON: %v", err)
				}

				if metric.ID != tt.metricName {
					t.Errorf("expected ID '%s', got '%s'", tt.metricName, metric.ID)
				}

				if metric.MType != "gauge" {
					t.Errorf("expected MType 'gauge', got '%s'", metric.MType)
				}

				if metric.Value == nil {
					t.Error("expected Value to be set")
				} else if *metric.Value != tt.metricValue {
					t.Errorf("expected Value %f, got %f", tt.metricValue, *metric.Value)
				}

				w.WriteHeader(tt.expectedStatus)
			}))
			defer server.Close()

			sender := NewMetricsSender(server.URL)
			err := sender.SendGaugeJSON(tt.metricName, tt.metricValue)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

func TestSendCounterJSON(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		metricValue    int64
		expectedStatus int
		wantError      bool
	}{
		{
			name:           "successful counter send",
			metricName:     "PollCount",
			metricValue:    5,
			expectedStatus: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "counter with zero value",
			metricName:     "TestCounter",
			metricValue:    0,
			expectedStatus: http.StatusOK,
			wantError:      false,
		},
		{
			name:           "server returns non-200 status",
			metricName:     "FailCounter",
			metricValue:    50,
			expectedStatus: http.StatusBadRequest,
			wantError:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("expected POST method, got %s", r.Method)
				}

				if r.URL.Path != "/update" {
					t.Errorf("expected path '/update', got '%s'", r.URL.Path)
				}

				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
				}

				body, err := io.ReadAll(r.Body)
				if err != nil {
					t.Fatalf("failed to read body: %v", err)
				}

				var metric model.Metrics
				if err := json.Unmarshal(body, &metric); err != nil {
					t.Fatalf("failed to unmarshal JSON: %v", err)
				}

				if metric.ID != tt.metricName {
					t.Errorf("expected ID '%s', got '%s'", tt.metricName, metric.ID)
				}

				if metric.MType != "counter" {
					t.Errorf("expected MType 'counter', got '%s'", metric.MType)
				}

				if metric.Delta == nil {
					t.Error("expected Delta to be set")
				} else if *metric.Delta != tt.metricValue {
					t.Errorf("expected Delta %d, got %d", tt.metricValue, *metric.Delta)
				}

				w.WriteHeader(tt.expectedStatus)
			}))
			defer server.Close()

			sender := NewMetricsSender(server.URL)
			err := sender.SendCounterJSON(tt.metricName, tt.metricValue)

			if tt.wantError && err == nil {
				t.Error("expected error, got nil")
			}

			if !tt.wantError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
