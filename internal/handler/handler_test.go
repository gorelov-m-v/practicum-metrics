package handler

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/practicum-metrics/internal/storage"
)

func TestNewMetricHandler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"create new handler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			handler := NewMetricHandler(store)

			if handler == nil {
				t.Fatal("NewMetricHandler returned nil")
			}

			if handler.storage == nil {
				t.Error("storage not set in handler")
			}
		})
	}
}

func TestUpdateMetric_Gauge(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		metricName     string
		expectedValue  float64
		checkValue     bool
	}{
		{
			name:           "valid gauge metric",
			method:         http.MethodPost,
			path:           "/update/gauge/Alloc/123.456",
			expectedStatus: http.StatusOK,
			metricName:     "Alloc",
			expectedValue:  123.456,
			checkValue:     true,
		},
		{
			name:           "gauge with zero value",
			method:         http.MethodPost,
			path:           "/update/gauge/TestMetric/0",
			expectedStatus: http.StatusOK,
			metricName:     "TestMetric",
			expectedValue:  0.0,
			checkValue:     true,
		},
		{
			name:           "gauge with negative value",
			method:         http.MethodPost,
			path:           "/update/gauge/NegativeMetric/-100.5",
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
		{
			name:           "invalid gauge value - not a number",
			method:         http.MethodPost,
			path:           "/update/gauge/InvalidMetric/abc",
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
		{
			name:           "missing gauge value",
			method:         http.MethodPost,
			path:           "/update/gauge/MissingValue/",
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
		{
			name:           "missing gauge name",
			method:         http.MethodPost,
			path:           "/update/gauge//123.456",
			expectedStatus: http.StatusNotFound,
			checkValue:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			handler := NewMetricHandler(store)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.UpdateMetric(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkValue && w.Code == http.StatusOK {
				value, exists := store.GetGauge(tt.metricName)

				if !exists {
					t.Errorf("metric %s was not stored", tt.metricName)
				} else if value != tt.expectedValue {
					t.Errorf("expected value %f, got %f", tt.expectedValue, value)
				}
			}
		})
	}
}

func TestUpdateMetric_Counter(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		path           string
		expectedStatus int
		metricName     string
		expectedValue  int64
		checkValue     bool
	}{
		{
			name:           "valid counter metric",
			method:         http.MethodPost,
			path:           "/update/counter/PollCount/5",
			expectedStatus: http.StatusOK,
			metricName:     "PollCount",
			expectedValue:  5,
			checkValue:     true,
		},
		{
			name:           "counter with zero value",
			method:         http.MethodPost,
			path:           "/update/counter/ZeroCounter/0",
			expectedStatus: http.StatusOK,
			metricName:     "ZeroCounter",
			expectedValue:  0,
			checkValue:     true,
		},
		{
			name:           "counter with negative value",
			method:         http.MethodPost,
			path:           "/update/counter/NegativeCounter/-10",
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
		{
			name:           "invalid counter value - float",
			method:         http.MethodPost,
			path:           "/update/counter/InvalidCounter/123.456",
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
		{
			name:           "invalid counter value - not a number",
			method:         http.MethodPost,
			path:           "/update/counter/InvalidCounter/abc",
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
		{
			name:           "missing counter value",
			method:         http.MethodPost,
			path:           "/update/counter/MissingValue/",
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
		{
			name:           "missing counter name",
			method:         http.MethodPost,
			path:           "/update/counter//5",
			expectedStatus: http.StatusNotFound,
			checkValue:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			handler := NewMetricHandler(store)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			handler.UpdateMetric(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkValue && w.Code == http.StatusOK {
				value, exists := store.GetCounter(tt.metricName)

				if !exists {
					t.Errorf("metric %s was not stored", tt.metricName)
				} else if value != tt.expectedValue {
					t.Errorf("expected value %d, got %d", tt.expectedValue, value)
				}
			}
		})
	}
}

func TestUpdateMetric_InvalidType(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		expectedStatus int
	}{
		{
			name:           "invalid metric type",
			path:           "/update/invalid/TestMetric/123",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "unknown metric type",
			path:           "/update/unknown/TestMetric/456",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			handler := NewMetricHandler(store)

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			w := httptest.NewRecorder()

			handler.UpdateMetric(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestUpdateMetric_MethodNotAllowed(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		expectedStatus int
	}{
		{"GET not allowed", http.MethodGet, http.StatusMethodNotAllowed},
		{"PUT not allowed", http.MethodPut, http.StatusMethodNotAllowed},
		{"DELETE not allowed", http.MethodDelete, http.StatusMethodNotAllowed},
		{"PATCH not allowed", http.MethodPatch, http.StatusMethodNotAllowed},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			handler := NewMetricHandler(store)

			req := httptest.NewRequest(tt.method, "/update/gauge/TestMetric/123.456", nil)
			w := httptest.NewRecorder()

			handler.UpdateMetric(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d for method %s, got %d", tt.expectedStatus, tt.method, w.Code)
			}
		})
	}
}

func TestUpdateMetric_ContentType(t *testing.T) {
	tests := []struct {
		name                string
		path                string
		expectedContentType string
	}{
		{
			name:                "correct content type",
			path:                "/update/gauge/TestMetric/123.456",
			expectedContentType: "text/plain; charset=utf-8",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			handler := NewMetricHandler(store)

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			w := httptest.NewRecorder()

			handler.UpdateMetric(w, req)

			if w.Code == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				if contentType != tt.expectedContentType {
					t.Errorf("expected Content-Type '%s', got '%s'", tt.expectedContentType, contentType)
				}
			}
		})
	}
}

func TestUpdateMetric_CounterAccumulation(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		updates        int
		valuePerUpdate int64
		expectedTotal  int64
	}{
		{"five updates of 1", "TestCounter", 5, 1, 5},
		{"three updates of 2", "Counter2", 3, 2, 6},
		{"ten updates of 1", "Counter3", 10, 1, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			handler := NewMetricHandler(store)

			for i := 0; i < tt.updates; i++ {
				path := fmt.Sprintf("/update/counter/%s/%d", tt.metricName, tt.valuePerUpdate)
				req := httptest.NewRequest(http.MethodPost, path, nil)
				w := httptest.NewRecorder()
				handler.UpdateMetric(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("request %d failed with status %d", i+1, w.Code)
				}
			}

			value, exists := store.GetCounter(tt.metricName)
			if !exists {
				t.Errorf("%s was not stored", tt.metricName)
			}

			if value != tt.expectedTotal {
				t.Errorf("expected counter value %d, got %d", tt.expectedTotal, value)
			}
		})
	}
}

func TestUpdateMetric_GaugeOverwrite(t *testing.T) {
	tests := []struct {
		name        string
		metricName  string
		firstValue  float64
		secondValue float64
		expected    float64
	}{
		{"overwrite 100 with 200", "TestGauge", 100.0, 200.0, 200.0},
		{"overwrite 500 with 50", "Gauge2", 500.0, 50.0, 50.0},
		{"overwrite with zero", "Gauge3", 100.0, 0.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			handler := NewMetricHandler(store)

			req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/gauge/%s/%v", tt.metricName, tt.firstValue), nil)
			w1 := httptest.NewRecorder()
			handler.UpdateMetric(w1, req1)

			req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/gauge/%s/%v", tt.metricName, tt.secondValue), nil)
			w2 := httptest.NewRecorder()
			handler.UpdateMetric(w2, req2)

			value, exists := store.GetGauge(tt.metricName)
			if !exists {
				t.Errorf("%s was not stored", tt.metricName)
			}

			if value != tt.expected {
				t.Errorf("expected gauge value %v, got %f", tt.expected, value)
			}
		})
	}
}
