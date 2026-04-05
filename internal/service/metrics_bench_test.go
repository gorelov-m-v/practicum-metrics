package service

import (
	"context"
	"fmt"
	"testing"

	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/storage"
)

func BenchmarkMetricsService_UpdateGauge(b *testing.B) {
	store := storage.NewMemStorage()
	svc := NewMetricsService(store, nil, nil, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.UpdateGauge(ctx, "test_gauge", 123.456)
	}
}

func BenchmarkMetricsService_UpdateCounter(b *testing.B) {
	store := storage.NewMemStorage()
	svc := NewMetricsService(store, nil, nil, nil)
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.UpdateCounter(ctx, "test_counter", 1)
	}
}

func BenchmarkMetricsService_GetGauge(b *testing.B) {
	store := storage.NewMemStorage()
	svc := NewMetricsService(store, nil, nil, nil)
	ctx := context.Background()
	svc.UpdateGauge(ctx, "test_gauge", 123.456)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetGauge(ctx, "test_gauge")
	}
}

func BenchmarkMetricsService_GetCounter(b *testing.B) {
	store := storage.NewMemStorage()
	svc := NewMetricsService(store, nil, nil, nil)
	ctx := context.Background()
	svc.UpdateCounter(ctx, "test_counter", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.GetCounter(ctx, "test_counter")
	}
}

func BenchmarkMetricsService_UpdateMetricsBatch(b *testing.B) {
	store := storage.NewMemStorage()
	svc := NewMetricsService(store, nil, nil, nil)
	ctx := context.Background()

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		svc.UpdateMetricsBatch(ctx, metrics)
	}
}
