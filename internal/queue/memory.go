package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gongahkia/kite/pkg/errors"
)

// MemoryQueue is an in-memory implementation of the Queue interface
type MemoryQueue struct {
	jobs     []*Job
	jobsMap  map[string]*Job
	mu       sync.RWMutex
	notEmpty chan struct{}
	closed   bool
}

// NewMemoryQueue creates a new MemoryQueue
func NewMemoryQueue() *MemoryQueue {
	return &MemoryQueue{
		jobs:     make([]*Job, 0),
		jobsMap:  make(map[string]*Job),
		notEmpty: make(chan struct{}, 1),
		closed:   false,
	}
}

// Enqueue adds a job to the queue
func (mq *MemoryQueue) Enqueue(ctx context.Context, job *Job) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	if mq.closed {
		return errors.QueueError("queue is closed", errors.ErrQueueFull)
	}

	// Add to map
	mq.jobsMap[job.ID] = job

	// Add to queue based on priority
	mq.jobs = append(mq.jobs, job)
	mq.sortByPriority()

	// Signal that queue is not empty
	select {
	case mq.notEmpty <- struct{}{}:
	default:
	}

	return nil
}

// Dequeue retrieves the next job from the queue
func (mq *MemoryQueue) Dequeue(ctx context.Context) (*Job, error) {
	for {
		mq.mu.Lock()

		// Check if closed
		if mq.closed && len(mq.jobs) == 0 {
			mq.mu.Unlock()
			return nil, errors.QueueError("queue is closed", errors.ErrQueueEmpty)
		}

		// Try to get a job
		if len(mq.jobs) > 0 {
			// Get highest priority job
			job := mq.jobs[0]
			mq.jobs = mq.jobs[1:]
			mq.mu.Unlock()

			// Mark as started
			job.MarkStarted()
			return job, nil
		}

		mq.mu.Unlock()

		// Wait for a job or context cancellation
		select {
		case <-mq.notEmpty:
			continue
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Ack acknowledges successful completion of a job
func (mq *MemoryQueue) Ack(ctx context.Context, jobID string) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	job, ok := mq.jobsMap[jobID]
	if !ok {
		return errors.QueueError("job not found", errors.ErrNotFound)
	}

	job.MarkCompleted(nil)
	delete(mq.jobsMap, jobID)

	return nil
}

// Nack negatively acknowledges a job
func (mq *MemoryQueue) Nack(ctx context.Context, jobID string, requeue bool) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	job, ok := mq.jobsMap[jobID]
	if !ok {
		return errors.QueueError("job not found", errors.ErrNotFound)
	}

	if requeue && job.ShouldRetry() {
		// Re-add to queue
		mq.jobs = append(mq.jobs, job)
		mq.sortByPriority()

		// Signal that queue is not empty
		select {
		case mq.notEmpty <- struct{}{}:
		default:
		}
	} else {
		// Mark as failed and remove
		job.MarkFailed(fmt.Errorf("job failed after %d attempts", job.Attempts))
		delete(mq.jobsMap, jobID)
	}

	return nil
}

// GetDepth returns the current queue depth
func (mq *MemoryQueue) GetDepth(ctx context.Context) (int, error) {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	return len(mq.jobs), nil
}

// Close closes the queue
func (mq *MemoryQueue) Close() error {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.closed = true
	close(mq.notEmpty)

	return nil
}

// sortByPriority sorts jobs by priority (highest first)
func (mq *MemoryQueue) sortByPriority() {
	// Simple bubble sort for small queues
	n := len(mq.jobs)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if mq.jobs[j].Priority < mq.jobs[j+1].Priority {
				mq.jobs[j], mq.jobs[j+1] = mq.jobs[j+1], mq.jobs[j]
			}
		}
	}
}

// GetStats returns queue statistics
func (mq *MemoryQueue) GetStats() QueueStats {
	mq.mu.RLock()
	defer mq.mu.RUnlock()

	stats := QueueStats{
		Depth: len(mq.jobs),
	}

	// Count jobs by status
	for _, job := range mq.jobsMap {
		switch job.Status {
		case JobStatusPending, JobStatusRetrying:
			stats.Pending++
		case JobStatusRunning:
			stats.Running++
		case JobStatusCompleted:
			stats.Completed++
		case JobStatusFailed:
			stats.Failed++
		}

		// Track last times
		if job.CreatedAt.After(stats.LastEnqueued) {
			stats.LastEnqueued = job.CreatedAt
		}
		if job.StartedAt != nil && job.StartedAt.After(stats.LastDequeued) {
			stats.LastDequeued = *job.StartedAt
		}
	}

	return stats
}

// Clear clears all jobs from the queue
func (mq *MemoryQueue) Clear() {
	mq.mu.Lock()
	defer mq.mu.Unlock()

	mq.jobs = make([]*Job, 0)
	mq.jobsMap = make(map[string]*Job)
}
