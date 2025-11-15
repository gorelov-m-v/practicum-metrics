package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/user/practicum-metrics/internal/agent"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	logger, err := zap.NewProduction()
	if err != nil {
		panic("Failed to initialize logger: " + err.Error())
	}
	defer logger.Sync()

	pollInterval := time.Duration(flagPollInterval) * time.Second
	reportInterval := time.Duration(flagReportInterval) * time.Second
	serverAddress := "http://" + flagRunAddr

	collector := agent.NewMetricsCollector()
	sender := agent.NewMetricsSender(serverAddress)

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

			for name, value := range gauges {
				if err := sender.SendGaugeJSON(name, value); err != nil {
					logger.Error("Failed to send gauge", zap.String("metric", name), zap.Error(err))
				}
			}

			if err := sender.SendCounterJSON(agent.MetricPollCount, pollCount); err != nil {
				logger.Error("Failed to send counter", zap.String("metric", agent.MetricPollCount), zap.Error(err))
			}
		}
	}
}
