package service

import (
	"fmt"

	"github.com/user/practicum-metrics/internal/storage"
)

type MetricsService struct {
	storage   storage.Storage
	persister *storage.Persister
}

func NewMetricsService(storage storage.Storage) *MetricsService {
	return &MetricsService{
		storage: storage,
	}
}

func NewMetricsServiceWithPersister(storage storage.Storage, persister *storage.Persister) *MetricsService {
	return &MetricsService{
		storage:   storage,
		persister: persister,
	}
}

func (s *MetricsService) saveIfNeeded() {
	if s.persister != nil && s.persister.IsSyncMode() {
		s.persister.SaveSync()
	}
}

func (s *MetricsService) UpdateGauge(name string, value float64) error {
	if value < 0 {
		return fmt.Errorf("gauge value cannot be negative")
	}
	s.storage.UpdateGauge(name, value)
	s.saveIfNeeded()
	return nil
}

func (s *MetricsService) UpdateCounter(name string, value int64) error {
	if value < 0 {
		return fmt.Errorf("counter value cannot be negative")
	}
	s.storage.UpdateCounter(name, value)
	s.saveIfNeeded()
	return nil
}

func (s *MetricsService) GetGauge(name string) (float64, bool) {
	return s.storage.GetGauge(name)
}

func (s *MetricsService) GetCounter(name string) (int64, bool) {
	return s.storage.GetCounter(name)
}

func (s *MetricsService) GetAllGauges() map[string]float64 {
	return s.storage.GetAllGauges()
}

func (s *MetricsService) GetAllCounters() map[string]int64 {
	return s.storage.GetAllCounters()
}
