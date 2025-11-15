package storage

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/user/practicum-metrics/internal/model"
)

func TestNewMemStorage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"create new storage"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			if storage == nil {
				t.Fatal("NewMemStorage returned nil")
			}

			if storage.gauges == nil {
				t.Error("gauges map not initialized")
			}

			if storage.counters == nil {
				t.Error("counters map not initialized")
			}
		})
	}
}

func TestUpdateGauge(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value float64
	}{
		{"positive value", "Alloc", 123.456},
		{"zero value", "ZeroMetric", 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()
			storage.UpdateGauge(tt.key, tt.value)

			value, exists := storage.GetGauge(tt.key)

			if !exists {
				t.Errorf("gauge %s was not stored", tt.key)
			}

			if value != tt.value {
				t.Errorf("expected value %f, got %f", tt.value, value)
			}
		})
	}
}

func TestUpdateGaugeOverwrite(t *testing.T) {
	tests := []struct {
		name       string
		key        string
		firstValue float64
		finalValue float64
	}{
		{"overwrite with larger value", "TestMetric", 100.0, 200.0},
		{"overwrite with smaller value", "TestMetric2", 200.0, 100.0},
		{"overwrite with zero", "TestMetric3", 100.0, 0.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			storage.UpdateGauge(tt.key, tt.firstValue)
			value, _ := storage.GetGauge(tt.key)
			if value != tt.firstValue {
				t.Errorf("expected initial value %f, got %f", tt.firstValue, value)
			}

			storage.UpdateGauge(tt.key, tt.finalValue)
			value, exists := storage.GetGauge(tt.key)

			if !exists {
				t.Errorf("gauge %s was not stored", tt.key)
			}

			if value != tt.finalValue {
				t.Errorf("expected value %f after overwrite, got %f", tt.finalValue, value)
			}
		})
	}
}

func TestUpdateCounter(t *testing.T) {
	tests := []struct {
		name          string
		key           string
		values        []int64
		expectedTotal int64
	}{
		{"single update", "Counter1", []int64{5}, 5},
		{"multiple updates", "Counter2", []int64{1, 2, 3, 4, 5}, 15},
		{"zero values", "Counter3", []int64{0, 0, 0}, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			for _, value := range tt.values {
				storage.UpdateCounter(tt.key, value)
			}

			total, exists := storage.GetCounter(tt.key)

			if !exists {
				t.Errorf("counter %s was not stored", tt.key)
			}

			if total != tt.expectedTotal {
				t.Errorf("expected total %d, got %d", tt.expectedTotal, total)
			}
		})
	}
}

func TestUpdateCounterAccumulation(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		updates  int
		expected int64
	}{
		{"three updates", "PollCount", 3, 3},
		{"five updates", "RequestCount", 5, 5},
		{"one update", "SingleCount", 1, 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			for i := 0; i < tt.updates; i++ {
				storage.UpdateCounter(tt.key, 1)
			}

			count, exists := storage.GetCounter(tt.key)

			if !exists {
				t.Errorf("counter %s was not stored", tt.key)
			}

			if count != tt.expected {
				t.Errorf("expected count %d, got %d", tt.expected, count)
			}
		})
	}
}

func TestMultipleMetrics(t *testing.T) {
	tests := []struct {
		name     string
		gauges   map[string]float64
		counters map[string]int64
	}{
		{
			name: "multiple gauges and counters",
			gauges: map[string]float64{
				"Alloc":      100.0,
				"HeapAlloc":  200.0,
				"TotalAlloc": 300.0,
			},
			counters: map[string]int64{
				"PollCount": 5,
				"Requests":  10,
				"Errors":    2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			for name, value := range tt.gauges {
				storage.UpdateGauge(name, value)
			}

			for name, value := range tt.counters {
				storage.UpdateCounter(name, value)
			}

			for name, expected := range tt.gauges {
				value, exists := storage.GetGauge(name)
				if !exists {
					t.Errorf("gauge %s not found", name)
				} else if value != expected {
					t.Errorf("gauge %s: expected %f, got %f", name, expected, value)
				}
			}

			for name, expected := range tt.counters {
				value, exists := storage.GetCounter(name)
				if !exists {
					t.Errorf("counter %s not found", name)
				} else if value != expected {
					t.Errorf("counter %s: expected %d, got %d", name, expected, value)
				}
			}
		})
	}
}

