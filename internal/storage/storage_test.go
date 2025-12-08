package storage

import (
	"context"
	"testing"
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
			ctx := context.Background()
			storage := NewMemStorage()
			storage.UpdateGauge(ctx, tt.key, tt.value)

			value, exists := storage.GetGauge(ctx, tt.key)

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
			ctx := context.Background()
			storage := NewMemStorage()

			storage.UpdateGauge(ctx, tt.key, tt.firstValue)
			value, _ := storage.GetGauge(ctx, tt.key)
			if value != tt.firstValue {
				t.Errorf("expected initial value %f, got %f", tt.firstValue, value)
			}

			storage.UpdateGauge(ctx, tt.key, tt.finalValue)
			value, exists := storage.GetGauge(ctx, tt.key)

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
			ctx := context.Background()
			storage := NewMemStorage()

			for _, value := range tt.values {
				storage.UpdateCounter(ctx, tt.key, value)
			}

			total, exists := storage.GetCounter(ctx, tt.key)

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
			ctx := context.Background()
			storage := NewMemStorage()

			for i := 0; i < tt.updates; i++ {
				storage.UpdateCounter(ctx, tt.key, 1)
			}

			count, exists := storage.GetCounter(ctx, tt.key)

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
			ctx := context.Background()
			storage := NewMemStorage()

			for name, value := range tt.gauges {
				storage.UpdateGauge(ctx, name, value)
			}

			for name, value := range tt.counters {
				storage.UpdateCounter(ctx, name, value)
			}

			for name, expected := range tt.gauges {
				value, exists := storage.GetGauge(ctx, name)
				if !exists {
					t.Errorf("gauge %s not found", name)
				} else if value != expected {
					t.Errorf("gauge %s: expected %f, got %f", name, expected, value)
				}
			}

			for name, expected := range tt.counters {
				value, exists := storage.GetCounter(ctx, name)
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
			ctx := context.Background()
			storage := NewMemStorage()

			for name, value := range tt.gauges {
				storage.UpdateGauge(ctx, name, value)
			}

			result := storage.GetAllGauges(ctx)

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
			ctx := context.Background()
			storage := NewMemStorage()

			for name, value := range tt.counters {
				storage.UpdateCounter(ctx, name, value)
			}

			result := storage.GetAllCounters(ctx)

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

func TestSetGauges(t *testing.T) {
	tests := []struct {
		name   string
		gauges map[string]float64
	}{
		{
			name:   "empty map",
			gauges: map[string]float64{},
		},
		{
			name:   "single gauge",
			gauges: map[string]float64{"Alloc": 123.456},
		},
		{
			name:   "multiple gauges",
			gauges: map[string]float64{"Alloc": 123.456, "HeapAlloc": 789.012, "Sys": 500.0},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := NewMemStorage()
			storage.SetGauges(ctx, tt.gauges)

			for name, expected := range tt.gauges {
				value, exists := storage.GetGauge(ctx, name)
				if !exists {
					t.Errorf("gauge %s not set", name)
				} else if value != expected {
					t.Errorf("gauge %s: expected %f, got %f", name, expected, value)
				}
			}
		})
	}
}

func TestSetCounters(t *testing.T) {
	tests := []struct {
		name     string
		counters map[string]int64
	}{
		{
			name:     "empty map",
			counters: map[string]int64{},
		},
		{
			name:     "single counter",
			counters: map[string]int64{"PollCount": 42},
		},
		{
			name:     "multiple counters",
			counters: map[string]int64{"PollCount": 42, "Requests": 100, "Errors": 5},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			storage := NewMemStorage()
			storage.SetCounters(ctx, tt.counters)

			for name, expected := range tt.counters {
				value, exists := storage.GetCounter(ctx, name)
				if !exists {
					t.Errorf("counter %s not set", name)
				} else if value != expected {
					t.Errorf("counter %s: expected %d, got %d", name, expected, value)
				}
			}
		})
	}
}

func TestSetGauges_Overwrite(t *testing.T) {
	ctx := context.Background()
	storage := NewMemStorage()
	storage.UpdateGauge(ctx, "Alloc", 100.0)

	storage.SetGauges(ctx, map[string]float64{"Alloc": 200.0, "Sys": 300.0})

	value, _ := storage.GetGauge(ctx, "Alloc")
	if value != 200.0 {
		t.Errorf("expected Alloc to be overwritten to 200.0, got %f", value)
	}

	sysValue, exists := storage.GetGauge(ctx, "Sys")
	if !exists {
		t.Error("expected Sys to be set")
	} else if sysValue != 300.0 {
		t.Errorf("expected Sys to be 300.0, got %f", sysValue)
	}
}

func TestSetCounters_Overwrite(t *testing.T) {
	ctx := context.Background()
	storage := NewMemStorage()
	storage.UpdateCounter(ctx, "PollCount", 10)

	storage.SetCounters(ctx, map[string]int64{"PollCount": 50, "Requests": 100})

	value, _ := storage.GetCounter(ctx, "PollCount")
	if value != 50 {
		t.Errorf("expected PollCount to be overwritten to 50, got %d", value)
	}

	requestsValue, exists := storage.GetCounter(ctx, "Requests")
	if !exists {
		t.Error("expected Requests to be set")
	} else if requestsValue != 100 {
		t.Errorf("expected Requests to be 100, got %d", requestsValue)
	}
}

func ptrFloat64(v float64) *float64 {
	return &v
}

func ptrInt64(v int64) *int64 {
	return &v
}
