package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"go.uber.org/zap"

	"github.com/user/practicum-metrics/internal/agent"
	"github.com/user/practicum-metrics/internal/encryption"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func printBuildInfo() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)
}

func main() {
	printBuildInfo()
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
	if flagCryptoKey != "" {
		publicKey, err := encryption.LoadPublicKey(flagCryptoKey)
		if err != nil {
			logger.Fatal("Failed to load public key", zap.Error(err))
		}
		sender.SetPublicKey(publicKey)
	}

	workerPool := agent.NewWorkerPool(flagRateLimit, sender, logger)
	workerPool.Start()

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	runtimeCollector.Collect()
	if err := gopsutilCollector.Collect(); err != nil {
		logger.Error("Failed to collect gopsutil metrics", zap.Error(err))
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		pollTicker := time.NewTicker(pollInterval)
		defer pollTicker.Stop()

		for {
			select {
			case <-ctx.Done():
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

		for {
			select {
			case <-ctx.Done():
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
			case <-ctx.Done():
				logger.Info("Metrics sender shutting down...")
				return
			case <-reportTicker.C:
				metrics := agent.BuildMetrics(
					runtimeCollector.GetGauges(),
					gopsutilCollector.GetGauges(),
					runtimeCollector.GetPollCount(),
				)
				workerPool.Submit(agent.MetricTask{Metrics: metrics})
			}
		}
	}()

	logger.Info("Agent started with worker pool", zap.Int("workers", flagRateLimit))

	<-ctx.Done()
	logger.Info("Shutting down agent gracefully...")
	stop()

	wg.Wait()

	finalMetrics := agent.BuildMetrics(
		runtimeCollector.GetGauges(),
		gopsutilCollector.GetGauges(),
		runtimeCollector.GetPollCount(),
	)
	workerPool.Submit(agent.MetricTask{Metrics: finalMetrics})
	workerPool.Stop()

	logger.Info("Agent stopped")
}
