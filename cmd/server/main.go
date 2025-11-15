package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/user/practicum-metrics/internal/handler"
	"github.com/user/practicum-metrics/internal/middleware"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger: ", err)
	}
	defer logger.Sync()

	store := storage.NewMemStorage()

	if flagRestore {
		if err := store.LoadFromFile(flagFileStoragePath); err != nil {
			logger.Warn("Failed to restore metrics from file", zap.Error(err))
		} else {
			logger.Info("Metrics restored from file")
		}
	}

	persister := storage.NewPersister(store, flagFileStoragePath, flagStoreInterval, logger)
	persister.Start()
	defer persister.Stop()

	metricsService := service.NewMetricsServiceWithPersister(store, persister)
	h, err := handler.NewMetricHandler(metricsService)
	if err != nil {
		logger.Fatal("Failed to create handler", zap.Error(err))
	}

	r := chi.NewRouter()

	r.Use(middleware.GzipDecompress)
	r.Use(middleware.GzipCompress)
	r.Use(middleware.Logging(logger))
	r.Use(chiMiddleware.StripSlashes)

	r.Post("/update", h.UpdateMetricJSON)
	r.Post("/value", h.GetMetricJSON)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	r.Get("/value/{type}/{name}", h.GetMetric)
	r.Get("/", h.ListMetrics)

	server := &http.Server{
		Addr:    flagRunAddr,
		Handler: r,
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	go func() {
		logger.Info("Starting server", zap.String("address", flagRunAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	<-stop
	logger.Info("Shutting down server...")

	persister.SaveSync()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}
