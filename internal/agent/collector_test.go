package agent

import (
	"testing"
)

func TestNewMetricsCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"create new collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMetricsCollector()
			if collector == nil {
				t.Fatal("NewMetricsCollector returned nil")
			}

			if collector.gauges == nil {
				t.Error("gauges map not initialized")
			}

			if collector.pollCount != 0 {
				t.Errorf("expected pollCount to be 0, got %d", collector.pollCount)
			}
		})
	}
}

func TestCollect(t *testing.T) {
	tests := []struct {
		name              string
		collectTimes      int
		expectedPollCount int64
		expectedGauges    []string
	}{
		{
			name:              "single collect",
			collectTimes:      1,
			expectedPollCount: 1,
			expectedGauges: []string{
				"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
				"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
				"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
				"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
				"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
				"Sys", "TotalAlloc", "RandomValue",
			},
		},
		{
			name:              "double collect",
			collectTimes:      2,
			expectedPollCount: 2,
			expectedGauges: []string{
				"Alloc", "BuckHashSys", "Frees", "GCCPUFraction", "GCSys",
				"HeapAlloc", "HeapIdle", "HeapInuse", "HeapObjects", "HeapReleased",
				"HeapSys", "LastGC", "Lookups", "MCacheInuse", "MCacheSys",
				"MSpanInuse", "MSpanSys", "Mallocs", "NextGC", "NumForcedGC",
				"NumGC", "OtherSys", "PauseTotalNs", "StackInuse", "StackSys",
				"Sys", "TotalAlloc", "RandomValue",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMetricsCollector()

			for i := 0; i < tt.collectTimes; i++ {
				collector.Collect()
			}

			if collector.GetPollCount() != tt.expectedPollCount {
				t.Errorf("expected pollCount to be %d, got %d", tt.expectedPollCount, collector.GetPollCount())
			}

			gauges := collector.GetGauges()

			if len(gauges) != len(tt.expectedGauges) {
				t.Errorf("expected %d gauges, got %d", len(tt.expectedGauges), len(gauges))
			}

			for _, name := range tt.expectedGauges {
				if _, exists := gauges[name]; !exists {
					t.Errorf("gauge %s not found", name)
				}
			}
		})
	}
}

func TestGetGauges(t *testing.T) {
	tests := []struct {
		name                string
		collectBeforeGet    bool
		shouldContainRandom bool
		checkNonNegative    bool
	}{
		{
			name:                "after collect",
			collectBeforeGet:    true,
			shouldContainRandom: true,
			checkNonNegative:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMetricsCollector()

			if tt.collectBeforeGet {
				collector.Collect()
			}

			gauges := collector.GetGauges()

			if gauges == nil {
				t.Fatal("GetGauges returned nil")
			}

			if tt.shouldContainRandom {
				if _, exists := gauges["RandomValue"]; !exists {
					t.Error("RandomValue not found in gauges")
				}
			}

			if tt.checkNonNegative {
				for name, value := range gauges {
					if value < 0 {
						t.Errorf("gauge %s has negative value: %f", name, value)
					}
				}
			}
		})
	}
}

func TestGetPollCount(t *testing.T) {
	tests := []struct {
		name         string
		collectTimes int
		expected     int64
	}{
		{"initial state", 0, 0},
		{"after one collect", 1, 1},
		{"after three collects", 3, 3},
		{"after five collects", 5, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMetricsCollector()

			for i := 0; i < tt.collectTimes; i++ {
				collector.Collect()
			}

			if collector.GetPollCount() != tt.expected {
				t.Errorf("expected pollCount to be %d, got %d", tt.expected, collector.GetPollCount())
			}
		})
	}
}

func TestRandomValueChanges(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"random value should be in range"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewMetricsCollector()

			collector.Collect()
			gauges1 := collector.GetGauges()
			value1 := gauges1["RandomValue"]

			collector.Collect()
			gauges2 := collector.GetGauges()
			value2 := gauges2["RandomValue"]

			if value1 == value2 {
				t.Log("Warning: RandomValue did not change (this is rare but possible)")
			}

			if value1 < 0 || value1 >= 1 {
				t.Errorf("RandomValue out of range: %f", value1)
			}
			if value2 < 0 || value2 >= 1 {
				t.Errorf("RandomValue out of range: %f", value2)
			}
		})
	}
}
