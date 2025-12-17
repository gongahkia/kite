package websocket

import (
	"github.com/gongahkia/kite/pkg/models"
)

// EventEmitter handles emitting WebSocket events
type EventEmitter struct {
	server *Server
}

// NewEventEmitter creates a new event emitter
func NewEventEmitter(server *Server) *EventEmitter {
	return &EventEmitter{
		server: server,
	}
}

// EmitScrapeStarted emits an event when scraping starts
func (e *EventEmitter) EmitScrapeStarted(jobID, jurisdiction string) {
	msg := NewMessage(MessageTypeScrapeStarted, map[string]interface{}{
		"job_id":       jobID,
		"jurisdiction": jurisdiction,
		"status":       "started",
	})

	e.server.BroadcastToRoom("scrape:"+jobID, msg)
	e.server.BroadcastToRoom("scrape:all", msg)
}

// EmitScrapeProgress emits scraping progress updates
func (e *EventEmitter) EmitScrapeProgress(jobID string, current, total int, message string) {
	msg := NewMessage(MessageTypeScrapeProgress, map[string]interface{}{
		"job_id":  jobID,
		"current": current,
		"total":   total,
		"percent": float64(current) / float64(total) * 100,
		"message": message,
	})

	e.server.BroadcastToRoom("scrape:"+jobID, msg)
	e.server.BroadcastToRoom("scrape:all", msg)
}

// EmitScrapeComplete emits an event when scraping completes
func (e *EventEmitter) EmitScrapeComplete(jobID string, casesScraped int, duration float64) {
	msg := NewMessage(MessageTypeScrapeComplete, map[string]interface{}{
		"job_id":        jobID,
		"cases_scraped": casesScraped,
		"duration_seconds": duration,
		"status":        "completed",
	})

	e.server.BroadcastToRoom("scrape:"+jobID, msg)
	e.server.BroadcastToRoom("scrape:all", msg)
}

// EmitScrapeError emits an error event during scraping
func (e *EventEmitter) EmitScrapeError(jobID, errorMsg string) {
	msg := NewMessage(MessageTypeScrapeError, map[string]interface{}{
		"job_id": jobID,
		"error":  errorMsg,
		"status": "failed",
	})

	e.server.BroadcastToRoom("scrape:"+jobID, msg)
	e.server.BroadcastToRoom("scrape:all", msg)
}

// EmitSearchResults emits search results
func (e *EventEmitter) EmitSearchResults(searchID string, results []*models.Case, total int) {
	// Convert cases to simple map for JSON
	casesData := make([]map[string]interface{}, len(results))
	for i, c := range results {
		casesData[i] = map[string]interface{}{
			"id":           c.ID,
			"case_name":    c.CaseName,
			"case_number":  c.CaseNumber,
			"court":        c.Court,
			"jurisdiction": c.Jurisdiction,
		}
	}

	msg := NewMessage(MessageTypeSearchResults, map[string]interface{}{
		"search_id":    searchID,
		"results":      casesData,
		"result_count": len(results),
		"total":        total,
	})

	e.server.BroadcastToRoom("search:"+searchID, msg)
}

// EmitCaseCreated emits an event when a case is created
func (e *EventEmitter) EmitCaseCreated(c *models.Case) {
	msg := NewMessage(MessageTypeCaseCreated, map[string]interface{}{
		"case_id":      c.ID,
		"case_name":    c.CaseName,
		"case_number":  c.CaseNumber,
		"court":        c.Court,
		"jurisdiction": c.Jurisdiction,
	})

	e.server.BroadcastToRoom("cases:all", msg)
	e.server.BroadcastToRoom("cases:"+c.Jurisdiction, msg)
}

// EmitCaseUpdated emits an event when a case is updated
func (e *EventEmitter) EmitCaseUpdated(c *models.Case) {
	msg := NewMessage(MessageTypeCaseUpdated, map[string]interface{}{
		"case_id":      c.ID,
		"case_name":    c.CaseName,
		"case_number":  c.CaseNumber,
	})

	e.server.BroadcastToRoom("cases:all", msg)
	e.server.BroadcastToRoom("case:"+c.ID, msg)
}

// EmitCaseDeleted emits an event when a case is deleted
func (e *EventEmitter) EmitCaseDeleted(caseID string) {
	msg := NewMessage(MessageTypeCaseDeleted, map[string]interface{}{
		"case_id": caseID,
	})

	e.server.BroadcastToRoom("cases:all", msg)
	e.server.BroadcastToRoom("case:"+caseID, msg)
}

// EmitValidationComplete emits validation completion
func (e *EventEmitter) EmitValidationComplete(caseID string, valid bool, qualityScore float64) {
	msg := NewMessage(MessageTypeValidationComplete, map[string]interface{}{
		"case_id":       caseID,
		"valid":         valid,
		"quality_score": qualityScore,
	})

	e.server.BroadcastToRoom("validation:all", msg)
	e.server.BroadcastToRoom("case:"+caseID, msg)
}

// EmitQualityAlert emits a data quality alert
func (e *EventEmitter) EmitQualityAlert(severity, message string, details map[string]interface{}) {
	msg := NewMessage(MessageTypeQualityAlert, map[string]interface{}{
		"severity": severity,
		"message":  message,
		"details":  details,
	})

	e.server.BroadcastToRoom("alerts:quality", msg)
	e.server.BroadcastToRoom("alerts:all", msg)
}

// EmitWorkerStatus emits worker pool status updates
func (e *EventEmitter) EmitWorkerStatus(activeWorkers, totalWorkers, queueSize int) {
	msg := NewMessage(MessageTypeWorkerStatus, map[string]interface{}{
		"active_workers": activeWorkers,
		"total_workers":  totalWorkers,
		"queue_size":     queueSize,
		"utilization":    float64(activeWorkers) / float64(totalWorkers),
	})

	e.server.BroadcastToRoom("workers:all", msg)
}

// EmitQueueUpdate emits job queue updates
func (e *EventEmitter) EmitQueueUpdate(pending, running, completed, failed int) {
	msg := NewMessage(MessageTypeQueueUpdate, map[string]interface{}{
		"pending":   pending,
		"running":   running,
		"completed": completed,
		"failed":    failed,
		"total":     pending + running,
	})

	e.server.BroadcastToRoom("queue:all", msg)
}

// EmitSystemAlert emits system-level alerts
func (e *EventEmitter) EmitSystemAlert(severity, component, message string) {
	msg := NewMessage(MessageTypeSystemAlert, map[string]interface{}{
		"severity":  severity,
		"component": component,
		"message":   message,
	})

	e.server.BroadcastToRoom("alerts:system", msg)
	e.server.BroadcastToRoom("alerts:all", msg)
}

// EmitMetricUpdate emits metric updates
func (e *EventEmitter) EmitMetricUpdate(metricName string, value float64, labels map[string]string) {
	msg := NewMessage(MessageTypeMetricUpdate, map[string]interface{}{
		"metric": metricName,
		"value":  value,
		"labels": labels,
	})

	e.server.BroadcastToRoom("metrics:all", msg)
	e.server.BroadcastToRoom("metrics:"+metricName, msg)
}
