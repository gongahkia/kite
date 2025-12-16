package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// JudgeHandler handles judge-related requests
type JudgeHandler struct {
	storage storage.Storage
	logger  *observability.Logger
}

// NewJudgeHandler creates a new JudgeHandler
func NewJudgeHandler(storage storage.Storage, logger *observability.Logger) *JudgeHandler {
	return &JudgeHandler{
		storage: storage,
		logger:  logger,
	}
}

// ListJudges handles GET /api/v1/judges
func (h *JudgeHandler) ListJudges(c *fiber.Ctx) error {
	filter := storage.JudgeFilter{
		Name:         c.Query("name"),
		Court:        c.Query("court"),
		Jurisdiction: c.Query("jurisdiction"),
		Limit:        c.QueryInt("limit", 10),
		Offset:       c.QueryInt("offset", 0),
	}

	judges, err := h.storage.ListJudges(c.Context(), filter)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data":   judges,
		"total":  len(judges),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetJudge handles GET /api/v1/judges/:id
func (h *JudgeHandler) GetJudge(c *fiber.Ctx) error {
	id := c.Params("id")

	judge, err := h.storage.GetJudge(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(judge)
}

// CreateJudge handles POST /api/v1/judges
func (h *JudgeHandler) CreateJudge(c *fiber.Ctx) error {
	var judge models.Judge
	if err := c.BodyParser(&judge); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.storage.SaveJudge(c.Context(), &judge); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(judge)
}

// UpdateJudge handles PUT /api/v1/judges/:id
func (h *JudgeHandler) UpdateJudge(c *fiber.Ctx) error {
	id := c.Params("id")

	var judge models.Judge
	if err := c.BodyParser(&judge); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	judge.ID = id

	if err := h.storage.UpdateJudge(c.Context(), &judge); err != nil {
		return err
	}

	return c.JSON(judge)
}
