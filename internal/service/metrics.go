// Package service implements business logic for metric operations.
package service

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/user/practicum-metrics/internal/database"
	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/repository"
	"github.com/user/practicum-metrics/internal/storage"
)

// ServiceTimeouts holds timeout configuration for service operations.
type ServiceTimeouts struct {
	DefaultOperationTimeout time.Duration
}

// DefaultServiceTimeouts returns the default timeout configuration.
func DefaultServiceTimeouts() ServiceTimeouts {
	return ServiceTimeouts{
		DefaultOperationTimeout: 5 * time.Second,
	}
}

// MetricsService provides business logic for updating and retrieving metrics.
type MetricsService struct {
	storage     storage.Storage
	txManager   database.TransactionManager
	gaugeRepo   repository.GaugeRepository
	counterRepo repository.CounterRepository
	timeouts    ServiceTimeouts
}

// NewMetricsService creates a new MetricsService with the given dependencies.
func NewMetricsService(storage storage.Storage, txManager database.TransactionManager, gaugeRepo repository.GaugeRepository, counterRepo repository.CounterRepository) *MetricsService {
	return &MetricsService{
		storage:     storage,
		txManager:   txManager,
		gaugeRepo:   gaugeRepo,
		counterRepo: counterRepo,
		timeouts:    DefaultServiceTimeouts(),
	}
}

func (s *MetricsService) UpdateGauge(ctx context.Context, name string, value float64) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeouts.DefaultOperationTimeout)
	defer cancel()

	if value < 0 {
		return fmt.Errorf("gauge value cannot be negative")
	}
	return s.storage.UpdateGauge(ctx, name, value)
}

func (s *MetricsService) UpdateCounter(ctx context.Context, name string, value int64) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeouts.DefaultOperationTimeout)
	defer cancel()

	if value < 0 {
		return fmt.Errorf("counter value cannot be negative")
	}
	return s.storage.UpdateCounter(ctx, name, value)
}

func (s *MetricsService) GetGauge(ctx context.Context, name string) (float64, bool) {
	ctx, cancel := context.WithTimeout(ctx, s.timeouts.DefaultOperationTimeout)
	defer cancel()

	return s.storage.GetGauge(ctx, name)
}

func (s *MetricsService) GetCounter(ctx context.Context, name string) (int64, bool) {
	ctx, cancel := context.WithTimeout(ctx, s.timeouts.DefaultOperationTimeout)
	defer cancel()

	return s.storage.GetCounter(ctx, name)
}

func (s *MetricsService) GetAllGauges(ctx context.Context) map[string]float64 {
	ctx, cancel := context.WithTimeout(ctx, s.timeouts.DefaultOperationTimeout)
	defer cancel()

	return s.storage.GetAllGauges(ctx)
}

func (s *MetricsService) GetAllCounters(ctx context.Context) map[string]int64 {
	ctx, cancel := context.WithTimeout(ctx, s.timeouts.DefaultOperationTimeout)
	defer cancel()

	return s.storage.GetAllCounters(ctx)
}

func (s *MetricsService) UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error {
	ctx, cancel := context.WithTimeout(ctx, s.timeouts.DefaultOperationTimeout)
	defer cancel()

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

	// For storage that supports batch updates (e.g. MemStorage)
	if batcher, ok := s.storage.(storage.BatchUpdater); ok && s.txManager == nil {
		return batcher.UpdateMetricsBatch(ctx, metrics)
	}

	// Fallback for storage without batch support and no transaction manager
	if s.txManager == nil {
		for _, metric := range metrics {
			switch storage.MetricType(metric.MType) {
			case storage.Gauge:
				if metric.Value != nil {
					if err := s.storage.UpdateGauge(ctx, metric.ID, *metric.Value); err != nil {
						return err
					}
				}
			case storage.Counter:
				if metric.Delta != nil {
					if err := s.storage.UpdateCounter(ctx, metric.ID, *metric.Delta); err != nil {
						return err
					}
				}
			}
		}
		return nil
	}

	// For DB storage - use transaction manager
	return s.txManager.WithTransaction(ctx, func(tx *sql.Tx) error {
		gauges := make(map[string]float64)
		counters := make(map[string]int64)

		for _, metric := range metrics {
			switch storage.MetricType(metric.MType) {
			case storage.Gauge:
				if metric.Value != nil {
					gauges[metric.ID] = *metric.Value
				}
			case storage.Counter:
				if metric.Delta != nil {
					counters[metric.ID] += *metric.Delta
				}
			}
		}

		if len(gauges) > 0 {
			if err := s.gaugeRepo.UpsertBatch(ctx, tx, gauges); err != nil {
				return fmt.Errorf("failed to upsert gauges: %w", err)
			}
		}

		if len(counters) > 0 {
			if err := s.counterRepo.AddBatch(ctx, tx, counters); err != nil {
				return fmt.Errorf("failed to add counters: %w", err)
			}
		}

		return nil
	})
}
