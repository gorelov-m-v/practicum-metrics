package handler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"

	"github.com/go-chi/chi/v5"

	"github.com/user/practicum-metrics/internal/handler"
	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

func newTestRouter() (*chi.Mux, *service.MetricsService) {
	store := storage.NewMemStorage()
	svc := service.NewMetricsService(store, nil, nil, nil)
	h, _ := handler.NewMetricHandler(svc, nil, nil, nil)
	r := chi.NewRouter()
	r.Post("/updates", h.UpdateMetricsBatch)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	r.Post("/update", h.UpdateMetricJSON)
	r.Get("/value/{type}/{name}", h.GetMetric)
	r.Post("/value", h.GetMetricJSON)
	r.Get("/", h.ListMetrics)
	return r, svc
}

func ExampleMetricHandler_UpdateMetricJSON() {
	router, _ := newTestRouter()

	val := 42.5
	metric := model.Metrics{ID: "Temperature", MType: "gauge", Value: &val}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp model.Metrics
	json.NewDecoder(w.Body).Decode(&resp)
	fmt.Printf("Status: %d, ID: %s, Value: %g\n", w.Code, resp.ID, *resp.Value)

	// Output:
	// Status: 200, ID: Temperature, Value: 42.5
}

func ExampleMetricHandler_UpdateMetric() {
	router, _ := newTestRouter()

	req := httptest.NewRequest(http.MethodPost, "/update/gauge/Alloc/123.456", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)

	// Output:
	// Status: 200
}

func ExampleMetricHandler_GetMetric() {
	router, svc := newTestRouter()
	svc.UpdateGauge(context.Background(), "HeapAlloc", 999.99)

	req := httptest.NewRequest(http.MethodGet, "/value/gauge/HeapAlloc", nil)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Printf("Status: %d, Body: %s\n", w.Code, w.Body.String())

	// Output:
	// Status: 200, Body: 999.99
}

func ExampleMetricHandler_GetMetricJSON() {
	router, svc := newTestRouter()
	svc.UpdateCounter(context.Background(), "PollCount", 42)

	metric := model.Metrics{ID: "PollCount", MType: "counter"}
	body, _ := json.Marshal(metric)

	req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	var resp model.Metrics
	json.NewDecoder(w.Body).Decode(&resp)
	fmt.Printf("Status: %d, ID: %s, Delta: %d\n", w.Code, resp.ID, *resp.Delta)

	// Output:
	// Status: 200, ID: PollCount, Delta: 42
}

func ExampleMetricHandler_UpdateMetricsBatch() {
	router, _ := newTestRouter()

	val := 100.5
	delta := int64(10)
	metrics := []model.Metrics{
		{ID: "Alloc", MType: "gauge", Value: &val},
		{ID: "PollCount", MType: "counter", Delta: &delta},
	}
	body, _ := json.Marshal(metrics)

	req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	fmt.Printf("Status: %d\n", w.Code)

	// Output:
	// Status: 200
}
