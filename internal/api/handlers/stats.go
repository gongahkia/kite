package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
)

// StatsHandler handles statistics-related requests
type StatsHandler struct {
	storage storage.Storage
	logger  *observability.Logger
}

// NewStatsHandler creates a new StatsHandler
func NewStatsHandler(storage storage.Storage, logger *observability.Logger) *StatsHandler {
	return &StatsHandler{
		storage: storage,
		logger:  logger,
	}
}

// GetStats handles GET /api/v1/stats
func (h *StatsHandler) GetStats(c *fiber.Ctx) error {
	// Get storage stats if available
	if memStorage, ok := h.storage.(*storage.MemoryStorage); ok {
		stats := memStorage.GetStats()
		return c.JSON(stats)
	}

	return c.JSON(fiber.Map{
		"message": "Statistics not available for this storage backend",
	})
}

// GetStorageStats handles GET /api/v1/stats/storage
func (h *StatsHandler) GetStorageStats(c *fiber.Ctx) error {
	// Get storage stats if available
	if memStorage, ok := h.storage.(*storage.MemoryStorage); ok {
		stats := memStorage.GetStats()
		return c.JSON(stats)
	}

	return c.JSON(fiber.Map{
		"message": "Storage statistics not available",
	})
}
