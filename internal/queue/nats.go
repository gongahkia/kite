package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/gongahkia/kite/pkg/errors"
)

// NATSQueue implements Queue using NATS JetStream
type NATSQueue struct {
	nc          *nats.Conn
	js          nats.JetStreamContext
	stream      string
	subject     string
	dlqSubject  string
	consumer    string
	mu          sync.RWMutex
	jobsMap     map[string]*Job
	stats       QueueStats
	dlq         DeadLetterQueue
	metrics     *QueueMetrics
}

// NATSQueueConfig represents NATS queue configuration
type NATSQueueConfig struct {
	URL            string
	Stream         string
	Subject        string
	Consumer       string
	DLQSubject     string
	MaxRetries     int
	RetryDelay     time.Duration
	AckWait        time.Duration
	MaxDeliver     int
	EnableMetrics  bool
}

// DefaultNATSQueueConfig returns default NATS configuration
func DefaultNATSQueueConfig() *NATSQueueConfig {
	return &NATSQueueConfig{
		URL:           nats.DefaultURL,
		Stream:        "KITE_JOBS",
		Subject:       "jobs.*",
		Consumer:      "kite-workers",
		DLQSubject:    "jobs.dlq",
		MaxRetries:    3,
		RetryDelay:    5 * time.Second,
		AckWait:       30 * time.Second,
		MaxDeliver:    3,
		EnableMetrics: true,
	}
}

// NewNATSQueue creates a new NATS-based queue
func NewNATSQueue(config *NATSQueueConfig) (*NATSQueue, error) {
	if config == nil {
		config = DefaultNATSQueueConfig()
	}

	// Connect to NATS
	nc, err := nats.Connect(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	queue := &NATSQueue{
		nc:         nc,
		js:         js,
		stream:     config.Stream,
		subject:    config.Subject,
		dlqSubject: config.DLQSubject,
		consumer:   config.Consumer,
		jobsMap:    make(map[string]*Job),
		dlq:        NewMemoryDLQ(),
	}

	if config.EnableMetrics {
		queue.metrics = NewQueueMetrics()
	}

	// Create stream if it doesn't exist
	if err := queue.createStream(config); err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create stream: %w", err)
	}

	return queue, nil
}

// createStream creates the JetStream stream
func (nq *NATSQueue) createStream(config *NATSQueueConfig) error {
	// Check if stream exists
	_, err := nq.js.StreamInfo(nq.stream)
	if err == nil {
		// Stream already exists
		return nil
	}

	// Create stream
	_, err = nq.js.AddStream(&nats.StreamConfig{
		Name:      nq.stream,
		Subjects:  []string{nq.subject, nq.dlqSubject},
		Storage:   nats.FileStorage,
		Retention: nats.WorkQueuePolicy,
		MaxAge:    7 * 24 * time.Hour, // 7 days retention
	})

	return err
}

// Enqueue adds a job to the queue
func (nq *NATSQueue) Enqueue(ctx context.Context, job *Job) error {
	nq.mu.Lock()
	nq.jobsMap[job.ID] = job
	nq.mu.Unlock()

	// Serialize job
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Determine subject based on priority
	subject := fmt.Sprintf("jobs.%s", job.Type)

	// Publish to NATS
	_, err = nq.js.Publish(subject, data)
	if err != nil {
		nq.mu.Lock()
		delete(nq.jobsMap, job.ID)
		nq.mu.Unlock()
		return fmt.Errorf("failed to publish job: %w", err)
	}

	// Update metrics
	if nq.metrics != nil {
		nq.metrics.RecordEnqueue(job)
	}

	// Update stats
	nq.mu.Lock()
	nq.stats.LastEnqueued = time.Now()
	nq.stats.Pending++
	nq.mu.Unlock()

	return nil
}

