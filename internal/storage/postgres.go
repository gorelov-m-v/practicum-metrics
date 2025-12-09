package storage

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/user/practicum-metrics/internal/model"
	"github.com/user/practicum-metrics/internal/repository"
	"github.com/user/practicum-metrics/internal/retry"
)

type DBStorage struct {
	db          *sql.DB
	gaugeRepo   repository.GaugeRepository
	counterRepo repository.CounterRepository
}

func NewDBStorage(db *sql.DB) *DBStorage {
	return &DBStorage{
		db:          db,
		gaugeRepo:   repository.NewGaugeRepository(db),
		counterRepo: repository.NewCounterRepository(db),
	}
}

func (s *DBStorage) UpdateGauge(ctx context.Context, name string, value float64) error {
	return s.gaugeRepo.Upsert(ctx, name, value)
}

func (s *DBStorage) UpdateCounter(ctx context.Context, name string, value int64) error {
	return s.counterRepo.Add(ctx, name, value)
}

func (s *DBStorage) GetGauge(ctx context.Context, name string) (float64, bool) {
	value, err := s.gaugeRepo.Get(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false
		}
		return 0, false
	}
	return value, true
}

func (s *DBStorage) GetCounter(ctx context.Context, name string) (int64, bool) {
	value, err := s.counterRepo.Get(ctx, name)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, false
		}
		return 0, false
	}
	return value, true
}

func (s *DBStorage) GetAllGauges(ctx context.Context) map[string]float64 {
	result, err := s.gaugeRepo.GetAll(ctx)
	if err != nil {
		return make(map[string]float64)
	}
	return result
}

func (s *DBStorage) GetAllCounters(ctx context.Context) map[string]int64 {
	result, err := s.counterRepo.GetAll(ctx)
	if err != nil {
		return make(map[string]int64)
	}
	return result
}

func (s *DBStorage) SetGauges(ctx context.Context, gauges map[string]float64) error {
	for name, value := range gauges {
		if err := s.gaugeRepo.Upsert(ctx, name, value); err != nil {
			return err
		}
	}
	return nil
}

func (s *DBStorage) SetCounters(ctx context.Context, counters map[string]int64) error {
	for name, value := range counters {
		if err := s.counterRepo.Set(ctx, name, value); err != nil {
			return err
		}
	}
	return nil
}

func (s *DBStorage) UpdateMetricsBatch(ctx context.Context, metrics []model.Metrics) error {
	return retry.WithRetry(func() error {
		tx, err := s.db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}
		defer tx.Rollback()

		gauges := make(map[string]float64)
		counters := make(map[string]int64)

		for _, metric := range metrics {
			switch MetricType(metric.MType) {
			case Gauge:
				if metric.Value != nil {
					gauges[metric.ID] = *metric.Value
				}
			case Counter:
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

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit transaction: %w", err)
		}

		return nil
	})
}
