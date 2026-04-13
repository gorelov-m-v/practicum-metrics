package storage

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/user/practicum-metrics/internal/model"
)

type Persister struct {
	storage       Storage
	filePath      string
	storeInterval int
	stopChan      chan struct{}
	logger        *zap.Logger
	stopOnce      sync.Once
	wg            sync.WaitGroup
}

func NewPersister(storage Storage, filePath string, storeInterval int, logger *zap.Logger) *Persister {
	return &Persister{
		storage:       storage,
		filePath:      filePath,
		storeInterval: storeInterval,
		stopChan:      make(chan struct{}),
		logger:        logger,
	}
}

func (p *Persister) Start() {
	if p.storeInterval > 0 {
		p.wg.Add(1)
		go p.periodicSave()
	}
}

func (p *Persister) Stop() {
	p.stopOnce.Do(func() {
		close(p.stopChan)
	})
	p.wg.Wait()
}

func (p *Persister) SaveSync() {
	if err := p.saveToFile(); err != nil {
		p.logger.Error("Failed to save metrics", zap.Error(err))
	}
}

func (p *Persister) periodicSave() {
	defer p.wg.Done()

	ticker := time.NewTicker(time.Duration(p.storeInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.saveToFile(); err != nil {
				p.logger.Error("Failed to save metrics to file", zap.Error(err))
			} else {
				p.logger.Info("Metrics saved to file")
			}
		case <-p.stopChan:
			return
		}
	}
}

func (p *Persister) IsSyncMode() bool {
	return p.storeInterval == 0
}

func (p *Persister) Restore() error {
	if err := p.loadFromFile(); err != nil {
		return err
	}
	p.logger.Info("Metrics restored from file")
	return nil
}

func (p *Persister) saveToFile() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	gauges := p.storage.GetAllGauges(ctx)
	counters := p.storage.GetAllCounters(ctx)

	var metrics []model.Metrics

	for name, value := range gauges {
		v := value
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: string(Gauge),
			Value: &v,
		})
	}

	for name, delta := range counters {
		d := delta
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: string(Counter),
			Delta: &d,
		})
	}

	data, err := json.MarshalIndent(metrics, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(p.filePath, data, 0644)
}

func (p *Persister) loadFromFile() error {
	data, err := os.ReadFile(p.filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}

	var metrics []model.Metrics
	if err := json.Unmarshal(data, &metrics); err != nil {
		return err
	}

	gauges := make(map[string]float64)
	counters := make(map[string]int64)

	for _, m := range metrics {
		switch MetricType(m.MType) {
		case Gauge:
			if m.Value != nil {
				gauges[m.ID] = *m.Value
			}
		case Counter:
			if m.Delta != nil {
				counters[m.ID] = *m.Delta
			}
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := p.storage.SetGauges(ctx, gauges); err != nil {
		return err
	}
	if err := p.storage.SetCounters(ctx, counters); err != nil {
		return err
	}

	return nil
}
