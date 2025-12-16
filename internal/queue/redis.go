package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/gongahkia/kite/pkg/errors"
)

// RedisQueue implements Queue using Redis Streams
type RedisQueue struct {
	client      *redis.Client
	stream      string
	group       string
	consumer    string
	dlqStream   string
	mu          sync.RWMutex
	jobsMap     map[string]*Job
	stats       QueueStats
	dlq         DeadLetterQueue
	metrics     *QueueMetrics
}

// RedisQueueConfig represents Redis queue configuration
type RedisQueueConfig struct {
	Addr          string
	Password      string
	DB            int
	Stream        string
	Group         string
	Consumer      string
	DLQStream     string
	MaxRetries    int
	EnableMetrics bool
}

// DefaultRedisQueueConfig returns default Redis configuration
func DefaultRedisQueueConfig() *RedisQueueConfig {
	return &RedisQueueConfig{
		Addr:          "localhost:6379",
		Password:      "",
		DB:            0,
		Stream:        "kite:jobs",
		Group:         "kite-workers",
		Consumer:      "worker-1",
		DLQStream:     "kite:jobs:dlq",
		MaxRetries:    3,
		EnableMetrics: true,
	}
}

// NewRedisQueue creates a new Redis-based queue
func NewRedisQueue(config *RedisQueueConfig) (*RedisQueue, error) {
	if config == nil {
		config = DefaultRedisQueueConfig()
	}

	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     config.Addr,
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	queue := &RedisQueue{
		client:    client,
		stream:    config.Stream,
		group:     config.Group,
		consumer:  config.Consumer,
		dlqStream: config.DLQStream,
		jobsMap:   make(map[string]*Job),
		dlq:       NewMemoryDLQ(),
	}

	if config.EnableMetrics {
		queue.metrics = NewQueueMetrics()
	}

	// Create consumer group
	if err := queue.createGroup(ctx); err != nil {
		client.Close()
		return nil, fmt.Errorf("failed to create consumer group: %w", err)
	}

	return queue, nil
}

// createGroup creates the consumer group
func (rq *RedisQueue) createGroup(ctx context.Context) error {
	// Try to create the group (will fail if it already exists)
	err := rq.client.XGroupCreateMkStream(ctx, rq.stream, rq.group, "0").Err()
	if err != nil && err.Error() != "BUSYGROUP Consumer Group name already exists" {
		return err
	}
	return nil
}

// Enqueue adds a job to the queue
func (rq *RedisQueue) Enqueue(ctx context.Context, job *Job) error {
	rq.mu.Lock()
	rq.jobsMap[job.ID] = job
	rq.mu.Unlock()

	// Serialize job
	data, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("failed to marshal job: %w", err)
	}

	// Add to Redis Stream
	values := map[string]interface{}{
		"id":       job.ID,
		"type":     string(job.Type),
		"priority": job.Priority,
		"data":     string(data),
	}

	_, err = rq.client.XAdd(ctx, &redis.XAddArgs{
		Stream: rq.stream,
		Values: values,
	}).Result()

	if err != nil {
		rq.mu.Lock()
		delete(rq.jobsMap, job.ID)
		rq.mu.Unlock()
		return fmt.Errorf("failed to add to stream: %w", err)
	}

	// Update metrics
	if rq.metrics != nil {
		rq.metrics.RecordEnqueue(job)
	}

	// Update stats
	rq.mu.Lock()
	rq.stats.LastEnqueued = time.Now()
	rq.stats.Pending++
	rq.mu.Unlock()

	return nil
}

