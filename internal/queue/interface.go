package queue

import (
	"context"
	"time"
)

// Queue defines the interface for job queue implementations
type Queue interface {
	// Enqueue adds a job to the queue
	Enqueue(ctx context.Context, job *Job) error

	// Dequeue retrieves the next job from the queue
	Dequeue(ctx context.Context) (*Job, error)

	// Ack acknowledges successful completion of a job
	Ack(ctx context.Context, jobID string) error

	// Nack negatively acknowledges a job (requeue or send to DLQ)
	Nack(ctx context.Context, jobID string, requeue bool) error

	// GetDepth returns the current queue depth
	GetDepth(ctx context.Context) (int, error)

	// Close closes the queue connection
	Close() error
}

// Job represents a job in the queue
type Job struct {
	ID          string                 `json:"id"`
	Type        JobType                `json:"type"`
	Priority    Priority               `json:"priority"`
	Payload     map[string]interface{} `json:"payload"`
	Status      JobStatus              `json:"status"`
	CreatedAt   time.Time              `json:"created_at"`
	UpdatedAt   time.Time              `json:"updated_at"`
	ScheduledAt *time.Time             `json:"scheduled_at,omitempty"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
	Attempts    int                    `json:"attempts"`
	MaxAttempts int                    `json:"max_attempts"`
	Error       string                 `json:"error,omitempty"`
	Result      map[string]interface{} `json:"result,omitempty"`
}

// JobType represents the type of job
type JobType string

const (
	JobTypeScrape     JobType = "scrape"
	JobTypeExtract    JobType = "extract"
	JobTypeValidate   JobType = "validate"
	JobTypeAnalyze    JobType = "analyze"
	JobTypeExport     JobType = "export"
	JobTypeCleanup    JobType = "cleanup"
)

// Priority represents job priority
type Priority int

const (
	PriorityLow    Priority = 1
	PriorityNormal Priority = 2
	PriorityHigh   Priority = 3
)

// JobStatus represents the status of a job
type JobStatus string

const (
	JobStatusPending    JobStatus = "pending"
	JobStatusRunning    JobStatus = "running"
	JobStatusCompleted  JobStatus = "completed"
	JobStatusFailed     JobStatus = "failed"
	JobStatusRetrying   JobStatus = "retrying"
	JobStatusCancelled  JobStatus = "cancelled"
)

// NewJob creates a new job
func NewJob(jobType JobType, payload map[string]interface{}) *Job {
	now := time.Now()
	return &Job{
		ID:          generateJobID(),
		Type:        jobType,
		Priority:    PriorityNormal,
		Payload:     payload,
		Status:      JobStatusPending,
		CreatedAt:   now,
		UpdatedAt:   now,
		Attempts:    0,
		MaxAttempts: 3,
	}
}

// MarkStarted marks the job as started
func (j *Job) MarkStarted() {
	now := time.Now()
	j.Status = JobStatusRunning
	j.StartedAt = &now
	j.UpdatedAt = now
	j.Attempts++
}

// MarkCompleted marks the job as completed
func (j *Job) MarkCompleted(result map[string]interface{}) {
	now := time.Now()
	j.Status = JobStatusCompleted
	j.CompletedAt = &now
	j.UpdatedAt = now
	j.Result = result
}

// MarkFailed marks the job as failed
func (j *Job) MarkFailed(err error) {
	now := time.Now()
	j.UpdatedAt = now

	if j.Attempts >= j.MaxAttempts {
		j.Status = JobStatusFailed
	} else {
		j.Status = JobStatusRetrying
	}

	if err != nil {
		j.Error = err.Error()
	}
}

// ShouldRetry returns true if the job should be retried
func (j *Job) ShouldRetry() bool {
	return j.Attempts < j.MaxAttempts && j.Status == JobStatusRetrying
}

// SetPriority sets the job priority
func (j *Job) SetPriority(priority Priority) {
	j.Priority = priority
	j.UpdatedAt = time.Now()
}

// Schedule schedules the job for a future time
func (j *Job) Schedule(scheduledAt time.Time) {
	j.ScheduledAt = &scheduledAt
	j.UpdatedAt = time.Now()
}

// IsScheduled returns true if the job is scheduled for the future
func (j *Job) IsScheduled() bool {
	return j.ScheduledAt != nil && j.ScheduledAt.After(time.Now())
}

// generateJobID generates a unique job ID
func generateJobID() string {
	return time.Now().Format("20060102150405") + "-" + randString(8)
}

// randString generates a random string of given length
func randString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyz0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[time.Now().UnixNano()%int64(len(letters))]
	}
	return string(b)
}

// QueueConfig represents queue configuration
type QueueConfig struct {
	Driver      string        `yaml:"driver"`
	URL         string        `yaml:"url"`
	MaxRetries  int           `yaml:"max_retries"`
	RetryDelay  time.Duration `yaml:"retry_delay"`
	Concurrency int           `yaml:"concurrency"`
}

// QueueStats represents queue statistics
type QueueStats struct {
	Depth         int       `json:"depth"`
	Pending       int       `json:"pending"`
	Running       int       `json:"running"`
	Completed     int       `json:"completed"`
	Failed        int       `json:"failed"`
	LastEnqueued  time.Time `json:"last_enqueued"`
	LastDequeued  time.Time `json:"last_dequeued"`
}
