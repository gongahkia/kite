package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
)

// HealthCheck handles GET /health
func HealthCheck(storage storage.Storage) fiber.Handler {
	return func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"status": "healthy",
			"service": "kite-api",
			"version": "4.0.0",
		})
	}
}

// ReadinessCheck handles GET /ready
func ReadinessCheck(storage storage.Storage) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check storage connection
		if err := storage.Ping(c.Context()); err != nil {
			return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
				"status": "not ready",
				"error":  "storage unavailable",
			})
		}

		return c.JSON(fiber.Map{
			"status": "ready",
			"service": "kite-api",
			"version": "4.0.0",
		})
	}
}

// MetricsHandler handles GET /metrics
func MetricsHandler(metrics *observability.Metrics) fiber.Handler {
	return adaptor.HTTPHandler(metrics.Handler())
}
