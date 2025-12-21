package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/practicum-metrics/internal/agent"
	"github.com/user/practicum-metrics/internal/model"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger: ", err)
	}
	defer logger.Sync()

	pollInterval := time.Duration(flagPollInterval) * time.Second
	reportInterval := time.Duration(flagReportInterval) * time.Second
	serverAddress := "http://" + flagRunAddr

	collector := agent.NewMetricsCollector()
	sender := agent.NewMetricsSender(serverAddress, flagKey)

	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	collector.Collect()

	logger.Info("Agent started, collecting and sending metrics...")

	for {
		select {
		case <-stop:
			logger.Info("Shutting down agent gracefully...")
			return

		case <-pollTicker.C:
			collector.Collect()

		case <-reportTicker.C:
			gauges := collector.GetGauges()
			pollCount := collector.GetPollCount()

			var metrics []model.Metrics
			for name, value := range gauges {
				v := value
				metrics = append(metrics, model.Metrics{
					ID:    name,
					MType: "gauge",
					Value: &v,
				})
			}

			delta := pollCount
			metrics = append(metrics, model.Metrics{
				ID:    agent.MetricPollCount,
				MType: "counter",
				Delta: &delta,
			})

			if err := sender.SendMetricsBatch(metrics); err != nil {
				logger.Error("Failed to send metrics batch", zap.Error(err))
			}
		}
	}
}
