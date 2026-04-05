package storage

import (
	"context"
	"fmt"
	"testing"

	"github.com/user/practicum-metrics/internal/model"
)

func BenchmarkMemStorage_UpdateGauge(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.UpdateGauge(ctx, "test_gauge", 123.456)
	}
}

func BenchmarkMemStorage_UpdateCounter(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.UpdateCounter(ctx, "test_counter", 1)
	}
}

func BenchmarkMemStorage_GetGauge(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()
	s.UpdateGauge(ctx, "test_gauge", 123.456)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetGauge(ctx, "test_gauge")
	}
}

func BenchmarkMemStorage_GetCounter(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()
	s.UpdateCounter(ctx, "test_counter", 100)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetCounter(ctx, "test_counter")
	}
}

func BenchmarkMemStorage_GetAllGauges(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		s.UpdateGauge(ctx, fmt.Sprintf("gauge_%d", i), float64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetAllGauges(ctx)
	}
}

func BenchmarkMemStorage_GetAllCounters(b *testing.B) {
	s := NewMemStorage()
	ctx := context.Background()
	for i := 0; i < 100; i++ {
		s.UpdateCounter(ctx, fmt.Sprintf("counter_%d", i), int64(i))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		s.GetAllCounters(ctx)
	}
}

func BenchmarkMemStorage_UpdateMetricsBatch(b *testing.B) {
	s := NewMemStorage()
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
		s.UpdateMetricsBatch(ctx, metrics)
	}
}
