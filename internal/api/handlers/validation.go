package handlers

import (
	"github.com/gofiber/fiber/v2"

	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/internal/storage"
	"github.com/gongahkia/kite/internal/validation"
	"github.com/gongahkia/kite/pkg/models"
)

// ValidationHandler handles validation requests
type ValidationHandler struct {
	pipeline  *validation.Pipeline
	detector  *validation.DuplicateDetector
	storage   storage.Storage
	logger    *observability.Logger
	metrics   *observability.Metrics
}

// NewValidationHandler creates a new validation handler
func NewValidationHandler(storage storage.Storage, logger *observability.Logger, metrics *observability.Metrics) *ValidationHandler {
	pipeline := validation.DefaultPipeline(logger, metrics)

	return &ValidationHandler{
		pipeline: pipeline,
		detector: validation.NewDuplicateDetector(),
		storage:  storage,
		logger:   logger,
		metrics:  metrics,
	}
}

// ValidateCaseRequest represents a validation request
type ValidateCaseRequest struct {
	CaseID string `json:"case_id"`
}

// ValidateCaseResponse represents a validation response
type ValidateCaseResponse struct {
	Valid         bool                      `json:"valid"`
	Score         float64                   `json:"score"`
	Completeness  float64                   `json:"completeness"`
	Errors        []validation.ValidationError   `json:"errors,omitempty"`
	Warnings      []validation.ValidationWarning `json:"warnings,omitempty"`
	ShouldReject  bool                      `json:"should_reject"`
	Duration      float64                   `json:"duration_ms"`
}

// ValidateCase handles POST /api/v1/validation/case
func (h *ValidationHandler) ValidateCase(c *fiber.Ctx) error {
	var req ValidateCaseRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	// Get case from storage
	caseData, err := h.storage.GetCase(c.Context(), req.CaseID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
			"error": "Case not found",
		})
	}

	// Validate case
	report, err := h.pipeline.Validate(c.Context(), caseData)
	if err != nil {
		h.logger.WithField("error", err).Error("Validation failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Validation failed",
		})
	}

	return c.JSON(ValidateCaseResponse{
		Valid:        report.Valid,
		Score:        report.OverallScore,
		Completeness: report.Completeness,
		Errors:       report.Errors,
		Warnings:     report.Warnings,
		ShouldReject: report.ShouldReject(),
		Duration:     float64(report.Duration.Milliseconds()),
	})
}

// ValidateBatchRequest represents a batch validation request
type ValidateBatchRequest struct {
	CaseIDs []string `json:"case_ids"`
}

// ValidateBatchResponse represents a batch validation response
type ValidateBatchResponse struct {
	Results []*ValidateCaseResponse `json:"results"`
	Summary *ValidationSummary      `json:"summary"`
}

// ValidationSummary summarizes batch validation results
type ValidationSummary struct {
	TotalCases    int     `json:"total_cases"`
	ValidCases    int     `json:"valid_cases"`
	InvalidCases  int     `json:"invalid_cases"`
	AverageScore  float64 `json:"average_score"`
	AverageCompl  float64 `json:"average_completeness"`
}

