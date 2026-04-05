package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

func TestUpdateMetricsBatch(t *testing.T) {
	tests := []struct {
		name           string
		requestBody    []model.Metrics
		expectedStatus int
		checkResults   bool
	}{
		{
			name: "valid batch with gauge and counter",
			requestBody: []model.Metrics{
				{
					ID:    "gauge1",
					MType: "gauge",
					Value: ptrFloat64(123.45),
				},
				{
					ID:    "gauge2",
					MType: "gauge",
					Value: ptrFloat64(67.89),
				},
				{
					ID:    "counter1",
					MType: "counter",
					Delta: ptrInt64(10),
				},
			},
			expectedStatus: http.StatusOK,
			checkResults:   true,
		},
		{
			name: "batch with multiple gauges",
			requestBody: []model.Metrics{
				{
					ID:    "test1",
					MType: "gauge",
					Value: ptrFloat64(1.1),
				},
				{
					ID:    "test2",
					MType: "gauge",
					Value: ptrFloat64(2.2),
				},
				{
					ID:    "test3",
					MType: "gauge",
					Value: ptrFloat64(3.3),
				},
			},
			expectedStatus: http.StatusOK,
			checkResults:   true,
		},
		{
			name: "batch with multiple counters",
			requestBody: []model.Metrics{
				{
					ID:    "counter1",
					MType: "counter",
					Delta: ptrInt64(5),
				},
				{
					ID:    "counter2",
					MType: "counter",
					Delta: ptrInt64(10),
				},
				{
					ID:    "counter3",
					MType: "counter",
					Delta: ptrInt64(15),
				},
			},
			expectedStatus: http.StatusOK,
			checkResults:   true,
		},
		{
			name:           "empty batch",
			requestBody:    []model.Metrics{},
			expectedStatus: http.StatusBadRequest,
			checkResults:   false,
		},
		{
			name: "batch with gauge missing value",
			requestBody: []model.Metrics{
				{
					ID:    "gauge1",
					MType: "gauge",
				},
			},
			expectedStatus: http.StatusOK,
			checkResults:   false,
		},
		{
			name: "batch with counter missing delta",
			requestBody: []model.Metrics{
				{
					ID:    "counter1",
					MType: "counter",
				},
			},
			expectedStatus: http.StatusOK,
			checkResults:   false,
		},
		{
			name: "batch with negative gauge value",
			requestBody: []model.Metrics{
				{
					ID:    "gauge1",
					MType: "gauge",
					Value: ptrFloat64(-123.45),
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResults:   false,
		},
		{
			name: "batch with negative counter delta",
			requestBody: []model.Metrics{
				{
					ID:    "counter1",
					MType: "counter",
					Delta: ptrInt64(-10),
				},
			},
			expectedStatus: http.StatusBadRequest,
			checkResults:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			metricsService := service.NewMetricsService(store, nil, nil, nil)
			handler, err := NewMetricHandler(metricsService, nil, nil, nil)
			if err != nil {
				t.Fatalf("NewMetricHandler failed: %v", err)
			}
			router := setupRouter(handler)

			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")
			w := httptest.NewRecorder()

			router.ServeHTTP(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d, body: %s", tt.expectedStatus, w.Code, w.Body.String())
			}

			if tt.checkResults && w.Code == http.StatusOK {
				ctx := context.Background()
				for _, metric := range tt.requestBody {
					if metric.MType == "gauge" && metric.Value != nil {
						value, exists := store.GetGauge(ctx, metric.ID)
						if !exists {
							t.Errorf("gauge %s not found in storage", metric.ID)
						}
						if value != *metric.Value {
							t.Errorf("gauge %s: expected value %f, got %f", metric.ID, *metric.Value, value)
						}
					}
					if metric.MType == "counter" && metric.Delta != nil {
						value, exists := store.GetCounter(ctx, metric.ID)
						if !exists {
							t.Errorf("counter %s not found in storage", metric.ID)
						}
						if value != *metric.Delta {
							t.Errorf("counter %s: expected delta %d, got %d", metric.ID, *metric.Delta, value)
						}
					}
				}
			}
		})
	}
}

func TestUpdateMetricsBatch_InvalidJSON(t *testing.T) {
	store := storage.NewMemStorage()
	metricsService := service.NewMetricsService(store, nil, nil, nil)
	handler, err := NewMetricHandler(metricsService, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewMetricHandler failed: %v", err)
	}
	router := setupRouter(handler)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("expected status %d for invalid JSON, got %d", http.StatusBadRequest, w.Code)
	}
}

func TestUpdateMetricsBatch_CounterAccumulation(t *testing.T) {
	store := storage.NewMemStorage()
	metricsService := service.NewMetricsService(store, nil, nil, nil)
	handler, err := NewMetricHandler(metricsService, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewMetricHandler failed: %v", err)
	}
	router := setupRouter(handler)

	batch1 := []model.Metrics{
		{
			ID:    "counter1",
			MType: "counter",
			Delta: ptrInt64(10),
		},
	}
	body1, _ := json.Marshal(batch1)
	req1 := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("first batch failed with status %d", w1.Code)
	}

	batch2 := []model.Metrics{
		{
			ID:    "counter1",
			MType: "counter",
			Delta: ptrInt64(20),
		},
	}
	body2, _ := json.Marshal(batch2)
	req2 := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("second batch failed with status %d", w2.Code)
	}

	ctx := context.Background()
	value, exists := store.GetCounter(ctx, "counter1")
	if !exists {
		t.Fatal("counter1 not found in storage")
	}
	if value != 30 {
		t.Errorf("expected counter1 value 30, got %d", value)
	}
}

func TestUpdateMetricsBatch_GaugeOverwrite(t *testing.T) {
	store := storage.NewMemStorage()
	metricsService := service.NewMetricsService(store, nil, nil, nil)
	handler, err := NewMetricHandler(metricsService, nil, nil, nil)
	if err != nil {
		t.Fatalf("NewMetricHandler failed: %v", err)
	}
	router := setupRouter(handler)

	batch1 := []model.Metrics{
		{
			ID:    "gauge1",
			MType: "gauge",
			Value: ptrFloat64(100.0),
		},
	}
	body1, _ := json.Marshal(batch1)
	req1 := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(body1))
	req1.Header.Set("Content-Type", "application/json")
	w1 := httptest.NewRecorder()
	router.ServeHTTP(w1, req1)

	if w1.Code != http.StatusOK {
		t.Fatalf("first batch failed with status %d", w1.Code)
	}

	batch2 := []model.Metrics{
		{
			ID:    "gauge1",
			MType: "gauge",
			Value: ptrFloat64(200.0),
		},
	}
	body2, _ := json.Marshal(batch2)
	req2 := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewBuffer(body2))
	req2.Header.Set("Content-Type", "application/json")
	w2 := httptest.NewRecorder()
	router.ServeHTTP(w2, req2)

	if w2.Code != http.StatusOK {
		t.Fatalf("second batch failed with status %d", w2.Code)
	}

	ctx := context.Background()
	value, exists := store.GetGauge(ctx, "gauge1")
	if !exists {
		t.Fatal("gauge1 not found in storage")
	}
	if value != 200.0 {
		t.Errorf("expected gauge1 value 200.0, got %f", value)
	}
}
