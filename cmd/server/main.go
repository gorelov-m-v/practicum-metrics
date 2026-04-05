package main

import (
	"context"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"go.uber.org/zap"

	"github.com/user/practicum-metrics/internal/audit"
	"github.com/user/practicum-metrics/internal/database"
	"github.com/user/practicum-metrics/internal/handler"
	"github.com/user/practicum-metrics/internal/middleware"
	"github.com/user/practicum-metrics/internal/repository"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
)

func main() {
	parseFlags()

	logger, err := zap.NewProduction()
	if err != nil {
		log.Fatal("Failed to initialize logger: ", err)
	}
	defer logger.Sync()

	var db *database.DB
	var store storage.Storage
	var persister *storage.Persister
	var txManager database.TransactionManager
	var gaugeRepo repository.GaugeRepository
	var counterRepo repository.CounterRepository

	if flagDatabaseDSN != "" {
		db, err = database.NewDB(flagDatabaseDSN)
		if err != nil {
			logger.Fatal("Failed to connect to database", zap.Error(err))
		}
		defer db.Close()
		logger.Info("Database connection established")

		if err := db.RunMigrations("internal/database/migrations"); err != nil {
			logger.Fatal("Failed to run migrations", zap.Error(err))
		}
		logger.Info("Database migrations completed")

		store = storage.NewDBStorage(db.GetConn())
		txManager = database.NewTransactionManager(db.GetConn())
		gaugeRepo = repository.NewGaugeRepository(db.GetConn())
		counterRepo = repository.NewCounterRepository(db.GetConn())
		logger.Info("Using PostgreSQL storage")
	} else if flagFileStoragePath != "" {
		store = storage.NewMemStorage()
		persister = storage.NewPersister(store, flagFileStoragePath, flagStoreInterval, logger)

		if flagRestore {
			if err := persister.Restore(); err != nil {
				logger.Warn("Failed to restore metrics from file", zap.Error(err))
			}
		}

		persister.Start()
		defer persister.Stop()
		logger.Info("Using file-backed memory storage", zap.String("path", flagFileStoragePath))
	} else {
		store = storage.NewMemStorage()
		logger.Info("Using in-memory storage")
	}

	metricsService := service.NewMetricsService(store, txManager, gaugeRepo, counterRepo)

	auditPublisher := audit.NewPublisher()
	if flagAuditFile != "" {
		auditPublisher.Subscribe(audit.NewFileListener(flagAuditFile))
		logger.Info("Audit file listener enabled", zap.String("path", flagAuditFile))
	}
	if flagAuditURL != "" {
		auditPublisher.Subscribe(audit.NewURLListener(flagAuditURL))
		logger.Info("Audit URL listener enabled", zap.String("url", flagAuditURL))
	}

	h, err := handler.NewMetricHandler(metricsService, persister, db, auditPublisher)
	if err != nil {
		logger.Fatal("Failed to create handler", zap.Error(err))
	}

	r := chi.NewRouter()

	r.Use(middleware.GzipDecompress)
	r.Use(middleware.HashVerify(flagKey))
	r.Use(middleware.GzipCompress)
	r.Use(middleware.Logging(logger))
	r.Use(chiMiddleware.StripSlashes)

	r.Mount("/debug", http.DefaultServeMux)

	r.Post("/updates", h.UpdateMetricsBatch)
	r.Post("/update", h.UpdateMetricJSON)
	r.Post("/value", h.GetMetricJSON)
	r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	r.Get("/value/{type}/{name}", h.GetMetric)
	r.Get("/", h.ListMetrics)
	r.Get("/ping", h.PingDB)

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

	if persister != nil {
		persister.SaveSync()
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	logger.Info("Server stopped gracefully")
}
