package validation

import (
	"context"
	"strings"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// QualityValidator assesses overall quality and completeness
type QualityValidator struct{}

func NewQualityValidator() *QualityValidator {
	return &QualityValidator{}
}

func (v *QualityValidator) Name() string {
	return "quality"
}

func (v *QualityValidator) Stage() ValidationStage {
	return StageQuality
}

func (v *QualityValidator) Validate(ctx context.Context, c *models.Case) (*ValidationResult, error) {
	start := time.Now()

	// Calculate completeness
	completeness := calculateCompleteness(c)

	// Calculate content quality
	contentQuality := calculateContentQuality(c)

	// Calculate metadata quality
	metadataQuality := calculateMetadataQuality(c)

	// Overall quality score
	qualityScore := (completeness + contentQuality + metadataQuality) / 3.0

	// Update case quality score
	c.QualityScore = &qualityScore

	result := &ValidationResult{
		Valid:        qualityScore >= 0.5, // Minimum quality threshold
		Stage:        StageQuality,
		Score:        qualityScore,
		Completeness: completeness,
		Errors:       []ValidationError{},
		Warnings:     []ValidationWarning{},
	}

	// Generate warnings for low quality
	if qualityScore < 0.7 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "quality",
			Message: "Case quality is below recommended threshold (0.7)",
			Code:    "LOW_QUALITY",
		})
	}

	if completeness < 0.6 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "completeness",
			Message: "Case data is incomplete (<60% complete)",
			Code:    "INCOMPLETE_DATA",
		})
	}

	result.Duration = time.Since(start)

	return result, nil
}

// calculateCompleteness calculates data completeness score (0-1)
func calculateCompleteness(c *models.Case) float64 {
	fields := 0
	filled := 0

	// Core fields (weight: 1.0)
	coreFields := map[string]bool{
		"id":           c.ID != "",
		"case_number":  c.CaseNumber != "",
		"case_name":    c.CaseName != "",
		"court":        c.Court != "",
		"jurisdiction": c.Jurisdiction != "",
	}

	for _, isFilled := range coreFields {
		fields++
		if isFilled {
			filled++
		}
	}

	// Important fields (weight: 0.8)
	importantFields := map[string]bool{
		"decision_date": c.DecisionDate != nil,
		"summary":       c.Summary != "",
		"court_level":   c.CourtLevel > 0,
	}

	for _, isFilled := range importantFields {
		fields++
		if isFilled {
			filled++
		}
	}

	// Optional enrichment fields (weight: 0.5)
	enrichmentFields := map[string]bool{
		"full_text":      c.FullText != "",
		"parties":        len(c.Parties) > 0,
		"judges":         len(c.Judges) > 0,
		"legal_concepts": len(c.LegalConcepts) > 0,
		"citations":      len(c.Citations) > 0,
		"url":            c.URL != "",
		"pdf_url":        c.PDFURL != "",
	}

	for _, isFilled := range enrichmentFields {
		fields++
		if isFilled {
			filled++
		}
	}

	if fields == 0 {
		return 0.0
	}

	return float64(filled) / float64(fields)
}

// calculateContentQuality assesses the quality of textual content
func calculateContentQuality(c *models.Case) float64 {
	score := 0.0
	checks := 0

	// Check 1: Case name quality
	if c.CaseName != "" {
		checks++
		score += assessTextQuality(c.CaseName, 10, 200)
	}

	// Check 2: Summary quality
	if c.Summary != "" {
		checks++
		score += assessTextQuality(c.Summary, 50, 5000)
	}

	// Check 3: Full text quality
	if c.FullText != "" {
		checks++
		score += assessTextQuality(c.FullText, 100, 100000)
	}

	// Check 4: Has meaningful parties
	if len(c.Parties) >= 2 {
		checks++
		score += 1.0
	} else if len(c.Parties) == 1 {
		checks++
		score += 0.5
	}

	// Check 5: Has judges
	if len(c.Judges) > 0 {
		checks++
		score += 1.0
	}

	if checks == 0 {
		return 0.5 // Default for minimal content
	}

	return score / float64(checks)
}

