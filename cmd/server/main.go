package main

import (
	"context"
	"fmt"
	"log"
	"net"
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
	"github.com/user/practicum-metrics/internal/encryption"
	"github.com/user/practicum-metrics/internal/grpcserver"
	"github.com/user/practicum-metrics/internal/handler"
	"github.com/user/practicum-metrics/internal/middleware"
	"github.com/user/practicum-metrics/internal/repository"
	"github.com/user/practicum-metrics/internal/service"
	"github.com/user/practicum-metrics/internal/storage"
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
		logger.Info("Using file-backed memory storage", zap.String("path", flagFileStoragePath))
	} else {
		store = storage.NewMemStorage()
		logger.Info("Using in-memory storage")
	}

	metricsService := service.NewMetricsService(store, txManager, gaugeRepo, counterRepo)

	auditPublisher := audit.NewPublisher()
	if flagAuditFile != "" {
		fl, err := audit.NewFileListener(flagAuditFile)
		if err != nil {
			logger.Fatal("Failed to create audit file listener", zap.Error(err))
		}
		defer fl.Close()
		auditPublisher.Subscribe(fl)
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

	decryptMiddleware := middleware.CryptoDecrypt(nil)
	if flagCryptoKey != "" {
		privateKey, err := encryption.LoadPrivateKey(flagCryptoKey)
		if err != nil {
			logger.Fatal("Failed to load private key", zap.Error(err))
		}
		decryptMiddleware = middleware.CryptoDecrypt(privateKey)
	}

	var trustedSubnet *net.IPNet
	if flagTrustedSubnet != "" {
		_, ipNet, err := net.ParseCIDR(flagTrustedSubnet)
		if err != nil {
			logger.Fatal("Failed to parse trusted subnet", zap.String("trusted_subnet", flagTrustedSubnet), zap.Error(err))
		}
		trustedSubnet = ipNet
		logger.Info("Trusted subnet enabled", zap.String("trusted_subnet", flagTrustedSubnet))
	}

	serverErrCh := make(chan error, 2)

	var metricsGRPCServer interface {
		Serve(net.Listener) error
		GracefulStop()
	}
	if flagGRPCAddr != "" {
		listener, err := net.Listen("tcp", flagGRPCAddr)
		if err != nil {
			logger.Fatal("Failed to listen gRPC address", zap.String("address", flagGRPCAddr), zap.Error(err))
		}

		grpcSrv := grpcserver.NewServer(metricsService, persister, trustedSubnet)
		metricsGRPCServer = grpcSrv

		go func() {
			logger.Info("Starting gRPC server", zap.String("address", flagGRPCAddr))
			if err := grpcSrv.Serve(listener); err != nil {
				serverErrCh <- fmt.Errorf("gRPC server failed: %w", err)
			}
		}()
	}

	r := chi.NewRouter()

	r.Use(decryptMiddleware)
	r.Use(middleware.GzipDecompress)
	r.Use(middleware.HashVerify(flagKey))
	r.Use(middleware.GzipCompress)
	r.Use(middleware.Logging(logger))
	r.Use(chiMiddleware.StripSlashes)

	r.Mount("/debug", http.DefaultServeMux)

	if trustedSubnet != nil {
		r.With(middleware.TrustedSubnet(trustedSubnet)).Post("/updates", h.UpdateMetricsBatch)
		r.With(middleware.TrustedSubnet(trustedSubnet)).Post("/update", h.UpdateMetricJSON)
		r.With(middleware.TrustedSubnet(trustedSubnet)).Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	} else {
		r.Post("/updates", h.UpdateMetricsBatch)
		r.Post("/update", h.UpdateMetricJSON)
		r.Post("/update/{type}/{name}/{value}", h.UpdateMetric)
	}
	r.Post("/value", h.GetMetricJSON)
	r.Get("/value/{type}/{name}", h.GetMetric)
	r.Get("/", h.ListMetrics)
	r.Get("/ping", h.PingDB)

	server := &http.Server{
		Addr:    flagRunAddr,
		Handler: r,
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	go func() {
		logger.Info("Starting server", zap.String("address", flagRunAddr))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			serverErrCh <- fmt.Errorf("HTTP server failed: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-serverErrCh:
		logger.Error("Server failed", zap.Error(err))
		stop()
	}
	logger.Info("Shutting down server...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Fatal("Server forced to shutdown", zap.Error(err))
	}

	if metricsGRPCServer != nil {
		metricsGRPCServer.GracefulStop()
	}

	if persister != nil {
		persister.Stop()
		persister.SaveSync()
	}

	logger.Info("Server stopped gracefully")
}
