package agent

import (
	"net/http"
	"net/http/httptest"
	"testing"
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
