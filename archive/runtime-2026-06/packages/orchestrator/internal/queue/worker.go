package queue

import (
	"context"
	"sync"
	"time"

	charmlog "charm.land/log/v2"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	orchlog "github.com/rluisb/lazyai/packages/orchestrator/internal/log"
)

// JobHandler processes a queued job.
type JobHandler interface {
	Handle(ctx context.Context, job *Job) error
}

// Worker consumes jobs from a queue and dispatches them to registered handlers.
type Worker struct {
	DB              *db.DB
	Queue           *Queue
	Handlers        map[string]JobHandler
	PollInterval    time.Duration
	ReclaimInterval time.Duration
	ReclaimTimeoutMs int

	stopCh chan struct{}
	stopWg sync.WaitGroup
	mu     sync.RWMutex
}

// RegisterHandler registers a handler for a job type.
// If no handler is registered for a job type, the job is silently dropped.
func (w *Worker) RegisterHandler(jobType string, h JobHandler) {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.Handlers == nil {
		w.Handlers = make(map[string]JobHandler)
	}
	w.Handlers[jobType] = h
}

// Start begins the worker loops.
// Goroutine 1 polls Dequeue and dispatches to handlers.
// Goroutine 2 periodically calls Reclaim.
// Both loops stop when ctx is cancelled.
func (w *Worker) Start(ctx context.Context) {
	w.mu.Lock()
	if w.stopCh != nil {
		w.mu.Unlock()
		return // already started
	}
	w.stopCh = make(chan struct{})
	w.mu.Unlock()

	pollInterval := w.PollInterval
	if pollInterval <= 0 {
		pollInterval = 2 * time.Second
	}
	reclaimInterval := w.ReclaimInterval
	if reclaimInterval <= 0 {
		reclaimInterval = 30 * time.Second
	}
	reclaimTimeout := w.ReclaimTimeoutMs
	if reclaimTimeout <= 0 {
		reclaimTimeout = 60000
	}

	logger := daemonLogger()

	w.stopWg.Add(2)

	// Goroutine 1: dequeue loop
	go func() {
		defer w.stopWg.Done()
		for {
			select {
			case <-w.stopCh:
				return
			case <-ctx.Done():
				return
			default:
			}

			job, err := w.Queue.Dequeue("")
			if err != nil {
				logger.Error("dequeue error", "error", err)
				time.Sleep(pollInterval)
				continue
			}
			if job == nil {
				time.Sleep(pollInterval)
				continue
			}

			logger.Info("job dequeued", "jobId", job.ID, "jobType", job.JobType)

			// Dispatch to handler in a separate goroutine so we don't block dequeue
			go func(j *Job) {
				handler := w.handlerFor(j.JobType)
				if handler == nil {
					logger.Warn("no handler registered for job type", "jobType", j.JobType, "jobId", j.ID)
					_ = w.Queue.Complete(j.ID)
					logger.Info("job completed (no handler)", "jobId", j.ID)
					return
				}

				if err := handler.Handle(ctx, j); err != nil {
					logger.Error("handler error", "jobId", j.ID, "jobType", j.JobType, "error", err)
					_ = w.Queue.Fail(j.ID, map[string]any{"error": err.Error()})
					logger.Info("job failed", "jobId", j.ID)
				} else {
					_ = w.Queue.Complete(j.ID)
					logger.Info("job completed", "jobId", j.ID)
				}
			}(job)

			// Small yield to allow other goroutines to run
			time.Sleep(10 * time.Millisecond)
		}
	}()

	// Goroutine 2: reclaim loop
	go func() {
		defer w.stopWg.Done()
		for {
			select {
			case <-w.stopCh:
				return
			case <-ctx.Done():
				return
			case <-time.After(reclaimInterval):
			}

			n, err := w.Queue.Reclaim(reclaimTimeout)
			if err != nil {
				logger.Error("reclaim error", "error", err)
				continue
			}
			if n > 0 {
				logger.Info("reclaimed stale jobs", "count", n)
			}
		}
	}()

	logger.Info("worker started",
		"pollInterval", pollInterval,
		"reclaimInterval", reclaimInterval,
		"reclaimTimeoutMs", reclaimTimeout)
}

// Stop signals the worker to stop and waits for goroutines to finish.
func (w *Worker) Stop() {
	w.mu.Lock()
	if w.stopCh == nil {
		w.mu.Unlock()
		return
	}
	close(w.stopCh)
	w.stopCh = nil
	w.mu.Unlock()

	w.stopWg.Wait()

	logger := daemonLogger()
	logger.Info("worker stopped")
}

// handlerFor returns the registered handler for a job type, or nil if none.
func (w *Worker) handlerFor(jobType string) JobHandler {
	w.mu.RLock()
	defer w.mu.RUnlock()
	if h, ok := w.Handlers[jobType]; ok {
		return h
	}
	return nil
}

// daemonLogger returns the orchestrator daemon logger.
func daemonLogger() *charmlog.Logger {
	return orchlog.Default().With("component", "worker")
}
