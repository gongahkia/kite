package validation

import (
	"context"
	"regexp"
	"strings"
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// StructuralValidator validates structural integrity
type StructuralValidator struct{}

func NewStructuralValidator() *StructuralValidator {
	return &StructuralValidator{}
}

func (v *StructuralValidator) Name() string {
	return "structural"
}

func (v *StructuralValidator) Stage() ValidationStage {
	return StageStructural
}

func (v *StructuralValidator) Validate(ctx context.Context, c *models.Case) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Valid:  true,
		Stage:  StageStructural,
		Errors: []ValidationError{},
	}

	// Required fields
	if c.CaseName == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "case_name",
			Message: "Case name is required",
			Code:    "REQUIRED_FIELD_MISSING",
		})
	}

	if c.Court == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "court",
			Message: "Court is required",
			Code:    "REQUIRED_FIELD_MISSING",
		})
	}

	if c.Jurisdiction == "" {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "jurisdiction",
			Message: "Jurisdiction is required",
			Code:    "REQUIRED_FIELD_MISSING",
		})
	}

	// Field length validation
	if len(c.CaseName) > 500 {
		result.Valid = false
		result.Errors = append(result.Errors, ValidationError{
			Field:   "case_name",
			Message: "Case name exceeds maximum length (500)",
			Code:    "FIELD_TOO_LONG",
		})
	}

	if len(c.Summary) > 10000 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "summary",
			Message: "Summary is very long (>10000 chars)",
			Code:    "FIELD_VERY_LONG",
		})
	}

	result.Duration = time.Since(start)
	result.Score = calculateStructuralScore(result)

	return result, nil
}

// SemanticValidator validates semantic correctness
type SemanticValidator struct{}

func NewSemanticValidator() *SemanticValidator {
	return &SemanticValidator{}
}

func (v *SemanticValidator) Name() string {
	return "semantic"
}

func (v *SemanticValidator) Stage() ValidationStage {
	return StageSemantic
}

func (v *SemanticValidator) Validate(ctx context.Context, c *models.Case) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Valid:    true,
		Stage:    StageSemantic,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Date validation
	if c.DecisionDate != nil {
		// Decision date should not be in the future
		if c.DecisionDate.After(time.Now()) {
			result.Valid = false
			result.Errors = append(result.Errors, ValidationError{
				Field:   "decision_date",
				Message: "Decision date cannot be in the future",
				Code:    "INVALID_DATE",
			})
		}

		// Decision date should not be too old (>200 years)
		minDate := time.Now().AddDate(-200, 0, 0)
		if c.DecisionDate.Before(minDate) {
			result.Warnings = append(result.Warnings, ValidationWarning{
				Field:   "decision_date",
				Message: "Decision date is very old (>200 years)",
				Code:    "SUSPICIOUS_DATE",
			})
		}
	}

	// Case number format validation
	if c.CaseNumber != "" && !isValidCaseNumber(c.CaseNumber) {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "case_number",
			Message: "Case number format appears invalid",
			Code:    "INVALID_FORMAT",
		})
	}

	// Court level validation
	if c.CourtLevel < 1 || c.CourtLevel > 3 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "court_level",
			Message: "Court level should be 1 (Trial), 2 (Appellate), or 3 (Supreme)",
			Code:    "INVALID_VALUE",
		})
	}

	// Party validation
	if len(c.Parties) > 0 {
		for _, party := range c.Parties {
			if strings.TrimSpace(party) == "" {
				result.Warnings = append(result.Warnings, ValidationWarning{
					Field:   "parties",
					Message: "Empty party name found",
					Code:    "EMPTY_VALUE",
				})
			}
		}
	}

	result.Duration = time.Since(start)
	result.Score = calculateSemanticScore(result)

	return result, nil
}

// BusinessRulesValidator validates business rules
type BusinessRulesValidator struct{}

func NewBusinessRulesValidator() *BusinessRulesValidator {
	return &BusinessRulesValidator{}
}

func (v *BusinessRulesValidator) Name() string {
	return "business_rules"
}

func (v *BusinessRulesValidator) Stage() ValidationStage {
	return StageBusinessRules
}

