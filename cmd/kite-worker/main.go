package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gongahkia/kite/internal/config"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/queue"
	"github.com/gongahkia/kite/internal/scraper"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/internal/worker"
)

func main() {
	// Load configuration
	cfg, err := config.Load("")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger := observability.NewLogger(cfg.Observability.LogLevel, cfg.Observability.LogFormat)
	logger.Info("Starting Kite Worker v4.0.0")

	// Initialize metrics
	metrics := observability.NewMetrics()
	logger.Info("Metrics initialized")

	// Initialize storage
	var store storage.Storage
	switch cfg.Database.Driver {
	case "memory", "":
		store = storage.NewMemoryStorage()
		logger.Info("Using in-memory storage")
	case "sqlite":
		dbPath := cfg.Database.Database
		if dbPath == "" {
			dbPath = "kite.db"
		}
		store, err = storage.NewSQLiteStorage(dbPath)
		if err != nil {
			logger.Error("Failed to initialize SQLite storage", "error", err)
			os.Exit(1)
		}
		logger.Info("Using SQLite storage", "path", dbPath)
	default:
		logger.Error("Unsupported database driver", "driver", cfg.Database.Driver)
		os.Exit(1)
	}
	defer store.Close()

	// Initialize queue
	var q queue.Queue
	switch cfg.Queue.Driver {
	case "memory", "":
		q = queue.NewMemoryQueue()
		logger.Info("Using in-memory queue")
	case "redis":
		q, err = queue.NewRedisQueue(cfg.Redis.Host, cfg.Redis.Port, cfg.Redis.Password, cfg.Redis.DB)
		if err != nil {
			logger.Error("Failed to initialize Redis queue", "error", err)
			os.Exit(1)
		}
		logger.Info("Using Redis queue", "host", cfg.Redis.Host, "port", cfg.Redis.Port)
	case "nats":
		q, err = queue.NewNATSQueue(cfg.Queue.NATSUrl, cfg.Queue.StreamName)
		if err != nil {
			logger.Error("Failed to initialize NATS queue", "error", err)
			os.Exit(1)
		}
		logger.Info("Using NATS queue", "url", cfg.Queue.NATSUrl)
	default:
		logger.Error("Unsupported queue driver", "driver", cfg.Queue.Driver)
		os.Exit(1)
	}
	defer q.Close()

	// Create job handler
	handler := worker.NewJobHandler(store, logger, metrics)
	logger.Info("Job handler initialized")

	// Create worker pool
	workerCount := cfg.Worker.PoolSize
	if workerCount <= 0 {
		workerCount = 5
	}
	
	pool := worker.NewPool(workerCount, q, handler, logger, metrics)
	logger.Info("Worker pool created", "workers", workerCount)

	// Start worker pool
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if err := pool.Start(ctx); err != nil {
		logger.Error("Failed to start worker pool", "error", err)
		os.Exit(1)
	}

	logger.Info("Worker pool started successfully")

	// Start metrics server
	if cfg.Observability.MetricsEnabled {
		go func() {
			metricsAddr := fmt.Sprintf(":%d", cfg.Observability.MetricsPort)
			logger.Info("Starting metrics server", "address", metricsAddr)
			if err := observability.StartMetricsServer(metricsAddr); err != nil {
				logger.Error("Metrics server error", "error", err)
			}
		}()
	}

	// Wait for interrupt signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)
	
	sig := <-sigChan
	logger.Info("Received shutdown signal", "signal", sig.String())

	// Graceful shutdown
	logger.Info("Shutting down worker pool...")
	
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := pool.Stop(shutdownCtx); err != nil {
		logger.Error("Error during worker pool shutdown", "error", err)
	} else {
		logger.Info("Worker pool stopped gracefully")
	}

	logger.Info("Kite Worker shutdown complete")
}
