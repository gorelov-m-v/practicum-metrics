package agent

import "github.com/user/practicum-metrics/internal/model"

func BuildMetrics(runtimeGauges, systemGauges map[string]float64, pollCount int64) []model.Metrics {
	metrics := make([]model.Metrics, 0, len(runtimeGauges)+len(systemGauges)+1)

	for name, value := range runtimeGauges {
		v := value
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &v,
		})
	}

	for name, value := range systemGauges {
		v := value
		metrics = append(metrics, model.Metrics{
			ID:    name,
			MType: "gauge",
			Value: &v,
		})
	}

	delta := pollCount
	metrics = append(metrics, model.Metrics{
		ID:    MetricPollCount,
		MType: "counter",
		Delta: &delta,
	})

	return metrics
}