func (v *BusinessRulesValidator) Validate(ctx context.Context, c *models.Case) (*ValidationResult, error) {
	start := time.Now()
	result := &ValidationResult{
		Valid:    true,
		Stage:    StageBusinessRules,
		Errors:   []ValidationError{},
		Warnings: []ValidationWarning{},
	}

	// Rule 1: Supreme Court cases should have court_level = 3
	if strings.Contains(strings.ToLower(c.Court), "supreme") && c.CourtLevel != 3 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "court_level",
			Message: "Supreme Court case should have court_level = 3",
			Code:    "RULE_VIOLATION",
		})
	}

	// Rule 2: Appeal Court cases should have court_level = 2
	if strings.Contains(strings.ToLower(c.Court), "appeal") && c.CourtLevel != 2 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "court_level",
			Message: "Appeal Court case should have court_level = 2",
			Code:    "RULE_VIOLATION",
		})
	}

	// Rule 3: Cases with PDF should have PDF URL
	if c.PDFURL == "" && c.URL != "" {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "pdf_url",
			Message: "Case has URL but no PDF URL",
			Code:    "MISSING_RELATED_FIELD",
		})
	}

	// Rule 4: Criminal cases should have certain fields
	isCriminal := containsAny(c.LegalConcepts, []string{"criminal law", "criminal procedure"})
	if isCriminal && len(c.Parties) < 2 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "parties",
			Message: "Criminal case should have at least 2 parties",
			Code:    "RULE_VIOLATION",
		})
	}

	// Rule 5: Summary should be present for significant cases
	if c.Summary == "" && c.CourtLevel >= 2 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "summary",
			Message: "Higher court cases should have a summary",
			Code:    "MISSING_RECOMMENDED_FIELD",
		})
	}

	// Rule 6: Legal concepts should be present
	if len(c.LegalConcepts) == 0 {
		result.Warnings = append(result.Warnings, ValidationWarning{
			Field:   "legal_concepts",
			Message: "No legal concepts tagged",
			Code:    "MISSING_ENRICHMENT",
		})
	}

	result.Duration = time.Since(start)
	result.Score = calculateBusinessRulesScore(result)

	return result, nil
}

// Helper functions

func isValidCaseNumber(caseNumber string) bool {
	// Common case number patterns
	patterns := []string{
		`^\d+/\d+$`,                    // e.g., "123/2023"
		`^\[\d+\]\s+\w+\s+\d+$`,        // e.g., "[2023] SCR 123"
		`^\d+\s+\w+\.?\s*\d*\s+\d+$`,  // e.g., "123 U.S. 456"
		`^\w+-\d+-\d+$`,                 // e.g., "CV-2023-001"
		`^\d{4}-\d+$`,                   // e.g., "2023-12345"
	}

	for _, pattern := range patterns {
		if matched, _ := regexp.MatchString(pattern, caseNumber); matched {
			return true
		}
	}

	// If it contains digits and has reasonable length, accept it
	hasDigits, _ := regexp.MatchString(`\d`, caseNumber)
	return hasDigits && len(caseNumber) >= 4 && len(caseNumber) <= 50
}

func containsAny(slice []string, items []string) bool {
	for _, s := range slice {
		for _, item := range items {
			if strings.Contains(strings.ToLower(s), strings.ToLower(item)) {
				return true
			}
		}
	}
	return false
}

func calculateStructuralScore(result *ValidationResult) float64 {
	if !result.Valid {
		return 0.0
	}

	score := 1.0

	// Deduct for warnings
	score -= float64(len(result.Warnings)) * 0.05

	if score < 0 {
		score = 0
	}

	return score
}

func calculateSemanticScore(result *ValidationResult) float64 {
	score := 1.0

	// Deduct for errors
	score -= float64(len(result.Errors)) * 0.2

	// Deduct for warnings
	score -= float64(len(result.Warnings)) * 0.05

	if score < 0 {
		score = 0
	}

	return score
}

func calculateBusinessRulesScore(result *ValidationResult) float64 {
	score := 1.0

	// Deduct for warnings (business rules are less critical)
	score -= float64(len(result.Warnings)) * 0.03

	if score < 0 {
		score = 0
	}

	return score
}