// ValidateBatch handles POST /api/v1/validation/batch
func (h *ValidationHandler) ValidateBatch(c *fiber.Ctx) error {
	var req ValidateBatchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	if len(req.CaseIDs) == 0 {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "No case IDs provided",
		})
	}

	// Get cases from storage
	cases := make([]*models.Case, 0, len(req.CaseIDs))
	for _, id := range req.CaseIDs {
		caseData, err := h.storage.GetCase(c.Context(), id)
		if err != nil {
			h.logger.WithField("case_id", id).Warn("Case not found in batch")
			continue
		}
		cases = append(cases, caseData)
	}

	// Validate batch
	reports, err := h.pipeline.ValidateBatch(c.Context(), cases)
	if err != nil {
		h.logger.WithField("error", err).Error("Batch validation failed")
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Batch validation failed",
		})
	}

	// Convert reports to responses
	results := make([]*ValidateCaseResponse, len(reports))
	validCount := 0
	totalScore := 0.0
	totalCompl := 0.0

	for i, report := range reports {
		if report == nil {
			continue
		}

		results[i] = &ValidateCaseResponse{
			Valid:        report.Valid,
			Score:        report.OverallScore,
			Completeness: report.Completeness,
			Errors:       report.Errors,
			Warnings:     report.Warnings,
			ShouldReject: report.ShouldReject(),
			Duration:     float64(report.Duration.Milliseconds()),
		}

		if report.Valid {
			validCount++
		}
		totalScore += report.OverallScore
		totalCompl += report.Completeness
	}

	summary := &ValidationSummary{
		TotalCases:   len(results),
		ValidCases:   validCount,
		InvalidCases: len(results) - validCount,
	}

	if len(results) > 0 {
		summary.AverageScore = totalScore / float64(len(results))
		summary.AverageCompl = totalCompl / float64(len(results))
	}

	return c.JSON(ValidateBatchResponse{
		Results: results,
		Summary: summary,
	})
}

// DetectDuplicatesRequest represents a duplicate detection request
type DetectDuplicatesRequest struct {
	CaseIDs []string `json:"case_ids,omitempty"`
}

// DetectDuplicatesResponse represents a duplicate detection response
type DetectDuplicatesResponse struct {
	Groups []*DuplicateGroupResponse `json:"groups"`
	Total  int                       `json:"total_duplicates"`
}

// DuplicateGroupResponse represents a group of duplicates
type DuplicateGroupResponse struct {
	CaseIDs    []string `json:"case_ids"`
	Similarity float64  `json:"similarity"`
	Type       string   `json:"type"`
}

// DetectDuplicates handles POST /api/v1/validation/duplicates
func (h *ValidationHandler) DetectDuplicates(c *fiber.Ctx) error {
	var req DetectDuplicatesRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "Invalid request body",
		})
	}

	var cases []*models.Case
	var err error

	if len(req.CaseIDs) > 0 {
		// Check specific cases
		cases = make([]*models.Case, 0, len(req.CaseIDs))
		for _, id := range req.CaseIDs {
			caseData, err := h.storage.GetCase(c.Context(), id)
			if err != nil {
				h.logger.WithField("case_id", id).Warn("Case not found")
				continue
			}
			cases = append(cases, caseData)
		}
	} else {
		// Check all cases (limited to 1000)
		filter := storage.CaseFilter{Limit: 1000}
		cases, err = h.storage.ListCases(c.Context(), filter)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to list cases",
			})
		}
	}

	// Detect duplicates
	groups := h.detector.FindDuplicateGroups(cases)

	// Convert to response
	groupResponses := make([]*DuplicateGroupResponse, len(groups))
	for i, g := range groups {
		groupResponses[i] = &DuplicateGroupResponse{
			CaseIDs:    g.CaseIDs,
			Similarity: g.Similarity,
			Type:       g.Type,
		}
	}

	return c.JSON(DetectDuplicatesResponse{
		Groups: groupResponses,
		Total:  len(groups),
	})
}

// GetQualityMetrics handles GET /api/v1/validation/metrics
func (h *ValidationHandler) GetQualityMetrics(c *fiber.Ctx) error {
	// Get all cases (limited to 1000 for metrics)
	filter := storage.CaseFilter{Limit: 1000}
	cases, err := h.storage.ListCases(c.Context(), filter)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to list cases",
		})
	}

	// Validate all cases
	reports, err := h.pipeline.ValidateBatch(c.Context(), cases)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to calculate metrics",
		})
	}

	// Calculate metrics
	metrics := validation.CalculateMetrics(reports)

	return c.JSON(fiber.Map{
		"total_cases":        metrics.TotalCases,
		"high_quality":       metrics.HighQuality,
		"medium_quality":     metrics.MediumQuality,
		"low_quality":        metrics.LowQuality,
		"average_score":      metrics.AverageScore,
		"average_complete":   metrics.AverageComplete,
		"completeness_histogram": metrics.CompletenessHist,
	})
}
