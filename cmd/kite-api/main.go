package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gongahkia/kite/internal/api"
	"github.com/gongahkia/kite/internal/config"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
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
	logger.Info("Starting Kite API server v4.0.0")

	// Initialize metrics
	metrics := observability.NewMetrics()
	logger.Info("Metrics initialized")

	// Initialize storage
	var store storage.Storage
	switch cfg.Database.Driver {
	case "memory", "":
		store = storage.NewMemoryStorage()
		logger.Info("Using in-memory storage")
	default:
		logger.Fatalf("Unsupported storage driver: %s", cfg.Database.Driver)
	}

	// Create API server
	server := api.NewServer(store, logger, metrics)
	server.SetupRoutes()

	// Start server in goroutine
	serverAddr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	go func() {
		logger.Infof("Starting HTTP server on %s", serverAddr)
		if err := server.Start(serverAddr); err != nil {
			logger.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer cancel()

	if err := server.Shutdown(); err != nil {
		logger.Errorf("Server forced to shutdown: %v", err)
	}

	// Close storage
	if err := store.Close(); err != nil {
		logger.Errorf("Failed to close storage: %v", err)
	}

	logger.Info("Server exited")

	// Wait for context timeout
	<-ctx.Done()
}
