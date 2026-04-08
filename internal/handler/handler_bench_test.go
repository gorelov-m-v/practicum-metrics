package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

func BenchmarkHandler_UpdateMetricJSON_Gauge(b *testing.B) {
	store := storage.NewMemStorage()
	svc := service.NewMetricsService(store, nil, nil, nil)
	h, _ := NewMetricHandler(svc, nil, nil, nil)
	router := setupRouter(h)

	val := 123.456
	metric := model.Metrics{ID: "BenchGauge", MType: "gauge", Value: &val}
	body, _ := json.Marshal(metric)

	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		b.StartTimer()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_UpdateMetricJSON_Counter(b *testing.B) {
	store := storage.NewMemStorage()
	svc := service.NewMetricsService(store, nil, nil, nil)
	h, _ := NewMetricHandler(svc, nil, nil, nil)
	router := setupRouter(h)

	delta := int64(1)
	metric := model.Metrics{ID: "BenchCounter", MType: "counter", Delta: &delta}
	body, _ := json.Marshal(metric)

	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/update", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		b.StartTimer()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_GetMetricJSON(b *testing.B) {
	store := storage.NewMemStorage()
	svc := service.NewMetricsService(store, nil, nil, nil)
	svc.UpdateGauge(context.Background(), "BenchGauge", 123.456)
	h, _ := NewMetricHandler(svc, nil, nil, nil)
	router := setupRouter(h)

	metric := model.Metrics{ID: "BenchGauge", MType: "gauge"}
	body, _ := json.Marshal(metric)

	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/value", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		b.StartTimer()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_UpdateMetricsBatch(b *testing.B) {
	store := storage.NewMemStorage()
	svc := service.NewMetricsService(store, nil, nil, nil)
	h, _ := NewMetricHandler(svc, nil, nil, nil)
	router := setupRouter(h)

	metrics := make([]model.Metrics, 0, 100)
	for i := 0; i < 50; i++ {
		v := float64(i)
		metrics = append(metrics, model.Metrics{
			ID:    fmt.Sprintf("gauge_%d", i),
			MType: "gauge",
			Value: &v,
		})
	}
	for i := 0; i < 50; i++ {
		d := int64(i)
		metrics = append(metrics, model.Metrics{
			ID:    fmt.Sprintf("counter_%d", i),
			MType: "counter",
			Delta: &d,
		})
	}
	body, _ := json.Marshal(metrics)

	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodPost, "/updates", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		b.StartTimer()
		router.ServeHTTP(w, req)
	}
}

func BenchmarkHandler_ListMetrics(b *testing.B) {
	store := storage.NewMemStorage()
	svc := service.NewMetricsService(store, nil, nil, nil)
	for i := 0; i < 30; i++ {
		svc.UpdateGauge(context.Background(), fmt.Sprintf("gauge_%d", i), float64(i))
		svc.UpdateCounter(context.Background(), fmt.Sprintf("counter_%d", i), int64(i))
	}
	h, _ := NewMetricHandler(svc, nil, nil, nil)
	router := setupRouter(h)

	for b.Loop() {
		b.StopTimer()
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()
		b.StartTimer()
		router.ServeHTTP(w, req)
	}
}
