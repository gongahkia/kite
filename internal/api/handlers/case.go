package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// CaseHandler handles case-related requests
type CaseHandler struct {
	storage storage.Storage
	logger  *observability.Logger
}

// NewCaseHandler creates a new CaseHandler
func NewCaseHandler(storage storage.Storage, logger *observability.Logger) *CaseHandler {
	return &CaseHandler{
		storage: storage,
		logger:  logger,
	}
}

// ListCases handles GET /api/v1/cases
func (h *CaseHandler) ListCases(c *fiber.Ctx) error {
	filter := storage.CaseFilter{
		Jurisdiction: c.Query("jurisdiction"),
		Court:        c.Query("court"),
		Limit:        c.QueryInt("limit", 10),
		Offset:       c.QueryInt("offset", 0),
	}

	cases, err := h.storage.ListCases(c.Context(), filter)
	if err != nil {
		return err
	}

	total, _ := h.storage.CountCases(c.Context(), filter)

	return c.JSON(fiber.Map{
		"data":  cases,
		"total": total,
		"limit": filter.Limit,
		"offset": filter.Offset,
	})
}

// GetCase handles GET /api/v1/cases/:id
func (h *CaseHandler) GetCase(c *fiber.Ctx) error {
	id := c.Params("id")

	caseData, err := h.storage.GetCase(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(caseData)
}

// CreateCase handles POST /api/v1/cases
func (h *CaseHandler) CreateCase(c *fiber.Ctx) error {
	var caseData models.Case
	if err := c.BodyParser(&caseData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.storage.SaveCase(c.Context(), &caseData); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(caseData)
}

// UpdateCase handles PUT /api/v1/cases/:id
func (h *CaseHandler) UpdateCase(c *fiber.Ctx) error {
	id := c.Params("id")

	var caseData models.Case
	if err := c.BodyParser(&caseData); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	caseData.ID = id

	if err := h.storage.UpdateCase(c.Context(), &caseData); err != nil {
		return err
	}

	return c.JSON(caseData)
}

// DeleteCase handles DELETE /api/v1/cases/:id
func (h *CaseHandler) DeleteCase(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := h.storage.DeleteCase(c.Context(), id); err != nil {
		return err
	}

	return c.SendStatus(fiber.StatusNoContent)
}

// SearchCases handles POST /api/v1/cases/search
func (h *CaseHandler) SearchCases(c *fiber.Ctx) error {
	var query storage.SearchQuery
	if err := c.BodyParser(&query); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if query.Limit == 0 {
		query.Limit = 10
	}

	cases, err := h.storage.SearchCases(c.Context(), query)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data":   cases,
		"query":  query.Query,
		"total":  len(cases),
		"limit":  query.Limit,
		"offset": query.Offset,
	})
}