// calculateMetadataQuality assesses metadata quality
func calculateMetadataQuality(c *models.Case) float64 {
	score := 0.0
	checks := 0

	// Check 1: Has valid decision date
	if c.DecisionDate != nil {
		checks++
		if c.DecisionDate.Before(time.Now()) && c.DecisionDate.After(time.Now().AddDate(-200, 0, 0)) {
			score += 1.0
		} else {
			score += 0.3
		}
	}

	// Check 2: Has court level
	if c.CourtLevel >= 1 && c.CourtLevel <= 3 {
		checks++
		score += 1.0
	}

	// Check 3: Has jurisdiction metadata
	if c.Jurisdiction != "" {
		checks++
		score += 1.0
	}

	// Check 4: Has court type
	if c.CourtType != "" {
		checks++
		score += 1.0
	}

	// Check 5: Has legal concepts
	if len(c.LegalConcepts) > 0 {
		checks++
		conceptScore := float64(len(c.LegalConcepts)) / 10.0 // Up to 10 concepts
		if conceptScore > 1.0 {
			conceptScore = 1.0
		}
		score += conceptScore
	}

	// Check 6: Has citations
	if len(c.Citations) > 0 {
		checks++
		citationScore := float64(len(c.Citations)) / 10.0 // Up to 10 citations
		if citationScore > 1.0 {
			citationScore = 1.0
		}
		score += citationScore
	}

	// Check 7: Has source information
	if c.SourceDatabase != "" {
		checks++
		score += 1.0
	}

	if checks == 0 {
		return 0.5 // Default
	}

	return score / float64(checks)
}

// assessTextQuality assesses the quality of a text field
func assessTextQuality(text string, minLen, maxLen int) float64 {
	if text == "" {
		return 0.0
	}

	text = strings.TrimSpace(text)
	length := len(text)

	// Check length
	if length < minLen {
		return 0.3 // Too short
	}

	if length > maxLen {
		return 0.8 // Too long but still has content
	}

	score := 1.0

	// Check for suspicious patterns
	lowerText := strings.ToLower(text)

	// Repeated characters (e.g., "aaaaa")
	if hasRepeatedChars(text, 5) {
		score -= 0.2
	}

	// All caps (SHOUTING)
	if text == strings.ToUpper(text) && length > 20 {
		score -= 0.1
	}

	// Common placeholder text
	placeholders := []string{"lorem ipsum", "test", "example", "placeholder", "todo", "tbd"}
	for _, placeholder := range placeholders {
		if strings.Contains(lowerText, placeholder) {
			score -= 0.3
		}
	}

	// Too many special characters
	specialCharCount := 0
	for _, r := range text {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == ' ') {
			specialCharCount++
		}
	}

	specialRatio := float64(specialCharCount) / float64(length)
	if specialRatio > 0.3 {
		score -= 0.2
	}

	if score < 0 {
		score = 0
	}

	return score
}

// hasRepeatedChars checks if text has too many repeated characters
func hasRepeatedChars(text string, threshold int) bool {
	if len(text) < threshold {
		return false
	}

	for i := 0; i <= len(text)-threshold; i++ {
		char := text[i]
		repeated := true
		for j := 1; j < threshold; j++ {
			if text[i+j] != char {
				repeated = false
				break
			}
		}
		if repeated {
			return true
		}
	}

	return false
}

// QualityMetrics represents aggregated quality metrics
type QualityMetrics struct {
	TotalCases       int
	HighQuality      int     // Score >= 0.8
	MediumQuality    int     // Score >= 0.6
	LowQuality       int     // Score < 0.6
	AverageScore     float64
	AverageComplete  float64
	CompletenessHist map[string]int
}

// CalculateMetrics calculates quality metrics for a batch of cases
func CalculateMetrics(reports []*ValidationReport) *QualityMetrics {
	metrics := &QualityMetrics{
		TotalCases:       len(reports),
		CompletenessHist: make(map[string]int),
	}

	totalScore := 0.0
	totalComplete := 0.0

	for _, report := range reports {
		totalScore += report.OverallScore
		totalComplete += report.Completeness

		// Categorize by quality
		if report.OverallScore >= 0.8 {
			metrics.HighQuality++
		} else if report.OverallScore >= 0.6 {
			metrics.MediumQuality++
		} else {
			metrics.LowQuality++
		}

		// Completeness histogram
		bucket := getBucket(report.Completeness)
		metrics.CompletenessHist[bucket]++
	}

	if metrics.TotalCases > 0 {
		metrics.AverageScore = totalScore / float64(metrics.TotalCases)
		metrics.AverageComplete = totalComplete / float64(metrics.TotalCases)
	}

	return metrics
}

func getBucket(value float64) string {
	switch {
	case value >= 0.9:
		return "90-100%"
	case value >= 0.8:
		return "80-90%"
	case value >= 0.7:
		return "70-80%"
	case value >= 0.6:
		return "60-70%"
	default:
		return "<60%"
	}
}
