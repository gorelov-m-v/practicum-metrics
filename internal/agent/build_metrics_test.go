package agent

import "testing"

func TestBuildMetrics(t *testing.T) {
	runtimeGauges := map[string]float64{
		"Alloc":       1,
		"RandomValue": 2,
	}
	systemGauges := map[string]float64{
		"TotalMemory": 3,
	}

	metrics := BuildMetrics(runtimeGauges, systemGauges, 4)

	if len(metrics) != 4 {
		t.Fatalf("expected 4 metrics, got %d", len(metrics))
	}

	found := map[string]bool{}
	for _, metric := range metrics {
		found[metric.ID] = true
	}

	for _, name := range []string{"Alloc", "RandomValue", "TotalMemory", MetricPollCount} {
		if !found[name] {
			t.Fatalf("metric %s not found", name)
		}
	}
}
