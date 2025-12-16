package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gongahkia/kite/internal/api"
	"github.com/gongahkia/kite/internal/api/middleware"
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
	var err error

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
			logger.Fatalf("Failed to initialize SQLite storage: %v", err)
		}
		logger.Infof("Using SQLite storage: %s", dbPath)

	case "postgres", "postgresql":
		connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.Username,
			cfg.Database.Password, cfg.Database.Database, cfg.Database.SSLMode)
		store, err = storage.NewPostgresStorage(connStr)
		if err != nil {
			logger.Fatalf("Failed to initialize PostgreSQL storage: %v", err)
		}
		logger.Infof("Using PostgreSQL storage: %s@%s:%d/%s",
			cfg.Database.Username, cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)

	case "mongodb", "mongo":
		uri := fmt.Sprintf("mongodb://%s:%s@%s:%d",
			cfg.Database.Username, cfg.Database.Password, cfg.Database.Host, cfg.Database.Port)
		if cfg.Database.Username == "" {
			uri = fmt.Sprintf("mongodb://%s:%d", cfg.Database.Host, cfg.Database.Port)
		}
		store, err = storage.NewMongoStorage(uri, cfg.Database.Database)
		if err != nil {
			logger.Fatalf("Failed to initialize MongoDB storage: %v", err)
		}
		logger.Infof("Using MongoDB storage: %s:%d/%s",
			cfg.Database.Host, cfg.Database.Port, cfg.Database.Database)

	default:
		logger.Fatalf("Unsupported storage driver: %s", cfg.Database.Driver)
	}

	// Initialize authentication configuration
	authConfig := &middleware.AuthConfig{
		APIKeys:       make(map[string]string),
		JWTSecret:     cfg.Security.JWTSecret,
		JWTExpiration: cfg.Security.JWTExpiration,
	}

	// Add default API keys from config if available
	for key, clientID := range cfg.Security.APIKeys {
		authConfig.APIKeys[key] = clientID
	}

	logger.Info("Authentication configured")

	// Create API server
	server := api.NewServer(store, logger, metrics, authConfig)
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
