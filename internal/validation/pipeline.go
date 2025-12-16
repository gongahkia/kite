package validation

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gongahkia/kite/internal/observability"
	"github.com/gongahkia/kite/pkg/models"
)

// ValidationStage represents a stage in the validation pipeline
type ValidationStage string

const (
	StageStructural    ValidationStage = "structural"
	StageSemantic      ValidationStage = "semantic"
	StageBusinessRules ValidationStage = "business_rules"
	StageQuality       ValidationStage = "quality"
	StageDuplication   ValidationStage = "duplication"
)

// ValidationResult represents the result of validation
type ValidationResult struct {
	Valid        bool
	Errors       []ValidationError
	Warnings     []ValidationWarning
	Score        float64
	Completeness float64
	Stage        ValidationStage
	Duration     time.Duration
}

// ValidationError represents a validation error
type ValidationError struct {
	Field   string
	Message string
	Code    string
}

// ValidationWarning represents a validation warning
type ValidationWarning struct {
	Field   string
	Message string
	Code    string
}

// Validator is the interface for validators
type Validator interface {
	Validate(ctx context.Context, c *models.Case) (*ValidationResult, error)
	Stage() ValidationStage
	Name() string
}

// Pipeline represents a multi-stage validation pipeline
type Pipeline struct {
	validators []Validator
	logger     *observability.Logger
	metrics    *observability.Metrics
	concurrent bool
	mu         sync.RWMutex
}

// PipelineConfig configures the validation pipeline
type PipelineConfig struct {
	Concurrent bool
	Stages     []ValidationStage
}

// NewPipeline creates a new validation pipeline
func NewPipeline(logger *observability.Logger, metrics *observability.Metrics, config *PipelineConfig) *Pipeline {
	p := &Pipeline{
		validators: make([]Validator, 0),
		logger:     logger,
		metrics:    metrics,
		concurrent: config.Concurrent,
	}

	// Register validators based on requested stages
	for _, stage := range config.Stages {
		switch stage {
		case StageStructural:
			p.AddValidator(NewStructuralValidator())
		case StageSemantic:
			p.AddValidator(NewSemanticValidator())
		case StageBusinessRules:
			p.AddValidator(NewBusinessRulesValidator())
		case StageQuality:
			p.AddValidator(NewQualityValidator())
		case StageDuplication:
			p.AddValidator(NewDuplicationValidator())
		}
	}

	return p
}

// DefaultPipeline creates a pipeline with all stages
func DefaultPipeline(logger *observability.Logger, metrics *observability.Metrics) *Pipeline {
	return NewPipeline(logger, metrics, &PipelineConfig{
		Concurrent: true,
		Stages: []ValidationStage{
			StageStructural,
			StageSemantic,
			StageBusinessRules,
			StageQuality,
			StageDuplication,
		},
	})
}

// AddValidator adds a validator to the pipeline
func (p *Pipeline) AddValidator(v Validator) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.validators = append(p.validators, v)
}

// Validate runs all validators in the pipeline
func (p *Pipeline) Validate(ctx context.Context, c *models.Case) (*ValidationReport, error) {
	start := time.Now()

	report := &ValidationReport{
		CaseID:    c.ID,
		Timestamp: time.Now(),
		Results:   make([]*ValidationResult, 0),
	}

	if p.concurrent {
		report.Results = p.validateConcurrent(ctx, c)
	} else {
		report.Results = p.validateSequential(ctx, c)
	}

	// Aggregate results
	report.Aggregate()

	duration := time.Since(start)
	report.Duration = duration

	// Record metrics
	status := "valid"
	if !report.Valid {
		status = "invalid"
	}
	p.metrics.ValidationTotal.WithLabelValues("case", status).Inc()

	p.logger.WithFields(map[string]interface{}{
		"case_id":  c.ID,
		"valid":    report.Valid,
		"score":    report.OverallScore,
		"duration": duration.Milliseconds(),
	}).Info("Validation completed")

	return report, nil
}

