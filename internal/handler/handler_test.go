package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

func setupRouter(h *MetricHandler) *chi.Mux {
	r := chi.NewRouter()
	r.Post("/updates", h.UpdateMetricsBatch)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	r.Post("/update/", h.UpdateMetricJSON)
	r.Post("/update", h.UpdateMetricJSON)
	r.Get("/value/{type}/{name}", h.GetMetric)
	r.Post("/value/", h.GetMetricJSON)
	r.Post("/value", h.GetMetricJSON)
	r.Get("/", h.ListMetrics)
	return r
}

func TestNewMetricHandler(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"create new handler"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}

			if handler == nil {
				t.Fatal("NewMetricHandler returned nil")
			}

			if handler.service == nil {
				t.Error("service not set in handler")
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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkValue && w.Code == http.StatusOK {
				ctx := context.Background()
				value, exists := store.GetGauge(ctx, tt.metricName)

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
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req := httptest.NewRequest(tt.method, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkValue && w.Code == http.StatusOK {
				ctx := context.Background()
				value, exists := store.GetCounter(ctx, tt.metricName)

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
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			w := httptest.NewRecorder()

			handler.UpdateMetric(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
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
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req := httptest.NewRequest(http.MethodPost, tt.path, nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

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
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			for i := 0; i < tt.updates; i++ {
				path := fmt.Sprintf("/update/counter/%s/%d", tt.metricName, tt.valuePerUpdate)
				req := httptest.NewRequest(http.MethodPost, path, nil)
				w := httptest.NewRecorder()
				router.ServeHTTP(w, req)

				if w.Code != http.StatusOK {
					t.Errorf("request %d failed with status %d", i+1, w.Code)
				}
			}

			ctx := context.Background()
			value, exists := store.GetCounter(ctx, tt.metricName)
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
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req1 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/gauge/%s/%v", tt.metricName, tt.firstValue), nil)
			w1 := httptest.NewRecorder()
			router.ServeHTTP(w1, req1)

			req2 := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/update/gauge/%s/%v", tt.metricName, tt.secondValue), nil)
			w2 := httptest.NewRecorder()
			router.ServeHTTP(w2, req2)

			ctx := context.Background()
			value, exists := store.GetGauge(ctx, tt.metricName)
			if !exists {
				t.Errorf("%s was not stored", tt.metricName)
			}

			if value != tt.expected {
				t.Errorf("expected gauge value %v, got %f", tt.expected, value)
			}
		})
	}
}

func TestGetMetric_Gauge(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		metricValue    float64
		setupMetric    bool
		expectedStatus int
		expectedBody   string
	}{
		{"existing gauge", "Alloc", 123.456, true, http.StatusOK, "123.456"},
		{"existing gauge with zero", "TestMetric", 0.0, true, http.StatusOK, "0"},
		{"non-existent gauge", "Unknown", 0.0, false, http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := storage.NewMemStorage()
			if tt.setupMetric {
				store.UpdateGauge(ctx, tt.metricName, tt.metricValue)
			}

			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/gauge/%s", tt.metricName), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				if w.Body.String() != tt.expectedBody {
					t.Errorf("expected body '%s', got '%s'", tt.expectedBody, w.Body.String())
				}
			}
		})
	}
}

func TestGetMetric_Counter(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		metricValue    int64
		setupMetric    bool
		expectedStatus int
		expectedBody   string
	}{
		{"existing counter", "PollCount", 42, true, http.StatusOK, "42"},
		{"existing counter with zero", "TestCounter", 0, true, http.StatusOK, "0"},
		{"non-existent counter", "Unknown", 0, false, http.StatusNotFound, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := storage.NewMemStorage()
			if tt.setupMetric {
				store.UpdateCounter(ctx, tt.metricName, tt.metricValue)
			}

			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/counter/%s", tt.metricName), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.expectedStatus == http.StatusOK {
				if w.Body.String() != tt.expectedBody {
					t.Errorf("expected body '%s', got '%s'", tt.expectedBody, w.Body.String())
				}
			}
		})
	}
}

