package agent

import (
	"sync"
	"testing"
)

func TestNewGopsutilCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{"create new gopsutil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewGopsutilCollector()
			if collector == nil {
				t.Fatal("NewGopsutilCollector returned nil")
			}

			if collector.gauges == nil {
				t.Error("gauges map not initialized")
			}

			if collector.cpuNum < 1 {
				t.Error("cpuNum should be at least 1")
			}
		})
	}
}

func TestGopsutilCollector_Collect(t *testing.T) {
	tests := []struct {
		name           string
		collectTimes   int
		expectedGauges []string
	}{
		{
			name:         "single collect",
			collectTimes: 1,
			expectedGauges: []string{
				MetricTotalMemory,
				MetricFreeMemory,
			},
		},
		{
			name:         "double collect",
			collectTimes: 2,
			expectedGauges: []string{
				MetricTotalMemory,
				MetricFreeMemory,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewGopsutilCollector()

			for i := 0; i < tt.collectTimes; i++ {
				err := collector.Collect()
				if err != nil {
					t.Fatalf("Collect() returned error: %v", err)
				}
			}

			gauges := collector.GetGauges()

			for _, metricName := range tt.expectedGauges {
				if _, exists := gauges[metricName]; !exists {
					t.Errorf("gauge %s not found", metricName)
				}
			}

			cpuMetricsFound := 0
			for name := range gauges {
				if len(name) > len(MetricCPUutilization1) && name[:len(MetricCPUutilization1)] == MetricCPUutilization1 {
					cpuMetricsFound++
				}
			}

			if cpuMetricsFound < 1 {
				t.Error("no CPU utilization metrics found")
			}
		})
	}
}

func TestGopsutilCollector_GetGauges(t *testing.T) {
	tests := []struct {
		name             string
		collectBeforeGet bool
		checkPositive    bool
	}{
		{
			name:             "after collect",
			collectBeforeGet: true,
			checkPositive:    true,
		},
		{
			name:             "without collect",
			collectBeforeGet: false,
			checkPositive:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := NewGopsutilCollector()

			if tt.collectBeforeGet {
				if err := collector.Collect(); err != nil {
					t.Fatalf("Collect() returned error: %v", err)
				}
			}

			gauges := collector.GetGauges()

			if gauges == nil {
				t.Fatal("GetGauges returned nil")
			}

			if tt.checkPositive {
				if totalMem, exists := gauges[MetricTotalMemory]; exists {
					if totalMem <= 0 {
						t.Errorf("TotalMemory should be positive, got %f", totalMem)
					}
				}

				if freeMem, exists := gauges[MetricFreeMemory]; exists {
					if freeMem < 0 {
						t.Errorf("FreeMemory should be non-negative, got %f", freeMem)
					}
				}

				for name, value := range gauges {
					if len(name) > len(MetricCPUutilization1) && name[:len(MetricCPUutilization1)] == MetricCPUutilization1 {
						if value < 0 || value > 100 {
							t.Errorf("CPU utilization %s out of range [0, 100]: %f", name, value)
						}
					}
				}
			}
		})
	}
}

func TestGopsutilCollector_ConcurrentCollect(t *testing.T) {
	collector := NewGopsutilCollector()
	iterations := 10
	goroutines := 5

	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				err := collector.Collect()
				if err != nil {
					t.Errorf("Collect() returned error: %v", err)
				}
			}
		}()
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = collector.GetGauges()
			}
		}()
	}

	wg.Wait()

	gauges := collector.GetGauges()
	if len(gauges) == 0 {
		t.Error("gauges map is empty after concurrent operations")
	}
}

func TestGopsutilCollector_MetricsValidity(t *testing.T) {
	collector := NewGopsutilCollector()

	err := collector.Collect()
	if err != nil {
		t.Fatalf("Collect() returned error: %v", err)
	}

	gauges := collector.GetGauges()

	if totalMem, exists := gauges[MetricTotalMemory]; exists {
		if totalMem < 1024*1024 {
			t.Errorf("TotalMemory seems too low: %f bytes", totalMem)
		}
	} else {
		t.Error("TotalMemory metric not found")
	}

	totalMem := gauges[MetricTotalMemory]
	freeMem := gauges[MetricFreeMemory]
	if freeMem > totalMem {
		t.Errorf("FreeMemory (%f) should not exceed TotalMemory (%f)", freeMem, totalMem)
	}
}
