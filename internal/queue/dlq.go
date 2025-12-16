package queue

import (
	"sync"
	"time"
)

// DeadLetterQueue represents a dead letter queue for failed jobs
type DeadLetterQueue interface {
	// Add adds a failed job to the DLQ
	Add(job *Job) error

	// Get retrieves a job from the DLQ by ID
	Get(jobID string) (*Job, error)

	// List lists all jobs in the DLQ
	List(limit, offset int) ([]*Job, error)

	// Remove removes a job from the DLQ
	Remove(jobID string) error

	// Retry retries a job from the DLQ
	Retry(jobID string) (*Job, error)

	// GetStats returns DLQ statistics
	GetStats() DLQStats

	// Clear clears all jobs from the DLQ
	Clear() error

	// GetSize returns the current size of the DLQ
	GetSize() int
}

// DLQStats represents dead letter queue statistics
type DLQStats struct {
	TotalJobs     int                `json:"total_jobs"`
	ByType        map[JobType]int    `json:"by_type"`
	ByError       map[string]int     `json:"by_error"`
	OldestJob     *time.Time         `json:"oldest_job,omitempty"`
	NewestJob     *time.Time         `json:"newest_job,omitempty"`
	AvgAttempts   float64            `json:"avg_attempts"`
}

// MemoryDLQ is an in-memory implementation of DeadLetterQueue
type MemoryDLQ struct {
	jobs     map[string]*Job
	jobsList []*Job
	mu       sync.RWMutex
}

// NewMemoryDLQ creates a new in-memory dead letter queue
func NewMemoryDLQ() *MemoryDLQ {
	return &MemoryDLQ{
		jobs:     make(map[string]*Job),
		jobsList: make([]*Job, 0),
	}
}

// Add adds a failed job to the DLQ
func (dlq *MemoryDLQ) Add(job *Job) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	// Check if job already exists
	if _, exists := dlq.jobs[job.ID]; exists {
		// Update existing job
		dlq.jobs[job.ID] = job
		return nil
	}

	// Add new job
	dlq.jobs[job.ID] = job
	dlq.jobsList = append(dlq.jobsList, job)

	return nil
}

// Get retrieves a job from the DLQ by ID
func (dlq *MemoryDLQ) Get(jobID string) (*Job, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	job, ok := dlq.jobs[jobID]
	if !ok {
		return nil, nil
	}

	return job, nil
}

// List lists all jobs in the DLQ
func (dlq *MemoryDLQ) List(limit, offset int) ([]*Job, error) {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	total := len(dlq.jobsList)

	// Handle offset
	if offset >= total {
		return []*Job{}, nil
	}

	start := offset
	end := offset + limit
	if end > total {
		end = total
	}

	if limit == 0 {
		end = total
	}

	return dlq.jobsList[start:end], nil
}

// Remove removes a job from the DLQ
func (dlq *MemoryDLQ) Remove(jobID string) error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	// Remove from map
	delete(dlq.jobs, jobID)

	// Remove from list
	for i, job := range dlq.jobsList {
		if job.ID == jobID {
			dlq.jobsList = append(dlq.jobsList[:i], dlq.jobsList[i+1:]...)
			break
		}
	}

	return nil
}

// Retry retries a job from the DLQ
func (dlq *MemoryDLQ) Retry(jobID string) (*Job, error) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	job, ok := dlq.jobs[jobID]
	if !ok {
		return nil, nil
	}

	// Reset job for retry
	job.Status = JobStatusPending
	job.Error = ""
	job.Attempts = 0
	job.StartedAt = nil
	job.CompletedAt = nil
	job.UpdatedAt = time.Now()

	// Remove from DLQ
	delete(dlq.jobs, jobID)
	for i, j := range dlq.jobsList {
		if j.ID == jobID {
			dlq.jobsList = append(dlq.jobsList[:i], dlq.jobsList[i+1:]...)
			break
		}
	}

	return job, nil
}

// GetStats returns DLQ statistics
func (dlq *MemoryDLQ) GetStats() DLQStats {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	stats := DLQStats{
		TotalJobs: len(dlq.jobs),
		ByType:    make(map[JobType]int),
		ByError:   make(map[string]int),
	}

	if stats.TotalJobs == 0 {
		return stats
	}

	var totalAttempts int
	var oldest, newest *time.Time

	for _, job := range dlq.jobs {
		// Count by type
		stats.ByType[job.Type]++

		// Count by error
		if job.Error != "" {
			stats.ByError[job.Error]++
		}

		// Track oldest and newest
		if oldest == nil || job.CreatedAt.Before(*oldest) {
			oldest = &job.CreatedAt
		}
		if newest == nil || job.CreatedAt.After(*newest) {
			newest = &job.CreatedAt
		}

		// Sum attempts
		totalAttempts += job.Attempts
	}

	stats.OldestJob = oldest
	stats.NewestJob = newest
	stats.AvgAttempts = float64(totalAttempts) / float64(stats.TotalJobs)

	return stats
}

// Clear clears all jobs from the DLQ
func (dlq *MemoryDLQ) Clear() error {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	dlq.jobs = make(map[string]*Job)
	dlq.jobsList = make([]*Job, 0)

	return nil
}

// GetSize returns the current size of the DLQ
func (dlq *MemoryDLQ) GetSize() int {
	dlq.mu.RLock()
	defer dlq.mu.RUnlock()

	return len(dlq.jobs)
}

// PersistentDLQ is a persistent DLQ implementation (placeholder)
// In production, this would use a database or file storage
type PersistentDLQ struct {
	*MemoryDLQ
	// Add persistence layer (database, file, etc.)
}

// NewPersistentDLQ creates a new persistent dead letter queue
func NewPersistentDLQ() *PersistentDLQ {
	return &PersistentDLQ{
		MemoryDLQ: NewMemoryDLQ(),
	}
}
