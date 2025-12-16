package worker

import (
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gongahkia/kite/internal/queue"
)

// Worker represents a single worker
type Worker struct {
	id             int
	queue          queue.Queue
	handler        JobHandler
	isBusy         atomic.Bool
	jobsProcessed  atomic.Int64
	jobsFailed     atomic.Int64
	totalDuration  atomic.Int64
	currentJob     *queue.Job
	mu             sync.RWMutex
}

// NewWorker creates a new worker
func NewWorker(id int, q queue.Queue, handler JobHandler) *Worker {
	return &Worker{
		id:      id,
		queue:   q,
		handler: handler,
	}
}

// Run starts the worker loop
func (w *Worker) Run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Try to get a job from the queue
			job, err := w.queue.Dequeue(ctx)
			if err != nil {
				if ctx.Err() != nil {
					// Context cancelled, exit
					return
				}
				// Queue error, wait and retry
				time.Sleep(100 * time.Millisecond)
				continue
			}

			// Process the job
			w.processJob(ctx, job)
		}
	}
}

// processJob processes a single job
func (w *Worker) processJob(ctx context.Context, job *queue.Job) {
	// Mark worker as busy
	w.isBusy.Store(true)
	w.setCurrentJob(job)
	defer func() {
		w.isBusy.Store(false)
		w.setCurrentJob(nil)
	}()

	// Track job duration
	startTime := time.Now()
	defer func() {
		duration := time.Since(startTime)
		w.totalDuration.Add(int64(duration))
	}()

	// Create job context with timeout
	jobCtx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	// Execute the job handler
	err := w.handler(jobCtx, job)

	if err != nil {
		// Job failed
		w.jobsFailed.Add(1)
		job.MarkFailed(err)

		// Nack the job (requeue if retries available)
		if nackErr := w.queue.Nack(ctx, job.ID, job.ShouldRetry()); nackErr != nil {
			// Log error
			fmt.Printf("Worker %d: Failed to nack job %s: %v\n", w.id, job.ID, nackErr)
		}
	} else {
		// Job succeeded
		w.jobsProcessed.Add(1)
		job.MarkCompleted(nil)

		// Ack the job
		if ackErr := w.queue.Ack(ctx, job.ID); ackErr != nil {
			// Log error
			fmt.Printf("Worker %d: Failed to ack job %s: %v\n", w.id, job.ID, ackErr)
		}
	}
}

// GetID returns the worker ID
func (w *Worker) GetID() int {
	return w.id
}

// IsBusy returns true if the worker is currently processing a job
func (w *Worker) IsBusy() bool {
	return w.isBusy.Load()
}

// GetCurrentJob returns the current job being processed
func (w *Worker) GetCurrentJob() *queue.Job {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.currentJob
}

// setCurrentJob sets the current job
func (w *Worker) setCurrentJob(job *queue.Job) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.currentJob = job
}

// GetStats returns worker statistics
func (w *Worker) GetStats() WorkerStats {
	jobsProcessed := w.jobsProcessed.Load()
	totalDuration := time.Duration(w.totalDuration.Load())

	var avgDuration time.Duration
	if jobsProcessed > 0 {
		avgDuration = totalDuration / time.Duration(jobsProcessed)
	}

	return WorkerStats{
		WorkerID:           w.id,
		IsBusy:             w.IsBusy(),
		JobsProcessed:      jobsProcessed,
		JobsFailed:         w.jobsFailed.Load(),
		AverageJobDuration: avgDuration,
		CurrentJob:         w.GetCurrentJob(),
	}
}

// WorkerStats represents worker statistics
type WorkerStats struct {
	WorkerID           int           `json:"worker_id"`
	IsBusy             bool          `json:"is_busy"`
	JobsProcessed      int64         `json:"jobs_processed"`
	JobsFailed         int64         `json:"jobs_failed"`
	AverageJobDuration time.Duration `json:"average_job_duration"`
	CurrentJob         *queue.Job    `json:"current_job,omitempty"`
}
