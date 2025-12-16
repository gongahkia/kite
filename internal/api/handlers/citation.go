package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/pkg/models"
)

// CitationHandler handles citation-related requests
type CitationHandler struct {
	storage storage.Storage
	logger  *observability.Logger
}

// NewCitationHandler creates a new CitationHandler
func NewCitationHandler(storage storage.Storage, logger *observability.Logger) *CitationHandler {
	return &CitationHandler{
		storage: storage,
		logger:  logger,
	}
}

// ListCitations handles GET /api/v1/citations
func (h *CitationHandler) ListCitations(c *fiber.Ctx) error {
	filter := storage.CitationFilter{
		CaseID: c.Query("case_id"),
		Format: c.Query("format"),
		Year:   c.QueryInt("year", 0),
		Limit:  c.QueryInt("limit", 10),
		Offset: c.QueryInt("offset", 0),
	}

	citations, err := h.storage.ListCitations(c.Context(), filter)
	if err != nil {
		return err
	}

	return c.JSON(fiber.Map{
		"data":   citations,
		"total":  len(citations),
		"limit":  filter.Limit,
		"offset": filter.Offset,
	})
}

// GetCitation handles GET /api/v1/citations/:id
func (h *CitationHandler) GetCitation(c *fiber.Ctx) error {
	id := c.Params("id")

	citation, err := h.storage.GetCitation(c.Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(citation)
}

// CreateCitation handles POST /api/v1/citations
func (h *CitationHandler) CreateCitation(c *fiber.Ctx) error {
	var citation models.Citation
	if err := c.BodyParser(&citation); err != nil {
		return fiber.NewError(fiber.StatusBadRequest, "Invalid request body")
	}

	if err := h.storage.SaveCitation(c.Context(), &citation); err != nil {
		return err
	}

	return c.Status(fiber.StatusCreated).JSON(citation)
}
