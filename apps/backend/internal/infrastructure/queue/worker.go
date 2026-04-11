package queue

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/gabriel-q7/portfolio/backend/pkg/logger"
	"github.com/gabriel-q7/portfolio/backend/pkg/metrics"
)

// Job is the interface that queued work must implement.
type Job interface {
	Execute(ctx context.Context) error
	Type() string
}

// QueueStats holds queue health statistics.
type QueueStats struct {
	Pending        int
	DeadLetterCount int
}

// Queue manages an in-memory job queue with a worker pool.
type Queue struct {
	jobs            chan Job
	workers         int
	wg              sync.WaitGroup
	logger          logger.Logger
	metrics         *metrics.Metrics
	deadLetterQueue []Job
	mu              sync.Mutex
}

// New creates a new Queue with the given buffer size and worker count.
func New(bufferSize, workers int, log logger.Logger, m *metrics.Metrics) *Queue {
	return &Queue{
		jobs:    make(chan Job, bufferSize),
		workers: workers,
		logger:  log,
		metrics: m,
	}
}

// Start launches the worker goroutines. It stops when ctx is cancelled.
func (q *Queue) Start(ctx context.Context) {
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go func(id int) {
			defer q.wg.Done()
			q.work(ctx, id)
		}(i)
	}
}

func (q *Queue) work(ctx context.Context, workerID int) {
	for {
		select {
		case <-ctx.Done():
			return
		case job, ok := <-q.jobs:
			if !ok {
				return
			}
			q.process(job, workerID)
		}
	}
}

func (q *Queue) process(job Job, workerID int) {
	const maxRetries = 3
	const retryDelay = 500 * time.Millisecond
	const jobTimeout = 60 * time.Second

	for attempt := 1; attempt <= maxRetries; attempt++ {
		ctx, cancel := context.WithTimeout(context.Background(), jobTimeout)
		err := job.Execute(ctx)
		cancel()

		if err == nil {
			q.logger.Info("Job completed", "type", job.Type(), "worker", workerID, "attempt", attempt)
			if q.metrics != nil {
				q.metrics.RecordQueueJob(job.Type(), "success")
			}
			return
		}

		q.logger.Warn("Job failed, retrying",
			"type", job.Type(), "worker", workerID,
			"attempt", attempt, "error", err,
		)

		if attempt < maxRetries {
			time.Sleep(retryDelay)
		}
	}

	q.logger.Error("Job moved to DLQ after max retries", "type", job.Type(), "worker", workerID)
	if q.metrics != nil {
		q.metrics.RecordQueueJob(job.Type(), "dead_letter")
	}
	q.mu.Lock()
	q.deadLetterQueue = append(q.deadLetterQueue, job)
	q.mu.Unlock()
}

// Enqueue submits a job. Returns an error if the buffer is full.
func (q *Queue) Enqueue(job Job) error {
	select {
	case q.jobs <- job:
		return nil
	default:
		return fmt.Errorf("queue buffer full, job type %q dropped", job.Type())
	}
}

// Drain closes the job channel and waits for all workers to finish.
func (q *Queue) Drain() {
	close(q.jobs)
	q.wg.Wait()
}

// Stats returns current queue statistics.
func (q *Queue) Stats() QueueStats {
	q.mu.Lock()
	dlqLen := len(q.deadLetterQueue)
	q.mu.Unlock()
	return QueueStats{
		Pending:        len(q.jobs),
		DeadLetterCount: dlqLen,
	}
}

// ProcessAIJobStruct is an example AI processing job.
type ProcessAIJobStruct struct {
	ProjectID string
	Prompt    string
}

func (j *ProcessAIJobStruct) Type() string { return "process_ai" }

func (j *ProcessAIJobStruct) Execute(ctx context.Context) error {
	// Actual AI processing would be wired in here.
	return nil
}

// FetchExternalDataJob is an example external API fetch job.
type FetchExternalDataJob struct {
	ResourceURL string
	ResourceID  string
}

func (j *FetchExternalDataJob) Type() string { return "fetch_external_data" }

func (j *FetchExternalDataJob) Execute(ctx context.Context) error {
	// Actual HTTP fetch would be wired in here.
	return nil
}
