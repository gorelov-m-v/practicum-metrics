package service

import (
	"testing"

	"github.com/user/practicum-metrics/internal/storage"
	"go.uber.org/zap"
)

func TestNewMetricsService(t *testing.T) {
	store := storage.NewMemStorage()
	service := NewMetricsService(store)

	if service == nil {
		t.Fatal("NewMetricsService returned nil")
	}

	if service.storage == nil {
		t.Error("storage not initialized")
	}
}

func TestMetricsService_UpdateGauge(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      float64
		wantErr    bool
	}{
		{
			name:       "valid positive value",
			metricName: "test_gauge",
			value:      123.45,
			wantErr:    false,
		},
		{
			name:       "valid zero value",
			metricName: "zero_gauge",
			value:      0,
			wantErr:    false,
		},
		{
			name:       "negative value should error",
			metricName: "negative_gauge",
			value:      -10.5,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			service := NewMetricsService(store)

			err := service.UpdateGauge(tt.metricName, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateGauge() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				value, exists := service.GetGauge(tt.metricName)
				if !exists {
					t.Errorf("gauge %s not found after update", tt.metricName)
				}
				if value != tt.value {
					t.Errorf("expected value %f, got %f", tt.value, value)
				}
			}
		})
	}
}

func TestMetricsService_UpdateCounter(t *testing.T) {
	tests := []struct {
		name       string
		metricName string
		value      int64
		wantErr    bool
	}{
		{
			name:       "valid positive value",
			metricName: "test_counter",
			value:      100,
			wantErr:    false,
		},
		{
			name:       "valid zero value",
			metricName: "zero_counter",
			value:      0,
			wantErr:    false,
		},
		{
			name:       "negative value should error",
			metricName: "negative_counter",
			value:      -50,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			service := NewMetricsService(store)

			err := service.UpdateCounter(tt.metricName, tt.value)

			if (err != nil) != tt.wantErr {
				t.Errorf("UpdateCounter() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				value, exists := service.GetCounter(tt.metricName)
				if !exists {
					t.Errorf("counter %s not found after update", tt.metricName)
				}
				if value != tt.value {
					t.Errorf("expected value %d, got %d", tt.value, value)
				}
			}
		})
	}
}

func TestMetricsService_GetGauge(t *testing.T) {
	store := storage.NewMemStorage()
	service := NewMetricsService(store)

	service.UpdateGauge("test_gauge", 42.5)

	tests := []struct {
		name       string
		metricName string
		wantValue  float64
		wantExists bool
	}{
		{
			name:       "existing gauge",
			metricName: "test_gauge",
			wantValue:  42.5,
			wantExists: true,
		},
		{
			name:       "non-existing gauge",
			metricName: "unknown_gauge",
			wantValue:  0,
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := service.GetGauge(tt.metricName)

			if exists != tt.wantExists {
				t.Errorf("GetGauge() exists = %v, want %v", exists, tt.wantExists)
			}

			if tt.wantExists && value != tt.wantValue {
				t.Errorf("GetGauge() value = %f, want %f", value, tt.wantValue)
			}
		})
	}
}

func TestMetricsService_GetCounter(t *testing.T) {
	store := storage.NewMemStorage()
	service := NewMetricsService(store)

	service.UpdateCounter("test_counter", 100)

	tests := []struct {
		name       string
		metricName string
		wantValue  int64
		wantExists bool
	}{
		{
			name:       "existing counter",
			metricName: "test_counter",
			wantValue:  100,
			wantExists: true,
		},
		{
			name:       "non-existing counter",
			metricName: "unknown_counter",
			wantValue:  0,
			wantExists: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, exists := service.GetCounter(tt.metricName)

			if exists != tt.wantExists {
				t.Errorf("GetCounter() exists = %v, want %v", exists, tt.wantExists)
			}

			if tt.wantExists && value != tt.wantValue {
				t.Errorf("GetCounter() value = %d, want %d", value, tt.wantValue)
			}
		})
	}
}

func TestMetricsService_GetAllGauges(t *testing.T) {
	store := storage.NewMemStorage()
	service := NewMetricsService(store)

	service.UpdateGauge("gauge1", 10.5)
	service.UpdateGauge("gauge2", 20.5)
	service.UpdateGauge("gauge3", 30.5)

	gauges := service.GetAllGauges()

	if len(gauges) != 3 {
		t.Errorf("expected 3 gauges, got %d", len(gauges))
	}

	expected := map[string]float64{
		"gauge1": 10.5,
		"gauge2": 20.5,
		"gauge3": 30.5,
	}

	for name, expectedValue := range expected {
		if value, exists := gauges[name]; !exists {
			t.Errorf("gauge %s not found in GetAllGauges()", name)
		} else if value != expectedValue {
			t.Errorf("gauge %s: expected value %f, got %f", name, expectedValue, value)
		}
	}
}

