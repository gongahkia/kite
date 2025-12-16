package events

import (
	"time"

	"github.com/gongahkia/kite/pkg/models"
)

// EventType represents the type of event
type EventType string

const (
	// Scraping events
	EventScrapeStarted   EventType = "scrape.started"
	EventScrapeCompleted EventType = "scrape.completed"
	EventScrapeFailed    EventType = "scrape.failed"

	// Validation events
	EventCaseValidated   EventType = "case.validated"
	EventValidationFailed EventType = "validation.failed"

	// Citation events
	EventCitationExtracted EventType = "citation.extracted"

	// Concept events
	EventConceptExtracted EventType = "concept.extracted"

	// Quality events
	EventQualityChecked EventType = "quality.checked"

	// Storage events
	EventCaseCreated EventType = "case.created"
	EventCaseUpdated EventType = "case.updated"
	EventCaseDeleted EventType = "case.deleted"

	// Worker events
	EventWorkerStarted EventType = "worker.started"
	EventWorkerStopped EventType = "worker.stopped"
	EventJobQueued     EventType = "job.queued"
	EventJobCompleted  EventType = "job.completed"
	EventJobFailed     EventType = "job.failed"
)

// Event represents a system event
type Event struct {
	ID        string                 `json:"id"`
	Type      EventType              `json:"type"`
	Timestamp time.Time              `json:"timestamp"`
	Source    string                 `json:"source"`
	Data      map[string]interface{} `json:"data"`
}

// NewEvent creates a new event
func NewEvent(eventType EventType, source string, data map[string]interface{}) *Event {
	return &Event{
		ID:        generateEventID(),
		Type:      eventType,
		Timestamp: time.Now(),
		Source:    source,
		Data:      data,
	}
}

// ScrapeStartedEvent creates a scrape started event
func ScrapeStartedEvent(scraper string, query interface{}) *Event {
	return NewEvent(EventScrapeStarted, scraper, map[string]interface{}{
		"scraper": scraper,
		"query":   query,
	})
}

// ScrapeCompletedEvent creates a scrape completed event
func ScrapeCompletedEvent(scraper string, casesCount int, duration time.Duration) *Event {
	return NewEvent(EventScrapeCompleted, scraper, map[string]interface{}{
		"scraper":     scraper,
		"cases_count": casesCount,
		"duration_ms": duration.Milliseconds(),
	})
}

// ScrapeFailedEvent creates a scrape failed event
func ScrapeFailedEvent(scraper string, err error) *Event {
	return NewEvent(EventScrapeFailed, scraper, map[string]interface{}{
		"scraper": scraper,
		"error":   err.Error(),
	})
}

// CaseValidatedEvent creates a case validated event
func CaseValidatedEvent(caseID string, score float64) *Event {
	return NewEvent(EventCaseValidated, "validator", map[string]interface{}{
		"case_id": caseID,
		"score":   score,
	})
}

// ValidationFailedEvent creates a validation failed event
func ValidationFailedEvent(caseID string, errors []string) *Event {
	return NewEvent(EventValidationFailed, "validator", map[string]interface{}{
		"case_id": caseID,
		"errors":  errors,
	})
}

// CitationExtractedEvent creates a citation extracted event
func CitationExtractedEvent(caseID string, citationCount int) *Event {
	return NewEvent(EventCitationExtracted, "citation-extractor", map[string]interface{}{
		"case_id":        caseID,
		"citation_count": citationCount,
	})
}

// ConceptExtractedEvent creates a concept extracted event
func ConceptExtractedEvent(caseID string, concepts []string) *Event {
	return NewEvent(EventConceptExtracted, "concept-extractor", map[string]interface{}{
		"case_id":  caseID,
		"concepts": concepts,
	})
}

// QualityCheckedEvent creates a quality checked event
func QualityCheckedEvent(caseID string, qualityScore float64) *Event {
	return NewEvent(EventQualityChecked, "quality-checker", map[string]interface{}{
		"case_id":       caseID,
		"quality_score": qualityScore,
	})
}

// CaseCreatedEvent creates a case created event
func CaseCreatedEvent(c *models.Case) *Event {
	return NewEvent(EventCaseCreated, "storage", map[string]interface{}{
		"case_id":      c.ID,
		"case_name":    c.CaseName,
		"jurisdiction": c.Jurisdiction,
	})
}

// CaseUpdatedEvent creates a case updated event
func CaseUpdatedEvent(c *models.Case) *Event {
	return NewEvent(EventCaseUpdated, "storage", map[string]interface{}{
		"case_id":   c.ID,
		"case_name": c.CaseName,
	})
}

// CaseDeletedEvent creates a case deleted event
func CaseDeletedEvent(caseID string) *Event {
	return NewEvent(EventCaseDeleted, "storage", map[string]interface{}{
		"case_id": caseID,
	})
}

// WorkerStartedEvent creates a worker started event
func WorkerStartedEvent(workerID string) *Event {
	return NewEvent(EventWorkerStarted, "worker-pool", map[string]interface{}{
		"worker_id": workerID,
	})
}

// WorkerStoppedEvent creates a worker stopped event
func WorkerStoppedEvent(workerID string) *Event {
	return NewEvent(EventWorkerStopped, "worker-pool", map[string]interface{}{
		"worker_id": workerID,
	})
}

// JobQueuedEvent creates a job queued event
func JobQueuedEvent(jobID string, jobType string) *Event {
	return NewEvent(EventJobQueued, "queue", map[string]interface{}{
		"job_id":   jobID,
		"job_type": jobType,
	})
}

// JobCompletedEvent creates a job completed event
func JobCompletedEvent(jobID string, duration time.Duration) *Event {
	return NewEvent(EventJobCompleted, "worker", map[string]interface{}{
		"job_id":      jobID,
		"duration_ms": duration.Milliseconds(),
	})
}

// JobFailedEvent creates a job failed event
func JobFailedEvent(jobID string, err error) *Event {
	return NewEvent(EventJobFailed, "worker", map[string]interface{}{
		"job_id": jobID,
		"error":  err.Error(),
	})
}

// generateEventID generates a unique event ID
func generateEventID() string {
	return time.Now().Format("20060102150405.000000000")
}
