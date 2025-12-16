package worker

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gongahkia/kite/internal/queue"
)

// Pool represents a pool of workers
type Pool struct {
	workers    []*Worker
	queue      queue.Queue
	handler    JobHandler
	mu         sync.RWMutex
	wg         sync.WaitGroup
	ctx        context.Context
	cancel     context.CancelFunc
	shutdownCh chan struct{}
	logger     interface{}
	metrics    interface{}
}

// JobHandler is a function that handles a job
type JobHandler func(ctx context.Context, job *queue.Job) error

// PoolConfig represents worker pool configuration
type PoolConfig struct {
	WorkerCount    int
	JobTimeout     time.Duration
	ShutdownGrace  time.Duration
}

// NewPool creates a new worker pool
func NewPool(cfg PoolConfig, q queue.Queue, handler JobHandler) *Pool {
	ctx, cancel := context.WithCancel(context.Background())

	return &Pool{
		workers:    make([]*Worker, 0, cfg.WorkerCount),
		queue:      q,
		handler:    handler,
		ctx:        ctx,
		cancel:     cancel,
		shutdownCh: make(chan struct{}),
	}
}

// Start starts all workers in the pool
func (p *Pool) Start(workerCount int) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i := 0; i < workerCount; i++ {
		worker := NewWorker(i, p.queue, p.handler)
		p.workers = append(p.workers, worker)

		p.wg.Add(1)
		go func(w *Worker) {
			defer p.wg.Done()
			w.Run(p.ctx)
		}(worker)
	}

	return nil
}

// Stop gracefully stops all workers
func (p *Pool) Stop(timeout time.Duration) error {
	// Cancel context to signal workers to stop
	p.cancel()

	// Wait for workers to finish with timeout
	done := make(chan struct{})
	go func() {
		p.wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// All workers stopped gracefully
		return nil
	case <-time.After(timeout):
		// Timeout reached
		return fmt.Errorf("worker pool shutdown timeout after %v", timeout)
	}
}

// GetWorkerCount returns the number of workers
func (p *Pool) GetWorkerCount() int {
	p.mu.RLock()
	defer p.mu.RUnlock()
	return len(p.workers)
}

// GetStats returns worker pool statistics
func (p *Pool) GetStats() PoolStats {
	p.mu.RLock()
	defer p.mu.RUnlock()

	stats := PoolStats{
		WorkerCount: len(p.workers),
	}

	for _, worker := range p.workers {
		workerStats := worker.GetStats()
		stats.TotalJobsProcessed += workerStats.JobsProcessed
		stats.TotalJobsFailed += workerStats.JobsFailed
		if workerStats.IsBusy {
			stats.BusyWorkers++
		}
		if workerStats.AverageJobDuration > stats.AverageJobDuration {
			stats.AverageJobDuration = workerStats.AverageJobDuration
		}
	}

	if stats.WorkerCount > 0 {
		stats.Utilization = float64(stats.BusyWorkers) / float64(stats.WorkerCount)
	}

	return stats
}

// PoolStats represents worker pool statistics
type PoolStats struct {
	WorkerCount        int           `json:"worker_count"`
	BusyWorkers        int           `json:"busy_workers"`
	TotalJobsProcessed int64         `json:"total_jobs_processed"`
	TotalJobsFailed    int64         `json:"total_jobs_failed"`
	Utilization        float64       `json:"utilization"`
	AverageJobDuration time.Duration `json:"average_job_duration"`
}