func TestMetricsService_GetAllCounters(t *testing.T) {
	store := storage.NewMemStorage()
	service := NewMetricsService(store)

	service.UpdateCounter("counter1", 10)
	service.UpdateCounter("counter2", 20)
	service.UpdateCounter("counter3", 30)

	counters := service.GetAllCounters()

	if len(counters) != 3 {
		t.Errorf("expected 3 counters, got %d", len(counters))
	}

	expected := map[string]int64{
		"counter1": 10,
		"counter2": 20,
		"counter3": 30,
	}

	for name, expectedValue := range expected {
		if value, exists := counters[name]; !exists {
			t.Errorf("counter %s not found in GetAllCounters()", name)
		} else if value != expectedValue {
			t.Errorf("counter %s: expected value %d, got %d", name, expectedValue, value)
		}
	}
}

func TestMetricsService_UpdateCounter_Accumulation(t *testing.T) {
	store := storage.NewMemStorage()
	service := NewMetricsService(store)

	service.UpdateCounter("accumulator", 10)
	service.UpdateCounter("accumulator", 20)
	service.UpdateCounter("accumulator", 30)

	value, exists := service.GetCounter("accumulator")
	if !exists {
		t.Fatal("counter not found")
	}

	expected := int64(60)
	if value != expected {
		t.Errorf("expected accumulated value %d, got %d", expected, value)
	}
}

func TestMetricsService_UpdateGauge_Overwrite(t *testing.T) {
	store := storage.NewMemStorage()
	service := NewMetricsService(store)

	service.UpdateGauge("temperature", 20.0)
	service.UpdateGauge("temperature", 25.0)
	service.UpdateGauge("temperature", 30.0)

	value, exists := service.GetGauge("temperature")
	if !exists {
		t.Fatal("gauge not found")
	}

	expected := 30.0
	if value != expected {
		t.Errorf("expected latest value %f, got %f", expected, value)
	}
}

func TestNewMetricsServiceWithPersister(t *testing.T) {
	tests := []struct {
		name          string
		storeInterval int
	}{
		{"with sync mode persister", 0},
		{"with async mode persister", 300},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := storage.NewMemStorage()
			logger := zap.NewNop()
			persister := storage.NewPersister(store, "/tmp/test.json", tt.storeInterval, logger)
			service := NewMetricsServiceWithPersister(store, persister)

			if service == nil {
				t.Fatal("NewMetricsServiceWithPersister returned nil")
			}

			if service.storage == nil {
				t.Error("storage not initialized")
			}

			if service.persister == nil {
				t.Error("persister not set")
			}

			if service.persister != persister {
				t.Error("persister mismatch")
			}
		})
	}
}

func TestMetricsService_SaveIfNeeded_WithSyncPersister(t *testing.T) {
	store := storage.NewMemStorage()
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/metrics.json"

	logger := zap.NewNop()
	persister := storage.NewPersister(store, tmpFile, 0, logger)
	service := NewMetricsServiceWithPersister(store, persister)

	err := service.UpdateGauge("TestMetric", 123.456)
	if err != nil {
		t.Fatalf("UpdateGauge failed: %v", err)
	}

	loadedStore := storage.NewMemStorage()
	if err := loadedStore.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("failed to load saved file: %v", err)
	}

	value, exists := loadedStore.GetGauge("TestMetric")
	if !exists {
		t.Error("metric was not saved by sync persister")
	} else if value != 123.456 {
		t.Errorf("expected value 123.456, got %f", value)
	}
}

func TestMetricsService_SaveIfNeeded_WithAsyncPersister(t *testing.T) {
	store := storage.NewMemStorage()
	tmpDir := t.TempDir()
	tmpFile := tmpDir + "/metrics.json"

	logger := zap.NewNop()
	persister := storage.NewPersister(store, tmpFile, 300, logger)
	service := NewMetricsServiceWithPersister(store, persister)

	err := service.UpdateCounter("TestCounter", 42)
	if err != nil {
		t.Fatalf("UpdateCounter failed: %v", err)
	}

	value, exists := store.GetCounter("TestCounter")
	if !exists {
		t.Error("counter was not stored")
	} else if value != 42 {
		t.Errorf("expected counter value 42, got %d", value)
	}
}

func TestMetricsService_SaveIfNeeded_NoPersister(t *testing.T) {
	store := storage.NewMemStorage()
	service := NewMetricsService(store)

	err := service.UpdateGauge("TestMetric", 100.0)
	if err != nil {
		t.Errorf("UpdateGauge with no persister should not fail: %v", err)
	}

	err = service.UpdateCounter("TestCounter", 50)
	if err != nil {
		t.Errorf("UpdateCounter with no persister should not fail: %v", err)
	}

	value, exists := service.GetGauge("TestMetric")
	if !exists || value != 100.0 {
		t.Error("gauge should be stored even without persister")
	}
}
