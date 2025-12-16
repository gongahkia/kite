package queue

import (
	"sync"
	"time"
)

// QueueMetrics tracks queue performance metrics
type QueueMetrics struct {
	mu sync.RWMutex

	// Job counts
	TotalEnqueued  int64
	TotalDequeued  int64
	TotalCompleted int64
	TotalFailed    int64
	TotalRetried   int64

	// Timing metrics
	EnqueueTimes  []time.Duration
	DequeueTimes  []time.Duration
	ProcessTimes  []time.Duration

	// Job type metrics
	ByType map[JobType]*JobTypeMetrics

	// Priority metrics
	ByPriority map[Priority]*PriorityMetrics

	// Time-based metrics
	LastHourCompleted int64
	LastHourFailed    int64

	// Latency metrics
	AvgProcessTime time.Duration
	MinProcessTime time.Duration
	MaxProcessTime time.Duration
	P50ProcessTime time.Duration
	P95ProcessTime time.Duration
	P99ProcessTime time.Duration
}

// JobTypeMetrics tracks metrics for a specific job type
type JobTypeMetrics struct {
	Enqueued      int64
	Completed     int64
	Failed        int64
	AvgProcessTime time.Duration
}

// PriorityMetrics tracks metrics for a specific priority level
type PriorityMetrics struct {
	Enqueued      int64
	Completed     int64
	Failed        int64
	AvgWaitTime   time.Duration
}

// NewQueueMetrics creates a new QueueMetrics instance
func NewQueueMetrics() *QueueMetrics {
	return &QueueMetrics{
		EnqueueTimes:  make([]time.Duration, 0, 1000),
		DequeueTimes:  make([]time.Duration, 0, 1000),
		ProcessTimes:  make([]time.Duration, 0, 1000),
		ByType:        make(map[JobType]*JobTypeMetrics),
		ByPriority:    make(map[Priority]*PriorityMetrics),
		MinProcessTime: time.Hour, // Start with a high value
	}
}

// RecordEnqueue records a job enqueue event
func (qm *QueueMetrics) RecordEnqueue(job *Job) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.TotalEnqueued++

	// Track by type
	if _, ok := qm.ByType[job.Type]; !ok {
		qm.ByType[job.Type] = &JobTypeMetrics{}
	}
	qm.ByType[job.Type].Enqueued++

	// Track by priority
	if _, ok := qm.ByPriority[job.Priority]; !ok {
		qm.ByPriority[job.Priority] = &PriorityMetrics{}
	}
	qm.ByPriority[job.Priority].Enqueued++
}

// RecordDequeue records a job dequeue event
func (qm *QueueMetrics) RecordDequeue(job *Job) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.TotalDequeued++

	// Calculate wait time (time from creation to dequeue)
	if job.CreatedAt.IsZero() == false {
		waitTime := time.Since(job.CreatedAt)
		if metrics, ok := qm.ByPriority[job.Priority]; ok {
			// Simple moving average
			if metrics.AvgWaitTime == 0 {
				metrics.AvgWaitTime = waitTime
			} else {
				metrics.AvgWaitTime = (metrics.AvgWaitTime + waitTime) / 2
			}
		}
	}
}

// RecordCompletion records a job completion event
func (qm *QueueMetrics) RecordCompletion(job *Job) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.TotalCompleted++
	qm.LastHourCompleted++

	// Track by type
	if metrics, ok := qm.ByType[job.Type]; ok {
		metrics.Completed++
	}

	// Track by priority
	if metrics, ok := qm.ByPriority[job.Priority]; ok {
		metrics.Completed++
	}

	// Calculate process time
	if job.StartedAt != nil && job.CompletedAt != nil {
		processTime := job.CompletedAt.Sub(*job.StartedAt)
		qm.recordProcessTime(processTime, job.Type)
	}
}

// RecordFailure records a job failure event
func (qm *QueueMetrics) RecordFailure(job *Job) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.TotalFailed++
	qm.LastHourFailed++

	// Track by type
	if metrics, ok := qm.ByType[job.Type]; ok {
		metrics.Failed++
	}

	// Track by priority
	if metrics, ok := qm.ByPriority[job.Priority]; ok {
		metrics.Failed++
	}
}

// RecordRetry records a job retry event
func (qm *QueueMetrics) RecordRetry(job *Job) {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.TotalRetried++
}

