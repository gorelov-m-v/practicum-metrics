package storage

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"go.uber.org/zap"
)

func TestNewPersister(t *testing.T) {
	tests := []struct {
		name          string
		storeInterval int
		filePath      string
	}{
		{"with interval", 300, "/tmp/test.json"},
		{"sync mode", 0, "/tmp/test.json"},
		{"short interval", 1, "/tmp/test.json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStorage()
			logger := zap.NewNop()
			persister := NewPersister(store, tt.filePath, tt.storeInterval, logger)

			if persister == nil {
				t.Fatal("NewPersister returned nil")
			}

			if persister.storage != store {
				t.Error("storage not set correctly")
			}

			if persister.filePath != tt.filePath {
				t.Errorf("expected filePath %s, got %s", tt.filePath, persister.filePath)
			}

			if persister.storeInterval != tt.storeInterval {
				t.Errorf("expected storeInterval %d, got %d", tt.storeInterval, persister.storeInterval)
			}

			if persister.stopChan == nil {
				t.Error("stopChan not initialized")
			}
		})
	}
}

func TestPersister_IsSyncMode(t *testing.T) {
	tests := []struct {
		name          string
		storeInterval int
		expected      bool
	}{
		{"sync mode", 0, true},
		{"async mode", 300, false},
		{"async with 1 second", 1, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStorage()
			logger := zap.NewNop()
			persister := NewPersister(store, "/tmp/test.json", tt.storeInterval, logger)

			if persister.IsSyncMode() != tt.expected {
				t.Errorf("expected IsSyncMode %v, got %v", tt.expected, persister.IsSyncMode())
			}
		})
	}
}

func TestPersister_SaveSync(t *testing.T) {
	tests := []struct {
		name   string
		gauges map[string]float64
	}{
		{
			name:   "save with metrics",
			gauges: map[string]float64{"Alloc": 123.456, "HeapAlloc": 789.012},
		},
		{
			name:   "save empty storage",
			gauges: map[string]float64{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := NewMemStorage()
			for name, value := range tt.gauges {
				store.UpdateGauge(name, value)
			}

			tmpFile := filepath.Join(t.TempDir(), "metrics.json")
			logger := zap.NewNop()
			persister := NewPersister(store, tmpFile, 0, logger)

			persister.SaveSync()

			if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
				t.Error("file was not created")
			}

			loadedStore := NewMemStorage()
			if err := loadedStore.LoadFromFile(tmpFile); err != nil {
				t.Fatalf("failed to load saved file: %v", err)
			}

			for name, expectedValue := range tt.gauges {
				value, exists := loadedStore.GetGauge(name)
				if !exists {
					t.Errorf("gauge %s not found after reload", name)
				} else if value != expectedValue {
					t.Errorf("gauge %s: expected %f, got %f", name, expectedValue, value)
				}
			}
		})
	}
}

func TestPersister_SaveSync_InvalidPath(t *testing.T) {
	store := NewMemStorage()
	store.UpdateGauge("TestMetric", 100.0)

	logger := zap.NewNop()
	persister := NewPersister(store, "/invalid/path/that/does/not/exist/metrics.json", 0, logger)

	persister.SaveSync()
}

func TestPersister_StartStop_SyncMode(t *testing.T) {
	store := NewMemStorage()
	tmpFile := filepath.Join(t.TempDir(), "metrics.json")
	logger := zap.NewNop()
	persister := NewPersister(store, tmpFile, 0, logger)

	persister.Start()

	time.Sleep(10 * time.Millisecond)

	persister.Stop()
}

func TestPersister_StartStop_AsyncMode(t *testing.T) {
	store := NewMemStorage()
	store.UpdateGauge("TestMetric", 123.456)

	tmpFile := filepath.Join(t.TempDir(), "metrics.json")

	logger := zap.NewNop()
	persister := NewPersister(store, tmpFile, 1, logger)
	persister.Start()

	time.Sleep(1500 * time.Millisecond)

	persister.Stop()

	if _, err := os.Stat(tmpFile); os.IsNotExist(err) {
		t.Error("file was not created by periodic save")
	}

	loadedStore := NewMemStorage()
	if err := loadedStore.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("failed to load saved file: %v", err)
	}

	value, exists := loadedStore.GetGauge("TestMetric")
	if !exists {
		t.Error("metric not found after periodic save")
	} else if value != 123.456 {
		t.Errorf("expected value 123.456, got %f", value)
	}
}

func TestPersister_PeriodicSave_MultipleUpdates(t *testing.T) {
	store := NewMemStorage()
	tmpFile := filepath.Join(t.TempDir(), "metrics.json")

	logger := zap.NewNop()
	persister := NewPersister(store, tmpFile, 1, logger)
	persister.Start()
	defer persister.Stop()

	store.UpdateGauge("Metric1", 100.0)
	time.Sleep(500 * time.Millisecond)

	store.UpdateGauge("Metric2", 200.0)
	time.Sleep(600 * time.Millisecond)

	loadedStore := NewMemStorage()
	if err := loadedStore.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("failed to load saved file: %v", err)
	}

	value1, exists1 := loadedStore.GetGauge("Metric1")
	if !exists1 {
		t.Error("Metric1 not found")
	} else if value1 != 100.0 {
		t.Errorf("Metric1: expected 100.0, got %f", value1)
	}
}

func TestPersister_Stop_MultipleTimes(t *testing.T) {
	store := NewMemStorage()
	tmpFile := filepath.Join(t.TempDir(), "metrics.json")
	logger := zap.NewNop()
	persister := NewPersister(store, tmpFile, 1, logger)

	persister.Start()
	time.Sleep(50 * time.Millisecond)

	persister.Stop()

	defer func() {
		if r := recover(); r != nil {
			t.Error("calling Stop multiple times should not panic")
		}
	}()
}
