package service

import (
	"context"
	"fmt"

	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/storage"
)

type MetricsService struct {
	storage storage.Storage
}

func NewMetricsService(storage storage.Storage) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

func (s *MetricsService) UpdateGauge(ctx context.Context, name string, value float64) error {
	if value < 0 {
		return fmt.Errorf("gauge value cannot be negative")
	}
	return s.storage.UpdateGauge(ctx, name, value)
}

func (s *MetricsService) UpdateCounter(ctx context.Context, name string, value int64) error {
	if value < 0 {
		return fmt.Errorf("counter value cannot be negative")
	}
	return s.storage.UpdateCounter(ctx, name, value)
}

func (s *MetricsService) GetGauge(ctx context.Context, name string) (float64, bool) {
	return s.storage.GetGauge(ctx, name)
}

func (s *MetricsService) GetCounter(ctx context.Context, name string) (int64, bool) {
	return s.storage.GetCounter(ctx, name)
}

func (s *MetricsService) GetAllGauges(ctx context.Context) map[string]float64 {
	return s.storage.GetAllGauges(ctx)
}

func (s *MetricsService) GetAllCounters(ctx context.Context) map[string]int64 {
	return s.storage.GetAllCounters(ctx)
}

func (s *MetricsService) UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error {
	for _, metric := range metrics {
		switch storage.MetricType(metric.MType) {
		case storage.Gauge:
			if metric.Value != nil && *metric.Value < 0 {
				return fmt.Errorf("gauge value cannot be negative for metric %s", metric.ID)
			}
		case storage.Counter:
			if metric.Delta != nil && *metric.Delta < 0 {
				return fmt.Errorf("counter value cannot be negative for metric %s", metric.ID)
			}
		}
	}
	return s.storage.UpdateMetricsBatch(ctx, metrics)
}
