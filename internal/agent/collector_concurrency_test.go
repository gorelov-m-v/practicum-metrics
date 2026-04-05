package agent

import (
	"sync"
	"testing"
)

func TestMetricsCollector_ConcurrentCollect(t *testing.T) {
	collector := NewMetricsCollector()
	iterations := 100
	goroutines := 10

	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				collector.Collect()
			}
		}()
	}

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				_ = collector.GetGauges()
				_ = collector.GetPollCount()
			}
		}()
	}

	wg.Wait()

	expectedPollCount := int64(goroutines * iterations)
	actualPollCount := collector.GetPollCount()

	if actualPollCount != expectedPollCount {
		t.Errorf("expected pollCount to be %d, got %d", expectedPollCount, actualPollCount)
	}

	gauges := collector.GetGauges()
	if len(gauges) == 0 {
		t.Error("gauges map is empty after concurrent operations")
	}
}

func TestMetricsCollector_ConcurrentGetGauges(t *testing.T) {
	collector := NewMetricsCollector()
	collector.Collect()

	goroutines := 50
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			gauges := collector.GetGauges()
			if len(gauges) == 0 {
				t.Error("GetGauges returned empty map")
			}
		}()
	}

	wg.Wait()
}

func TestMetricsCollector_ConcurrentGetPollCount(t *testing.T) {
	collector := NewMetricsCollector()
	collector.Collect()

	goroutines := 50
	var wg sync.WaitGroup

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			count := collector.GetPollCount()
			if count < 1 {
				t.Errorf("expected pollCount >= 1, got %d", count)
			}
		}()
	}

	wg.Wait()
}
