package api

import (
	"time"

	"github.com/gofiber/fiber/v2"
	swagger "github.com/swaggo/fiber-swagger"
	"github.com/gongahkia/kite/internal/api/handlers"
	"github.com/gongahkia/kite/internal/api/middleware"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
	_ "github.com/gongahkia/kite/docs" // Import generated docs
)

// Server represents the HTTP server
type Server struct {
	app        *fiber.App
	storage    storage.Storage
	logger     *observability.Logger
	metrics    *observability.Metrics
	authConfig *middleware.AuthConfig
}

// NewServer creates a new API server
func NewServer(storage storage.Storage, logger *observability.Logger, metrics *observability.Metrics, authConfig *middleware.AuthConfig) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "Kite API v4.0.0",
		ServerHeader: "Kite",
		ErrorHandler: middleware.ErrorHandler(logger),
	})

	// Set default auth config if not provided
	if authConfig == nil {
		authConfig = middleware.DefaultAuthConfig()
		// Set a default JWT secret (in production, load from config/env)
		authConfig.JWTSecret = "your-secret-key-change-in-production"
		authConfig.JWTExpiration = 24 * time.Hour
	}

	return &Server{
		app:        app,
		storage:    storage,
		logger:     logger,
		metrics:    metrics,
		authConfig: authConfig,
	}
}

// SetupRoutes configures all API routes
func (s *Server) SetupRoutes() {
	// Apply global middleware
	s.app.Use(middleware.RequestID())
	s.app.Use(middleware.Logger(s.logger))
	s.app.Use(middleware.CORS())
	s.app.Use(middleware.Recovery(s.logger))
	s.app.Use(middleware.Metrics(s.metrics))
	s.app.Use(middleware.IPRateLimit(100, 200, s.logger)) // Global rate limit: 100 req/s

	// Swagger UI documentation
	s.app.Get("/swagger/*", swagger.HandlerDefault)
	s.app.Get("/docs", func(c *fiber.Ctx) error {
		return c.Redirect("/swagger/index.html", fiber.StatusMovedPermanently)
	})

	// Health endpoints (no auth required)
	s.app.Get("/health", handlers.HealthCheck(s.storage))
	s.app.Get("/ready", handlers.ReadinessCheck(s.storage))

	// Metrics endpoint (no auth required)
	s.app.Get("/metrics", handlers.MetricsHandler(s.metrics))

	// API v1 routes
	api := s.app.Group("/api/v1")

	// Auth endpoints (no auth required for these)
	authHandler := handlers.NewAuthHandler(s.logger, s.authConfig)
	auth := api.Group("/auth")
	auth.Post("/login", authHandler.Login)
	auth.Post("/api-key", middleware.JWTAuth(s.authConfig, s.logger), authHandler.GenerateAPIKey)
	auth.Post("/refresh", middleware.JWTAuth(s.authConfig, s.logger), authHandler.RefreshToken)
	auth.Get("/validate", middleware.OptionalAuth(s.authConfig, s.logger), authHandler.ValidateToken)

	// Configure endpoint-specific rate limiting
	endpointRateLimitConfig := middleware.DefaultEndpointRateLimitConfig()
	endpointRateLimitConfig.UseClientID = true

	// Apply optional auth to all API routes (allows both authenticated and unauthenticated access)
	// For production, you may want to require auth for all routes except public endpoints
	api.Use(middleware.OptionalAuth(s.authConfig, s.logger))
	api.Use(middleware.EndpointRateLimit(endpointRateLimitConfig, s.logger))

	// Case routes
	caseHandler := handlers.NewCaseHandler(s.storage, s.logger)
	cases := api.Group("/cases")
	cases.Get("/", caseHandler.ListCases)
	cases.Get("/:id", caseHandler.GetCase)
	cases.Post("/", caseHandler.CreateCase)
	cases.Put("/:id", caseHandler.UpdateCase)
	cases.Delete("/:id", caseHandler.DeleteCase)
	cases.Post("/search", caseHandler.SearchCases)

	// Judge routes
	judgeHandler := handlers.NewJudgeHandler(s.storage, s.logger)
	judges := api.Group("/judges")
	judges.Get("/", judgeHandler.ListJudges)
	judges.Get("/:id", judgeHandler.GetJudge)
	judges.Post("/", judgeHandler.CreateJudge)
	judges.Put("/:id", judgeHandler.UpdateJudge)

	// Citation routes
	citationHandler := handlers.NewCitationHandler(s.storage, s.logger)
	citations := api.Group("/citations")
	citations.Get("/", citationHandler.ListCitations)
	citations.Get("/:id", citationHandler.GetCitation)
	citations.Post("/", citationHandler.CreateCitation)

	// Search routes (advanced search API)
	searchHandler := handlers.NewSearchHandler(s.storage, s.logger, s.metrics)
	searchGroup := api.Group("/search")
	searchGroup.Post("/", searchHandler.Search)
	searchGroup.Get("/suggest", searchHandler.Suggest)
	searchGroup.Get("/autocomplete", searchHandler.Autocomplete)

	// Validation routes
	validationHandler := handlers.NewValidationHandler(s.storage, s.logger, s.metrics)
	validation := api.Group("/validation")
	validation.Post("/case", validationHandler.ValidateCase)
	validation.Post("/batch", validationHandler.ValidateBatch)
	validation.Post("/duplicates", validationHandler.DetectDuplicates)
	validation.Get("/metrics", validationHandler.GetQualityMetrics)

	// Stats routes
	statsHandler := handlers.NewStatsHandler(s.storage, s.logger)
	stats := api.Group("/stats")
	stats.Get("/", statsHandler.GetStats)
	stats.Get("/storage", statsHandler.GetStorageStats)

	// 404 handler
	s.app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Resource not found",
			"path":  c.Path(),
		})
	})
}

// GetApp returns the Fiber app
func (s *Server) GetApp() *fiber.App {
	return s.app
}

// Start starts the HTTP server
func (s *Server) Start(address string) error {
	return s.app.Listen(address)
}

// Shutdown gracefully shuts down the server
func (s *Server) Shutdown() error {
	return s.app.Shutdown()
}