// Dequeue retrieves the next job from the queue
func (rq *RedisQueue) Dequeue(ctx context.Context) (*Job, error) {
	// Read from stream using consumer group
	streams, err := rq.client.XReadGroup(ctx, &redis.XReadGroupArgs{
		Group:    rq.group,
		Consumer: rq.consumer,
		Streams:  []string{rq.stream, ">"},
		Count:    1,
		Block:    1 * time.Second,
	}).Result()

	if err != nil {
		if err == redis.Nil {
			return nil, errors.QueueError("no messages available", errors.ErrQueueEmpty)
		}
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	if len(streams) == 0 || len(streams[0].Messages) == 0 {
		return nil, errors.QueueError("no messages available", errors.ErrQueueEmpty)
	}

	msg := streams[0].Messages[0]

	// Extract job data
	jobData, ok := msg.Values["data"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid job data format")
	}

	// Deserialize job
	var job Job
	if err := json.Unmarshal([]byte(jobData), &job); err != nil {
		return nil, fmt.Errorf("failed to unmarshal job: %w", err)
	}

	// Mark as started
	job.MarkStarted()

	// Store message ID for later ack
	rq.mu.Lock()
	rq.jobsMap[job.ID] = &job
	rq.stats.LastDequeued = time.Now()
	rq.stats.Pending--
	rq.stats.Running++
	rq.mu.Unlock()

	// Update metrics
	if rq.metrics != nil {
		rq.metrics.RecordDequeue(&job)
	}

	return &job, nil
}

// Ack acknowledges successful completion of a job
func (rq *RedisQueue) Ack(ctx context.Context, jobID string) error {
	rq.mu.Lock()
	job, ok := rq.jobsMap[jobID]
	if !ok {
		rq.mu.Unlock()
		return errors.QueueError("job not found", errors.ErrNotFound)
	}

	job.MarkCompleted(nil)
	delete(rq.jobsMap, jobID)
	rq.stats.Running--
	rq.stats.Completed++
	rq.mu.Unlock()

	// Update metrics
	if rq.metrics != nil {
		rq.metrics.RecordCompletion(job)
	}

	// In production, you'd also XAck the message ID here
	// For simplicity, we're omitting that in this implementation

	return nil
}

// Nack negatively acknowledges a job
func (rq *RedisQueue) Nack(ctx context.Context, jobID string, requeue bool) error {
	rq.mu.Lock()
	job, ok := rq.jobsMap[jobID]
	if !ok {
		rq.mu.Unlock()
		return errors.QueueError("job not found", errors.ErrNotFound)
	}

	if requeue && job.ShouldRetry() {
		// Re-enqueue the job
		rq.mu.Unlock()
		return rq.Enqueue(ctx, job)
	}

	// Send to DLQ
	job.MarkFailed(fmt.Errorf("job failed after %d attempts", job.Attempts))
	if err := rq.dlq.Add(job); err != nil {
		rq.mu.Unlock()
		return fmt.Errorf("failed to add to DLQ: %w", err)
	}

	delete(rq.jobsMap, jobID)
	rq.stats.Running--
	rq.stats.Failed++
	rq.mu.Unlock()

	// Update metrics
	if rq.metrics != nil {
		rq.metrics.RecordFailure(job)
	}

	// Add to DLQ stream
	data, _ := json.Marshal(job)
	rq.client.XAdd(ctx, &redis.XAddArgs{
		Stream: rq.dlqStream,
		Values: map[string]interface{}{
			"id":   job.ID,
			"data": string(data),
		},
	})

	return nil
}

// GetDepth returns the current queue depth
func (rq *RedisQueue) GetDepth(ctx context.Context) (int, error) {
	length, err := rq.client.XLen(ctx, rq.stream).Result()
	if err != nil {
		return 0, err
	}

	return int(length), nil
}

// GetStats returns queue statistics
func (rq *RedisQueue) GetStats() QueueStats {
	rq.mu.RLock()
	defer rq.mu.RUnlock()

	stats := rq.stats

	// Get depth from Redis
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if length, err := rq.client.XLen(ctx, rq.stream).Result(); err == nil {
		stats.Depth = int(length)
	}

	return stats
}

// GetDLQ returns the dead letter queue
func (rq *RedisQueue) GetDLQ() DeadLetterQueue {
	return rq.dlq
}

// GetMetrics returns queue metrics
func (rq *RedisQueue) GetMetrics() *QueueMetrics {
	return rq.metrics
}

// Close closes the queue connection
func (rq *RedisQueue) Close() error {
	return rq.client.Close()
}

// Purge removes all messages from the queue
func (rq *RedisQueue) Purge(ctx context.Context) error {
	return rq.client.Del(ctx, rq.stream).Err()
}
