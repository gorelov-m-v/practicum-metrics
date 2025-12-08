package storage

import (
	"context"
	"sync"

	"github.com/user/practicum-metrics/internal/model"
)

type MetricType string

const (
	Gauge   MetricType = "gauge"
	Counter MetricType = "counter"
)

type Storage interface {
	UpdateGauge(ctx context.Context, name string, value float64) error
	UpdateCounter(ctx context.Context, name string, value int64) error
	GetGauge(ctx context.Context, name string) (float64, bool)
	GetCounter(ctx context.Context, name string) (int64, bool)
	GetAllGauges(ctx context.Context) map[string]float64
	GetAllCounters(ctx context.Context) map[string]int64
	SetGauges(ctx context.Context, gauges map[string]float64) error
	SetCounters(ctx context.Context, counters map[string]int64) error
	UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error
}

type MemStorage struct {
	mu       sync.RWMutex
	gauges   map[string]float64
	counters map[string]int64
}

func NewMemStorage() *MemStorage {
	return &MemStorage{
		gauges:   make(map[string]float64),
		counters: make(map[string]int64),
	}
}

func (s *MemStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.gauges[name] = value
	return nil
}

func (s *MemStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.counters[name] += value
	return nil
}

func (s *MemStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.gauges[name]
	return value, exists
}

func (s *MemStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.counters[name]
	return value, exists
}

func (s *MemStorage) GetAllGauges(ctx context.Context) map[string]float64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]float64, len(s.gauges))
	for k, v := range s.gauges {
		result[k] = v
	}
	return result
}

func (s *MemStorage) GetAllCounters(ctx context.Context) map[string]int64 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make(map[string]int64, len(s.counters))
	for k, v := range s.counters {
		result[k] = v
	}
	return result
}

func (s *MemStorage) SetGauges(ctx context.Context, gauges map[string]float64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for name, value := range gauges {
		s.gauges[name] = value
	}
	return nil
}

func (s *MemStorage) SetCounters(ctx context.Context, counters map[string]int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for name, value := range counters {
		s.counters[name] = value
	}
	return nil
}

func (s *MemStorage) UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, metric := range metrics {
		switch MetricType(metric.MType) {
		case Gauge:
			if metric.Value != nil {
				s.gauges[metric.ID] = *metric.Value
			}
		case Counter:
			if metric.Delta != nil {
				s.counters[metric.ID] += *metric.Delta
			}
		}
	}
	return nil
}