func TestGetMetric_InvalidType(t *testing.T) {
	tests := []struct {
		name           string
		metricType     string
		metricName     string
		expectedStatus int
	}{
		{"invalid type", "invalid", "TestMetric", http.StatusBadRequest},
		{"unknown type", "unknown", "TestMetric", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/value/%s/%s", tt.metricType, tt.metricName), nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestListMetrics(t *testing.T) {
	tests := []struct {
		name           string
		gauges         map[string]float64
		counters       map[string]int64
		expectedStatus int
		checkContent   []string
	}{
		{
			name:           "with metrics",
			gauges:         map[string]float64{"Alloc": 123.456, "HeapAlloc": 200.0},
			counters:       map[string]int64{"PollCount": 5, "Requests": 10},
			expectedStatus: http.StatusOK,
			checkContent:   []string{"Alloc", "123.456", "HeapAlloc", "200", "PollCount", "5", "Requests", "10"},
		},
		{
			name:           "empty storage",
			gauges:         map[string]float64{},
			counters:       map[string]int64{},
			expectedStatus: http.StatusOK,
			checkContent:   []string{"Metrics"},
		},
		{
			name:           "only gauges",
			gauges:         map[string]float64{"Alloc": 100.0},
			counters:       map[string]int64{},
			expectedStatus: http.StatusOK,
			checkContent:   []string{"Alloc", "100"},
		},
		{
			name:           "only counters",
			gauges:         map[string]float64{},
			counters:       map[string]int64{"PollCount": 3},
			expectedStatus: http.StatusOK,
			checkContent:   []string{"PollCount", "3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := storage.NewMemStorage()

			for name, value := range tt.gauges {
				store.UpdateGauge(ctx, name, value)
			}

			for name, value := range tt.counters {
				store.UpdateCounter(ctx, name, value)
			}

			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req := httptest.NewRequest(http.MethodGet, "/", nil)
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			contentType := w.Header().Get("Content-Type")
			if contentType != "text/html; charset=utf-8" {
				t.Errorf("expected Content-Type 'text/html; charset=utf-8', got '%s'", contentType)
			}

			body := w.Body.String()
			for _, content := range tt.checkContent {
				if !strings.Contains(body, content) {
					t.Errorf("expected body to contain '%s'", content)
				}
			}
		})
	}
}

func TestUpdateMetricJSON_Gauge(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    model.Metrics
		expectedStatus int
		checkValue     bool
		expectedValue  float64
	}{
		{
			name: "valid gauge metric",
			requestBody: model.Metrics{
				ID:    "Alloc",
				MType: "gauge",
				Value: ptrFloat64(123.456),
			},
			expectedStatus: http.StatusOK,
			checkValue:     true,
			expectedValue:  123.456,
		},
		{
			name: "gauge with zero value",
			requestBody: model.Metrics{
				ID:    "TestMetric",
				MType: "gauge",
				Value: ptrFloat64(0.0),
			},
			expectedStatus: http.StatusOK,
			checkValue:     true,
			expectedValue:  0.0,
		},
		{
			name: "gauge missing value",
			requestBody: model.Metrics{
				ID:    "InvalidMetric",
				MType: "gauge",
			},
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
		{
			name: "invalid metric type",
			requestBody: model.Metrics{
				ID:    "TestMetric",
				MType: "invalid",
				Value: ptrFloat64(100.0),
			},
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkValue && w.Code == http.StatusOK {
				ctx := context.Background()
				value, exists := store.GetGauge(ctx, tt.requestBody.ID)
				if !exists {
					t.Errorf("metric %s was not stored", tt.requestBody.ID)
				} else if value != tt.expectedValue {
					t.Errorf("expected value %f, got %f", tt.expectedValue, value)
				}

				var resp model.Metrics
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.ID != tt.requestBody.ID {
					t.Errorf("expected ID '%s', got '%s'", tt.requestBody.ID, resp.ID)
				}

				if resp.MType != "gauge" {
					t.Errorf("expected MType 'gauge', got '%s'", resp.MType)
				}

				if resp.Value == nil {
					t.Error("expected Value to be set in response")
				} else if *resp.Value != tt.expectedValue {
					t.Errorf("expected Value %f in response, got %f", tt.expectedValue, *resp.Value)
				}
			}

			if w.Code == http.StatusOK {
				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
				}
			}
		})
	}
}

