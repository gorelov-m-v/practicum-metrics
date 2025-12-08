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
	"github.com/user/practicum-metrics/internal/database"
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

	var db *database.DB
	var store storage.Storage
	var persister *storage.Persister

	// Priority: PostgreSQL -> File -> Memory
	if flagDatabaseDSN != "" {
		// Use PostgreSQL storage
		db, err = database.NewDB(flagDatabaseDSN)
		if err != nil {
			logger.Fatal("Failed to connect to database", zap.Error(err))
		}
		defer db.Close()
		logger.Info("Database connection established")

		// Run migrations
		if err := db.RunMigrations("internal/database/migrations"); err != nil {
			logger.Fatal("Failed to run migrations", zap.Error(err))
		}
		logger.Info("Database migrations completed")

		store = storage.NewDBStorage(db.GetConn())
		logger.Info("Using PostgreSQL storage")
	} else if flagFileStoragePath != "" {
		// Use file-backed memory storage
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
		// Use in-memory storage only
		store = storage.NewMemStorage()
		logger.Info("Using in-memory storage")
	}

	metricsService := service.NewMetricsService(store)
	h, err := handler.NewMetricHandler(metricsService, persister, db)
	if err != nil {
		logger.Fatal("Failed to create handler", zap.Error(err))
	}

	r := chi.NewRouter()

	r.Use(middleware.GzipDecompress)
	r.Use(middleware.GzipCompress)
	r.Use(middleware.Logging(logger))
	r.Use(chiMiddleware.StripSlashes)

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