// validateSequential runs validators one by one
func (p *Pipeline) validateSequential(ctx context.Context, c *models.Case) []*ValidationResult {
	results := make([]*ValidationResult, 0, len(p.validators))

	for _, validator := range p.validators {
		result, err := validator.Validate(ctx, c)
		if err != nil {
			p.logger.WithFields(map[string]interface{}{
				"validator": validator.Name(),
				"error":     err,
			}).Error("Validator failed")
			continue
		}
		results = append(results, result)

		// Stop on critical errors (optional)
		if !result.Valid && len(result.Errors) > 0 {
			// Could add early stopping logic here
		}
	}

	return results
}

// validateConcurrent runs validators concurrently
func (p *Pipeline) validateConcurrent(ctx context.Context, c *models.Case) []*ValidationResult {
	var wg sync.WaitGroup
	resultsChan := make(chan *ValidationResult, len(p.validators))

	for _, validator := range p.validators {
		wg.Add(1)
		go func(v Validator) {
			defer wg.Done()

			result, err := v.Validate(ctx, c)
			if err != nil {
				p.logger.WithFields(map[string]interface{}{
					"validator": v.Name(),
					"error":     err,
				}).Error("Validator failed")
				return
			}

			resultsChan <- result
		}(validator)
	}

	// Wait for all validators to finish
	go func() {
		wg.Wait()
		close(resultsChan)
	}()

	// Collect results
	results := make([]*ValidationResult, 0, len(p.validators))
	for result := range resultsChan {
		results = append(results, result)
	}

	return results
}

// ValidateBatch validates multiple cases
func (p *Pipeline) ValidateBatch(ctx context.Context, cases []*models.Case) ([]*ValidationReport, error) {
	reports := make([]*ValidationReport, len(cases))
	var wg sync.WaitGroup
	sem := make(chan struct{}, 10) // Limit concurrent validations

	for i, c := range cases {
		wg.Add(1)
		go func(idx int, caseData *models.Case) {
			defer wg.Done()
			sem <- struct{}{}        // Acquire
			defer func() { <-sem }() // Release

			report, err := p.Validate(ctx, caseData)
			if err != nil {
				p.logger.WithField("case_id", caseData.ID).Error("Batch validation failed")
				return
			}
			reports[idx] = report
		}(i, c)
	}

	wg.Wait()

	return reports, nil
}

// ValidationReport represents the complete validation report
type ValidationReport struct {
	CaseID        string
	Timestamp     time.Time
	Duration      time.Duration
	Results       []*ValidationResult
	Valid         bool
	OverallScore  float64
	Completeness  float64
	TotalErrors   int
	TotalWarnings int
	Errors        []ValidationError
	Warnings      []ValidationWarning
}

// Aggregate aggregates all validation results
func (r *ValidationReport) Aggregate() {
	r.Valid = true
	totalScore := 0.0
	totalCompleteness := 0.0
	count := 0

	for _, result := range r.Results {
		if !result.Valid {
			r.Valid = false
		}

		totalScore += result.Score
		totalCompleteness += result.Completeness
		count++

		r.Errors = append(r.Errors, result.Errors...)
		r.Warnings = append(r.Warnings, result.Warnings...)
	}

	if count > 0 {
		r.OverallScore = totalScore / float64(count)
		r.Completeness = totalCompleteness / float64(count)
	}

	r.TotalErrors = len(r.Errors)
	r.TotalWarnings = len(r.Warnings)
}

// String returns a string representation of the report
func (r *ValidationReport) String() string {
	status := "VALID"
	if !r.Valid {
		status = "INVALID"
	}

	return fmt.Sprintf("Validation Report [%s] Case: %s | Score: %.2f | Completeness: %.2f%% | Errors: %d | Warnings: %d",
		status, r.CaseID, r.OverallScore, r.Completeness*100, r.TotalErrors, r.TotalWarnings)
}

// HasCriticalErrors checks if there are critical errors
func (r *ValidationReport) HasCriticalErrors() bool {
	for _, err := range r.Errors {
		if err.Code == "CRITICAL" || err.Code == "REQUIRED_FIELD_MISSING" {
			return true
		}
	}
	return false
}

// ShouldReject determines if the case should be rejected
func (r *ValidationReport) ShouldReject() bool {
	// Reject if:
	// 1. Has critical errors
	// 2. Overall score below threshold (0.5)
	// 3. Completeness below threshold (0.6)
	return r.HasCriticalErrors() || r.OverallScore < 0.5 || r.Completeness < 0.6
}