func TestUpdateMetricJSON_Counter(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    model.Metrics
		expectedStatus int
		checkValue     bool
		expectedValue  int64
	}{
		{
			name: "valid counter metric",
			requestBody: model.Metrics{
				ID:    "PollCount",
				MType: "counter",
				Delta: ptrInt64(5),
			},
			expectedStatus: http.StatusOK,
			checkValue:     true,
			expectedValue:  5,
		},
		{
			name: "counter with zero delta",
			requestBody: model.Metrics{
				ID:    "TestCounter",
				MType: "counter",
				Delta: ptrInt64(0),
			},
			expectedStatus: http.StatusOK,
			checkValue:     true,
			expectedValue:  0,
		},
		{
			name: "counter missing delta",
			requestBody: model.Metrics{
				ID:    "InvalidCounter",
				MType: "counter",
			},
			expectedStatus: http.StatusBadRequest,
			checkValue:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkValue && w.Code == http.StatusOK {
				ctx := context.Background()
				value, exists := store.GetCounter(ctx, tt.requestBody.ID)
				if !exists {
					t.Errorf("metric %s was not stored", tt.requestBody.ID)
				} else if value != tt.expectedValue {
					t.Errorf("expected value %d, got %d", tt.expectedValue, value)
				}

				var resp model.Metrics
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.ID != tt.requestBody.ID {
					t.Errorf("expected ID '%s', got '%s'", tt.requestBody.ID, resp.ID)
				}

				if resp.MType != "counter" {
					t.Errorf("expected MType 'counter', got '%s'", resp.MType)
				}

				if resp.Delta == nil {
					t.Error("expected Delta to be set in response")
				} else if *resp.Delta != tt.expectedValue {
					t.Errorf("expected Delta %d in response, got %d", tt.expectedValue, *resp.Delta)
				}
			}
		})
	}
}

func TestUpdateMetricJSON_InvalidJSON(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "invalid JSON",
			requestBody:    "{invalid json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req := httptest.NewRequest(http.MethodPost, "/update", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestGetMetricJSON_Gauge(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		metricValue    float64
		setupMetric    bool
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "existing gauge",
			metricName:     "Alloc",
			metricValue:    123.456,
			setupMetric:    true,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "non-existent gauge",
			metricName:     "Unknown",
			setupMetric:    false,
			expectedStatus: http.StatusNotFound,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := storage.NewMemStorage()
			if tt.setupMetric {
				store.UpdateGauge(ctx, tt.metricName, tt.metricValue)
			}

			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			reqBody := model.Metrics{
				ID:    tt.metricName,
				MType: "gauge",
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse && w.Code == http.StatusOK {
				var resp model.Metrics
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.ID != tt.metricName {
					t.Errorf("expected ID '%s', got '%s'", tt.metricName, resp.ID)
				}

				if resp.MType != "gauge" {
					t.Errorf("expected MType 'gauge', got '%s'", resp.MType)
				}

				if resp.Value == nil {
					t.Error("expected Value to be set")
				} else if *resp.Value != tt.metricValue {
					t.Errorf("expected Value %f, got %f", tt.metricValue, *resp.Value)
				}

				contentType := w.Header().Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
				}
			}
		})
	}
}

func TestGetMetricJSON_Counter(t *testing.T) {
	tests := []struct {
		name           string
		metricName     string
		metricValue    int64
		setupMetric    bool
		expectedStatus int
		checkResponse  bool
	}{
		{
			name:           "existing counter",
			metricName:     "PollCount",
			metricValue:    42,
			setupMetric:    true,
			expectedStatus: http.StatusOK,
			checkResponse:  true,
		},
		{
			name:           "non-existent counter",
			metricName:     "Unknown",
			setupMetric:    false,
			expectedStatus: http.StatusNotFound,
			checkResponse:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			store := storage.NewMemStorage()
			if tt.setupMetric {
				store.UpdateCounter(ctx, tt.metricName, tt.metricValue)
			}

			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			reqBody := model.Metrics{
				ID:    tt.metricName,
				MType: "counter",
			}
			body, _ := json.Marshal(reqBody)
			req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkResponse && w.Code == http.StatusOK {
				var resp model.Metrics
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.ID != tt.metricName {
					t.Errorf("expected ID '%s', got '%s'", tt.metricName, resp.ID)
				}

				if resp.MType != "counter" {
					t.Errorf("expected MType 'counter', got '%s'", resp.MType)
				}

				if resp.Delta == nil {
					t.Error("expected Delta to be set")
				} else if *resp.Delta != tt.metricValue {
					t.Errorf("expected Delta %d, got %d", tt.metricValue, *resp.Delta)
				}
			}
		})
	}
}

