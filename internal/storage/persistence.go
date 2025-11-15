package storage

import (
	"time"

	"go.uber.org/zap"
)

type Persister struct {
	storage       Storage
	filePath      string
	storeInterval int
	stopChan      chan struct{}
	logger        *zap.Logger
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
		go p.periodicSave()
	}
}

func (p *Persister) Stop() {
	close(p.stopChan)
}

func (p *Persister) SaveSync() {
	if err := p.storage.SaveToFile(p.filePath); err != nil {
		p.logger.Error("Failed to save metrics", zap.Error(err))
	}
}

func (p *Persister) periodicSave() {
	ticker := time.NewTicker(time.Duration(p.storeInterval) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := p.storage.SaveToFile(p.filePath); err != nil {
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
