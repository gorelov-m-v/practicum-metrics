package main

import (
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/user/practicum-metrics/internal/agent"
	"github.com/user/practicum-metrics/internal/model"
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

	runtimeCollector := agent.NewMetricsCollector()
	gopsutilCollector := agent.NewGopsutilCollector()
	sender := agent.NewMetricsSender(serverAddress, flagKey)

	workerPool := agent.NewWorkerPool(flagRateLimit, sender, logger)
	workerPool.Start()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		pollTicker := time.NewTicker(pollInterval)
		defer pollTicker.Stop()

		runtimeCollector.Collect()

		for {
			select {
			case <-stop:
				logger.Info("Runtime collector shutting down...")
				return
			case <-pollTicker.C:
				runtimeCollector.Collect()
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		pollTicker := time.NewTicker(pollInterval)
		defer pollTicker.Stop()

		if err := gopsutilCollector.Collect(); err != nil {
			logger.Error("Failed to collect gopsutil metrics", zap.Error(err))
		}

		for {
			select {
			case <-stop:
				logger.Info("Gopsutil collector shutting down...")
				return
			case <-pollTicker.C:
				if err := gopsutilCollector.Collect(); err != nil {
					logger.Error("Failed to collect gopsutil metrics", zap.Error(err))
				}
			}
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		reportTicker := time.NewTicker(reportInterval)
		defer reportTicker.Stop()

		for {
			select {
			case <-stop:
				logger.Info("Metrics sender shutting down...")
				return
			case <-reportTicker.C:
				var metrics []model.Metrics

				runtimeGauges := runtimeCollector.GetGauges()
				for name, value := range runtimeGauges {
					v := value
					metrics = append(metrics, model.Metrics{
						ID:    name,
						MType: "gauge",
						Value: &v,
					})
				}

				gopsutilGauges := gopsutilCollector.GetGauges()
				for name, value := range gopsutilGauges {
					v := value
					metrics = append(metrics, model.Metrics{
						ID:    name,
						MType: "gauge",
						Value: &v,
					})
				}

				pollCount := runtimeCollector.GetPollCount()
				delta := pollCount
				metrics = append(metrics, model.Metrics{
					ID:    agent.MetricPollCount,
					MType: "counter",
					Delta: &delta,
				})

				workerPool.Submit(agent.MetricTask{Metrics: metrics})
			}
		}
	}()

	logger.Info("Agent started with worker pool", zap.Int("workers", flagRateLimit))

	<-stop
	logger.Info("Shutting down agent gracefully...")

	workerPool.Stop()
	wg.Wait()

	logger.Info("Agent stopped")
}
