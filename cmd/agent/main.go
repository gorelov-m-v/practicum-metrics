package main

import (
	"fmt"
	"time"

	"github.com/user/practicum-metrics/internal/agent"
)

const (
	pollInterval   = 2 * time.Second
	reportInterval = 10 * time.Second
	serverAddress  = "http://localhost:8080"
)

func main() {
	collector := agent.NewMetricsCollector()
	sender := agent.NewMetricsSender(serverAddress)

	pollTicker := time.NewTicker(pollInterval)
	defer pollTicker.Stop()

	reportTicker := time.NewTicker(reportInterval)
	defer reportTicker.Stop()

	collector.Collect()

	for {
		select {
		case <-pollTicker.C:
			collector.Collect()

		case <-reportTicker.C:
			gauges := collector.GetGauges()
			pollCount := collector.GetPollCount()

			for name, value := range gauges {
				if err := sender.SendGauge(name, value); err != nil {
					fmt.Printf("Failed to send gauge %s: %v\n", name, err)
				}
			}

			if err := sender.SendCounter(agent.MetricPollCount, pollCount); err != nil {
				fmt.Printf("Failed to send counter %s: %v\n", agent.MetricPollCount, err)
			}
		}
	}
}