func TestGetAllGauges(t *testing.T) {
	tests := []struct {
		name     string
		gauges   map[string]float64
		expected int
	}{
		{"empty storage", map[string]float64{}, 0},
		{"single gauge", map[string]float64{"Alloc": 123.456}, 1},
		{"multiple gauges", map[string]float64{"Alloc": 100.0, "HeapAlloc": 200.0, "Sys": 300.0}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			for name, value := range tt.gauges {
				storage.UpdateGauge(name, value)
			}

			result := storage.GetAllGauges()

			if len(result) != tt.expected {
				t.Errorf("expected %d gauges, got %d", tt.expected, len(result))
			}

			for name, expectedValue := range tt.gauges {
				value, exists := result[name]
				if !exists {
					t.Errorf("gauge %s not found in result", name)
				} else if value != expectedValue {
					t.Errorf("gauge %s: expected %f, got %f", name, expectedValue, value)
				}
			}
		})
	}
}

func TestGetAllCounters(t *testing.T) {
	tests := []struct {
		name     string
		counters map[string]int64
		expected int
	}{
		{"empty storage", map[string]int64{}, 0},
		{"single counter", map[string]int64{"PollCount": 5}, 1},
		{"multiple counters", map[string]int64{"PollCount": 10, "Requests": 20, "Errors": 3}, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			for name, value := range tt.counters {
				storage.UpdateCounter(name, value)
			}

			result := storage.GetAllCounters()

			if len(result) != tt.expected {
				t.Errorf("expected %d counters, got %d", tt.expected, len(result))
			}

			for name, expectedValue := range tt.counters {
				value, exists := result[name]
				if !exists {
					t.Errorf("counter %s not found in result", name)
				} else if value != expectedValue {
					t.Errorf("counter %s: expected %d, got %d", name, expectedValue, value)
				}
			}
		})
	}
}

func TestStorageInterface(t *testing.T) {
	var _ Storage = (*MemStorage)(nil)
}

func TestSaveToFile(t *testing.T) {
	tests := []struct {
		name     string
		gauges   map[string]float64
		counters map[string]int64
	}{
		{
			name:     "empty storage",
			gauges:   map[string]float64{},
			counters: map[string]int64{},
		},
		{
			name:     "only gauges",
			gauges:   map[string]float64{"Alloc": 123.456, "HeapAlloc": 789.012},
			counters: map[string]int64{},
		},
		{
			name:     "only counters",
			gauges:   map[string]float64{},
			counters: map[string]int64{"PollCount": 5, "Requests": 10},
		},
		{
			name:     "mixed metrics",
			gauges:   map[string]float64{"Alloc": 100.0, "Sys": 200.0},
			counters: map[string]int64{"PollCount": 42, "Errors": 2},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage := NewMemStorage()

			for name, value := range tt.gauges {
				storage.UpdateGauge(name, value)
			}

			for name, value := range tt.counters {
				storage.UpdateCounter(name, value)
			}

			tmpFile := filepath.Join(t.TempDir(), "metrics.json")

			if err := storage.SaveToFile(tmpFile); err != nil {
				t.Fatalf("SaveToFile failed: %v", err)
			}

			data, err := os.ReadFile(tmpFile)
			if err != nil {
				t.Fatalf("failed to read saved file: %v", err)
			}

			var metrics []model.Metrics
			if err := json.Unmarshal(data, &metrics); err != nil {
				t.Fatalf("failed to unmarshal JSON: %v", err)
			}

			gaugeCount := 0
			counterCount := 0

			for _, m := range metrics {
				switch m.MType {
				case "gauge":
					gaugeCount++
					expectedValue, exists := tt.gauges[m.ID]
					if !exists {
						t.Errorf("unexpected gauge in file: %s", m.ID)
					} else if m.Value == nil || *m.Value != expectedValue {
						if m.Value == nil {
							t.Errorf("gauge %s has nil value", m.ID)
						} else {
							t.Errorf("gauge %s: expected %f, got %f", m.ID, expectedValue, *m.Value)
						}
					}
				case "counter":
					counterCount++
					expectedDelta, exists := tt.counters[m.ID]
					if !exists {
						t.Errorf("unexpected counter in file: %s", m.ID)
					} else if m.Delta == nil || *m.Delta != expectedDelta {
						if m.Delta == nil {
							t.Errorf("counter %s has nil delta", m.ID)
						} else {
							t.Errorf("counter %s: expected %d, got %d", m.ID, expectedDelta, *m.Delta)
						}
					}
				}
			}

			if gaugeCount != len(tt.gauges) {
				t.Errorf("expected %d gauges in file, got %d", len(tt.gauges), gaugeCount)
			}

			if counterCount != len(tt.counters) {
				t.Errorf("expected %d counters in file, got %d", len(tt.counters), counterCount)
			}
		})
	}
}

