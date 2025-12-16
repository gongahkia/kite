package validation

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/gongahkia/kite/pkg/errors"
	"github.com/gongahkia/kite/pkg/models"
)

// Validator provides validation functionality for data models
type Validator struct {
	validate *validator.Validate
}

// NewValidator creates a new Validator
func NewValidator() *Validator {
	v := validator.New()

	// Register custom validators
	v.RegisterValidation("case_number", validateCaseNumber)
	v.RegisterValidation("url_accessible", validateURLAccessible)

	return &Validator{
		validate: v,
	}
}

// ValidateCase validates a Case model
func (v *Validator) ValidateCase(c *models.Case) error {
	if err := v.validate.Struct(c); err != nil {
		return errors.ValidationError("case validation failed", err)
	}

	// Additional business rule validations
	if c.DecisionDate != nil && c.FilingDate != nil {
		if c.DecisionDate.Before(*c.FilingDate) {
			return errors.ValidationError("decision date cannot be before filing date", nil)
		}
	}

	// Validate court level
	if c.CourtLevel < 1 || c.CourtLevel > 5 {
		return errors.ValidationError("invalid court level", nil)
	}

	return nil
}

// ValidateCitation validates a Citation model
func (v *Validator) ValidateCitation(c *models.Citation) error {
	if err := v.validate.Struct(c); err != nil {
		return errors.ValidationError("citation validation failed", err)
	}

	// Validate year range
	if c.CaseYear != 0 && (c.CaseYear < 1600 || c.CaseYear > time.Now().Year()+1) {
		return errors.ValidationError("invalid case year", nil)
	}

	return nil
}

// ValidateJudge validates a Judge model
func (v *Validator) ValidateJudge(j *models.Judge) error {
	if err := v.validate.Struct(j); err != nil {
		return errors.ValidationError("judge validation failed", err)
	}

	// Validate appointment and retirement dates
	if j.AppointmentDate != nil && j.RetirementDate != nil {
		if j.RetirementDate.Before(*j.AppointmentDate) {
			return errors.ValidationError("retirement date cannot be before appointment date", nil)
		}
	}

	return nil
}

// ValidateConcept validates a LegalConcept model
func (v *Validator) ValidateConcept(lc *models.LegalConcept) error {
	if err := v.validate.Struct(lc); err != nil {
		return errors.ValidationError("legal concept validation failed", err)
	}

	if len(lc.Keywords) == 0 {
		return errors.ValidationError("legal concept must have at least one keyword", nil)
	}

	return nil
}

// Custom validation functions

func validateCaseNumber(fl validator.FieldLevel) bool {
	caseNum := fl.Field().String()
	if caseNum == "" {
		return false
	}

	// Basic validation: at least 3 characters, contains alphanumeric
	alphanumericPattern := regexp.MustCompile(`[a-zA-Z0-9]`)
	return len(caseNum) >= 3 && alphanumericPattern.MatchString(caseNum)
}

func validateURLAccessible(fl validator.FieldLevel) bool {
	url := fl.Field().String()
	// Basic URL format validation
	urlPattern := regexp.MustCompile(`^https?://`)
	return urlPattern.MatchString(url)
}

// QualityScorer calculates quality scores for cases
type QualityScorer struct{}

// NewQualityScorer creates a new QualityScorer
func NewQualityScorer() *QualityScorer {
	return &QualityScorer{}
}

// ScoreCase calculates a quality score for a case based on completeness
func (qs *QualityScorer) ScoreCase(c *models.Case) float64 {
	score := 0.0
	maxScore := 0.0

	// Required fields (already validated)
	requiredFields := []bool{
		c.ID != "",
		c.CaseNumber != "",
		c.CaseName != "",
		c.DecisionDate != nil,
		c.Court != "",
		c.Jurisdiction != "",
		c.URL != "",
	}

	for _, present := range requiredFields {
		maxScore += 1.0
		if present {
			score += 1.0
		}
	}

	// Optional but valuable fields
	optionalFields := []bool{
		c.FilingDate != nil,
		c.Summary != "",
		c.FullText != "",
		len(c.Judges) > 0,
		len(c.Citations) > 0,
		len(c.LegalConcepts) > 0,
		c.Outcome != "",
		c.ChiefJudge != "",
		len(c.Parties) > 0,
	}

	for _, present := range optionalFields {
		maxScore += 0.5
		if present {
			score += 0.5
		}
	}

	// Normalize to 0-1 range
	if maxScore == 0 {
		return 0.0
	}

	return score / maxScore
}

// DeduplicationService handles duplicate detection
type DeduplicationService struct {
	seenHashes map[string]bool
}

// NewDeduplicationService creates a new DeduplicationService
func NewDeduplicationService() *DeduplicationService {
	return &DeduplicationService{
		seenHashes: make(map[string]bool),
	}
}

// ComputeCaseHash computes a hash for a case to detect duplicates
func (ds *DeduplicationService) ComputeCaseHash(c *models.Case) string {
	// Create a unique string from key fields
	hashInput := fmt.Sprintf("%s|%s|%s|%s",
		strings.ToLower(c.CaseNumber),
		strings.ToLower(c.CaseName),
		c.Court,
		c.Jurisdiction,
	)

	if c.DecisionDate != nil {
		hashInput += "|" + c.DecisionDate.Format("2006-01-02")
	}

	hash := sha256.Sum256([]byte(hashInput))
	return hex.EncodeToString(hash[:])
}

// IsDuplicate checks if a case hash has been seen before
func (ds *DeduplicationService) IsDuplicate(hash string) bool {
	if _, exists := ds.seenHashes[hash]; exists {
		return true
	}
	ds.seenHashes[hash] = true
	return false
}

// Reset clears the deduplication cache
func (ds *DeduplicationService) Reset() {
	ds.seenHashes = make(map[string]bool)
}

// CompletenessChecker checks data completeness
type CompletenessChecker struct{}

// NewCompletenessChecker creates a new CompletenessChecker
func NewCompletenessChecker() *CompletenessChecker {
	return &CompletenessChecker{}
}

// CheckCaseCompleteness checks if a case has all essential fields
func (cc *CompletenessChecker) CheckCaseCompleteness(c *models.Case) (bool, []string) {
	missingFields := []string{}

	if c.ID == "" {
		missingFields = append(missingFields, "id")
	}
	if c.CaseNumber == "" {
		missingFields = append(missingFields, "case_number")
	}
	if c.CaseName == "" {
		missingFields = append(missingFields, "case_name")
	}
	if c.DecisionDate == nil {
		missingFields = append(missingFields, "decision_date")
	}
	if c.Court == "" {
		missingFields = append(missingFields, "court")
	}
	if c.Jurisdiction == "" {
		missingFields = append(missingFields, "jurisdiction")
	}
	if c.URL == "" {
		missingFields = append(missingFields, "url")
	}

	return len(missingFields) == 0, missingFields
}