// recordProcessTime records processing time for a job
func (qm *QueueMetrics) recordProcessTime(duration time.Duration, jobType JobType) {
	// Add to list (keep last 1000)
	qm.ProcessTimes = append(qm.ProcessTimes, duration)
	if len(qm.ProcessTimes) > 1000 {
		qm.ProcessTimes = qm.ProcessTimes[1:]
	}

	// Update min/max
	if duration < qm.MinProcessTime {
		qm.MinProcessTime = duration
	}
	if duration > qm.MaxProcessTime {
		qm.MaxProcessTime = duration
	}

	// Calculate average for this job type
	if metrics, ok := qm.ByType[jobType]; ok {
		if metrics.AvgProcessTime == 0 {
			metrics.AvgProcessTime = duration
		} else {
			// Simple moving average
			metrics.AvgProcessTime = (metrics.AvgProcessTime + duration) / 2
		}
	}

	// Calculate overall average
	qm.calculatePercentiles()
}

// calculatePercentiles calculates percentile metrics
func (qm *QueueMetrics) calculatePercentiles() {
	if len(qm.ProcessTimes) == 0 {
		return
	}

	// Sort times for percentile calculation
	times := make([]time.Duration, len(qm.ProcessTimes))
	copy(times, qm.ProcessTimes)

	// Simple bubble sort (ok for small arrays)
	n := len(times)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if times[j] > times[j+1] {
				times[j], times[j+1] = times[j+1], times[j]
			}
		}
	}

	// Calculate percentiles
	qm.P50ProcessTime = times[len(times)*50/100]
	qm.P95ProcessTime = times[len(times)*95/100]
	qm.P99ProcessTime = times[len(times)*99/100]

	// Calculate average
	var sum time.Duration
	for _, t := range times {
		sum += t
	}
	qm.AvgProcessTime = sum / time.Duration(len(times))
}

// GetSummary returns a summary of queue metrics
func (qm *QueueMetrics) GetSummary() MetricsSummary {
	qm.mu.RLock()
	defer qm.mu.RUnlock()

	successRate := float64(0)
	if qm.TotalCompleted+qm.TotalFailed > 0 {
		successRate = float64(qm.TotalCompleted) / float64(qm.TotalCompleted+qm.TotalFailed) * 100
	}

	return MetricsSummary{
		TotalEnqueued:   qm.TotalEnqueued,
		TotalDequeued:   qm.TotalDequeued,
		TotalCompleted:  qm.TotalCompleted,
		TotalFailed:     qm.TotalFailed,
		TotalRetried:    qm.TotalRetried,
		SuccessRate:     successRate,
		AvgProcessTime:  qm.AvgProcessTime,
		MinProcessTime:  qm.MinProcessTime,
		MaxProcessTime:  qm.MaxProcessTime,
		P50ProcessTime:  qm.P50ProcessTime,
		P95ProcessTime:  qm.P95ProcessTime,
		P99ProcessTime:  qm.P99ProcessTime,
		LastHourCompleted: qm.LastHourCompleted,
		LastHourFailed:    qm.LastHourFailed,
	}
}

// MetricsSummary represents a summary of queue metrics
type MetricsSummary struct {
	TotalEnqueued     int64         `json:"total_enqueued"`
	TotalDequeued     int64         `json:"total_dequeued"`
	TotalCompleted    int64         `json:"total_completed"`
	TotalFailed       int64         `json:"total_failed"`
	TotalRetried      int64         `json:"total_retried"`
	SuccessRate       float64       `json:"success_rate"`
	AvgProcessTime    time.Duration `json:"avg_process_time"`
	MinProcessTime    time.Duration `json:"min_process_time"`
	MaxProcessTime    time.Duration `json:"max_process_time"`
	P50ProcessTime    time.Duration `json:"p50_process_time"`
	P95ProcessTime    time.Duration `json:"p95_process_time"`
	P99ProcessTime    time.Duration `json:"p99_process_time"`
	LastHourCompleted int64         `json:"last_hour_completed"`
	LastHourFailed    int64         `json:"last_hour_failed"`
}

// Reset resets all metrics
func (qm *QueueMetrics) Reset() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.TotalEnqueued = 0
	qm.TotalDequeued = 0
	qm.TotalCompleted = 0
	qm.TotalFailed = 0
	qm.TotalRetried = 0
	qm.EnqueueTimes = make([]time.Duration, 0, 1000)
	qm.DequeueTimes = make([]time.Duration, 0, 1000)
	qm.ProcessTimes = make([]time.Duration, 0, 1000)
	qm.ByType = make(map[JobType]*JobTypeMetrics)
	qm.ByPriority = make(map[Priority]*PriorityMetrics)
	qm.LastHourCompleted = 0
	qm.LastHourFailed = 0
	qm.MinProcessTime = time.Hour
	qm.MaxProcessTime = 0
}

// ResetHourlyCounters resets hourly counters (should be called every hour)
func (qm *QueueMetrics) ResetHourlyCounters() {
	qm.mu.Lock()
	defer qm.mu.Unlock()

	qm.LastHourCompleted = 0
	qm.LastHourFailed = 0
}
