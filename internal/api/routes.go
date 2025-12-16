package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gongahkia/kite/internal/api/handlers"
	"github.com/gongahkia/kite/internal/api/middleware"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
)

// Server represents the HTTP server
type Server struct {
	app     *fiber.App
	storage storage.Storage
	logger  *observability.Logger
	metrics *observability.Metrics
}

// NewServer creates a new API server
func NewServer(storage storage.Storage, logger *observability.Logger, metrics *observability.Metrics) *Server {
	app := fiber.New(fiber.Config{
		AppName:      "Kite API v4.0.0",
		ServerHeader: "Kite",
		ErrorHandler: middleware.ErrorHandler(logger),
	})

	return &Server{
		app:     app,
		storage: storage,
		logger:  logger,
		metrics: metrics,
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

	// Health endpoints
	s.app.Get("/health", handlers.HealthCheck(s.storage))
	s.app.Get("/ready", handlers.ReadinessCheck(s.storage))

	// Metrics endpoint
	s.app.Get("/metrics", handlers.MetricsHandler(s.metrics))

	// API v1 routes
	api := s.app.Group("/api/v1")

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