// Dequeue retrieves the next job from the queue
func (nq *NATSQueue) Dequeue(ctx context.Context) (*Job, error) {
	// Subscribe to all job subjects
	sub, err := nq.js.PullSubscribe(nq.subject, nq.consumer)
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe: %w", err)
	}
	defer sub.Unsubscribe()

	// Fetch one message
	msgs, err := sub.Fetch(1, nats.Context(ctx))
	if err != nil {
		if err == context.Canceled || err == context.DeadlineExceeded {
			return nil, err
		}
		return nil, errors.QueueError("failed to fetch message", err)
	}

	if len(msgs) == 0 {
		return nil, errors.QueueError("no messages available", errors.ErrQueueEmpty)
	}

	msg := msgs[0]

	// Deserialize job
	var job Job
	if err := json.Unmarshal(msg.Data, &job); err != nil {
		msg.Nak()
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Mark as started
	job.MarkStarted()

	// Store message metadata for later ack/nack
	nq.mu.Lock()
	nq.jobsMap[job.ID] = &job
	nq.stats.LastDequeued = time.Now()
	nq.stats.Pending--
	nq.stats.Running++
	nq.mu.Unlock()

	// Update metrics
	if nq.metrics != nil {
		nq.metrics.RecordDequeue(&job)
	}

	return &job, nil
}

// Ack acknowledges successful completion of a job
func (nq *NATSQueue) Ack(ctx context.Context, jobID string) error {
	nq.mu.Lock()
	job, ok := nq.jobsMap[jobID]
	if !ok {
		nq.mu.Unlock()
		return errors.QueueError("job not found", errors.ErrNotFound)
	}

	job.MarkCompleted(nil)
	delete(nq.jobsMap, jobID)
	nq.stats.Running--
	nq.stats.Completed++
	nq.mu.Unlock()

	// Update metrics
	if nq.metrics != nil {
		nq.metrics.RecordCompletion(job)
	}

	return nil
}

// Nack negatively acknowledges a job
func (nq *NATSQueue) Nack(ctx context.Context, jobID string, requeue bool) error {
	nq.mu.Lock()
	job, ok := nq.jobsMap[jobID]
	if !ok {
		nq.mu.Unlock()
		return errors.QueueError("job not found", errors.ErrNotFound)
	}

	if requeue && job.ShouldRetry() {
		// Re-enqueue the job
		nq.mu.Unlock()
		return nq.Enqueue(ctx, job)
	}

	// Send to DLQ
	job.MarkFailed(fmt.Errorf("job failed after %d attempts", job.Attempts))
	if err := nq.dlq.Add(job); err != nil {
		nq.mu.Unlock()
		return fmt.Errorf("failed to add to DLQ: %w", err)
	}

	delete(nq.jobsMap, jobID)
	nq.stats.Running--
	nq.stats.Failed++
	nq.mu.Unlock()

	// Update metrics
	if nq.metrics != nil {
		nq.metrics.RecordFailure(job)
	}

	// Publish to DLQ subject
	data, _ := json.Marshal(job)
	nq.js.Publish(nq.dlqSubject, data)

	return nil
}

// GetDepth returns the current queue depth
func (nq *NATSQueue) GetDepth(ctx context.Context) (int, error) {
	info, err := nq.js.StreamInfo(nq.stream)
	if err != nil {
		return 0, err
	}

	return int(info.State.Msgs), nil
}

// GetStats returns queue statistics
func (nq *NATSQueue) GetStats() QueueStats {
	nq.mu.RLock()
	defer nq.mu.RUnlock()

	stats := nq.stats

	// Get depth from NATS
	if info, err := nq.js.StreamInfo(nq.stream); err == nil {
		stats.Depth = int(info.State.Msgs)
	}

	return stats
}

// GetDLQ returns the dead letter queue
func (nq *NATSQueue) GetDLQ() DeadLetterQueue {
	return nq.dlq
}

// GetMetrics returns queue metrics
func (nq *NATSQueue) GetMetrics() *QueueMetrics {
	return nq.metrics
}

// Close closes the queue connection
func (nq *NATSQueue) Close() error {
	nq.nc.Close()
	return nil
}

// Purge removes all messages from the queue
func (nq *NATSQueue) Purge() error {
	return nq.js.PurgeStream(nq.stream)
}