func TestLoadFromFile(t *testing.T) {
	tests := []struct {
		name    string
		metrics []model.Metrics
		wantErr bool
	}{
		{
			name: "valid metrics file",
			metrics: []model.Metrics{
				{ID: "Alloc", MType: "gauge", Value: ptrFloat64(123.456)},
				{ID: "HeapAlloc", MType: "gauge", Value: ptrFloat64(789.012)},
				{ID: "PollCount", MType: "counter", Delta: ptrInt64(42)},
			},
			wantErr: false,
		},
		{
			name:    "empty file",
			metrics: []model.Metrics{},
			wantErr: false,
		},
		{
			name: "only gauges",
			metrics: []model.Metrics{
				{ID: "Gauge1", MType: "gauge", Value: ptrFloat64(1.0)},
				{ID: "Gauge2", MType: "gauge", Value: ptrFloat64(2.0)},
			},
			wantErr: false,
		},
		{
			name: "only counters",
			metrics: []model.Metrics{
				{ID: "Counter1", MType: "counter", Delta: ptrInt64(10)},
				{ID: "Counter2", MType: "counter", Delta: ptrInt64(20)},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tmpFile := filepath.Join(t.TempDir(), "metrics.json")

			data, err := json.Marshal(tt.metrics)
			if err != nil {
				t.Fatalf("failed to marshal test data: %v", err)
			}

			if err := os.WriteFile(tmpFile, data, 0644); err != nil {
				t.Fatalf("failed to write test file: %v", err)
			}

			storage := NewMemStorage()

			err = storage.LoadFromFile(tmpFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadFromFile() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			for _, m := range tt.metrics {
				switch m.MType {
				case "gauge":
					if m.Value != nil {
						value, exists := storage.GetGauge(m.ID)
						if !exists {
							t.Errorf("gauge %s not loaded", m.ID)
						} else if value != *m.Value {
							t.Errorf("gauge %s: expected %f, got %f", m.ID, *m.Value, value)
						}
					}
				case "counter":
					if m.Delta != nil {
						delta, exists := storage.GetCounter(m.ID)
						if !exists {
							t.Errorf("counter %s not loaded", m.ID)
						} else if delta != *m.Delta {
							t.Errorf("counter %s: expected %d, got %d", m.ID, *m.Delta, delta)
						}
					}
				}
			}
		})
	}
}

func TestLoadFromFile_NonExistentFile(t *testing.T) {
	storage := NewMemStorage()
	err := storage.LoadFromFile("/nonexistent/path/metrics.json")
	if err != nil {
		t.Errorf("LoadFromFile should not error on non-existent file, got: %v", err)
	}
}

func TestLoadFromFile_InvalidJSON(t *testing.T) {
	tmpFile := filepath.Join(t.TempDir(), "invalid.json")
	if err := os.WriteFile(tmpFile, []byte("invalid json {{{"), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	storage := NewMemStorage()
	err := storage.LoadFromFile(tmpFile)
	if err == nil {
		t.Error("LoadFromFile should error on invalid JSON")
	}
}

func TestSaveAndLoadRoundTrip(t *testing.T) {
	storage1 := NewMemStorage()

	storage1.UpdateGauge("Alloc", 123.456)
	storage1.UpdateGauge("HeapAlloc", 789.012)
	storage1.UpdateGauge("Sys", 333.333)
	storage1.UpdateCounter("PollCount", 10)
	storage1.UpdateCounter("PollCount", 32)
	storage1.UpdateCounter("Requests", 100)

	tmpFile := filepath.Join(t.TempDir(), "metrics.json")

	if err := storage1.SaveToFile(tmpFile); err != nil {
		t.Fatalf("SaveToFile failed: %v", err)
	}

	storage2 := NewMemStorage()
	if err := storage2.LoadFromFile(tmpFile); err != nil {
		t.Fatalf("LoadFromFile failed: %v", err)
	}

	gauges1 := storage1.GetAllGauges()
	gauges2 := storage2.GetAllGauges()

	if len(gauges1) != len(gauges2) {
		t.Errorf("gauge count mismatch: expected %d, got %d", len(gauges1), len(gauges2))
	}

	for name, value1 := range gauges1 {
		value2, exists := gauges2[name]
		if !exists {
			t.Errorf("gauge %s not found after load", name)
		} else if value1 != value2 {
			t.Errorf("gauge %s: expected %f, got %f", name, value1, value2)
		}
	}

	counters1 := storage1.GetAllCounters()
	counters2 := storage2.GetAllCounters()

	if len(counters1) != len(counters2) {
		t.Errorf("counter count mismatch: expected %d, got %d", len(counters1), len(counters2))
	}

	for name, value1 := range counters1 {
		value2, exists := counters2[name]
		if !exists {
			t.Errorf("counter %s not found after load", name)
		} else if value1 != value2 {
			t.Errorf("counter %s: expected %d, got %d", name, value1, value2)
		}
	}
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}