func TestGetMetricJSON_InvalidJSON(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "invalid JSON",
			requestBody:    "{invalid json}",
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "empty body",
			requestBody:    "",
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			req := httptest.NewRequest(http.MethodPost, "/value", strings.NewReader(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func ptrFloat64(f float64) *float64 {
	return &f
}

func ptrInt64(i int64) *int64 {
	return &i
}

func TestUpdateMetricJSON_WithTrailingSlash(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		requestBody    model.Metrics
		expectedStatus int
	}{
		{
			name: "POST /update/ with gauge",
			path: "/update/",
			requestBody: model.Metrics{
				ID:    "TestGauge",
				MType: "gauge",
				Value: ptrFloat64(100.5),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "POST /update with gauge",
			path: "/update",
			requestBody: model.Metrics{
				ID:    "TestGauge2",
				MType: "gauge",
				Value: ptrFloat64(200.5),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "POST /update/ with counter",
			path: "/update/",
			requestBody: model.Metrics{
				ID:    "TestCounter",
				MType: "counter",
				Delta: ptrInt64(42),
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "POST /update with counter",
			path: "/update",
			requestBody: model.Metrics{
				ID:    "TestCounter2",
				MType: "counter",
				Delta: ptrInt64(99),
			},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if w.Code == http.StatusOK {
				var resp model.Metrics
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.ID != tt.requestBody.ID {
					t.Errorf("expected ID '%s', got '%s'", tt.requestBody.ID, resp.ID)
				}

				if resp.MType != tt.requestBody.MType {
					t.Errorf("expected MType '%s', got '%s'", tt.requestBody.MType, resp.MType)
				}
			}
		})
	}
}

func TestGetMetricJSON_WithTrailingSlash(t *testing.T) {
	tests := []struct {
		name           string
		path           string
		setupMetric    func(*storage.MemStorage)
		requestBody    model.Metrics
		expectedStatus int
		checkValue     bool
	}{
		{
			name: "POST /value/ for gauge",
			path: "/value/",
			setupMetric: func(s *storage.MemStorage) {
				ctx := context.Background()
				s.UpdateGauge(ctx, "TestGauge", 123.456)
			},
			requestBody: model.Metrics{
				ID:    "TestGauge",
				MType: "gauge",
			},
			expectedStatus: http.StatusOK,
			checkValue:     true,
		},
		{
			name: "POST /value for gauge",
			path: "/value",
			setupMetric: func(s *storage.MemStorage) {
				ctx := context.Background()
				s.UpdateGauge(ctx, "TestGauge2", 456.789)
			},
			requestBody: model.Metrics{
				ID:    "TestGauge2",
				MType: "gauge",
			},
			expectedStatus: http.StatusOK,
			checkValue:     true,
		},
		{
			name: "POST /value/ for counter",
			path: "/value/",
			setupMetric: func(s *storage.MemStorage) {
				ctx := context.Background()
				s.UpdateCounter(ctx, "TestCounter", 100)
			},
			requestBody: model.Metrics{
				ID:    "TestCounter",
				MType: "counter",
			},
			expectedStatus: http.StatusOK,
			checkValue:     true,
		},
		{
			name: "POST /value for counter",
			path: "/value",
			setupMetric: func(s *storage.MemStorage) {
				ctx := context.Background()
				s.UpdateCounter(ctx, "TestCounter2", 200)
			},
			requestBody: model.Metrics{
				ID:    "TestCounter2",
				MType: "counter",
			},
			expectedStatus: http.StatusOK,
			checkValue:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			tt.setupMetric(store)

			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, tt.path, bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, w.Code)
			}

			if tt.checkValue && w.Code == http.StatusOK {
				var resp model.Metrics
				if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}

				if resp.ID != tt.requestBody.ID {
					t.Errorf("expected ID '%s', got '%s'", tt.requestBody.ID, resp.ID)
				}

				if resp.MType != tt.requestBody.MType {
					t.Errorf("expected MType '%s', got '%s'", tt.requestBody.MType, resp.MType)
				}

				if tt.requestBody.MType == "gauge" && resp.Value == nil {
					t.Error("expected Value to be set for gauge metric")
				}

				if tt.requestBody.MType == "counter" && resp.Delta == nil {
					t.Error("expected Delta to be set for counter metric")
				}
			}
		})
	}
}
