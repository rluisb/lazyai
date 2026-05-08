package queue

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

// mockHandler records invocations for testing.
type mockHandler struct {
	calls     atomic.Int64
	failNext  atomic.Bool
	failError error
}

func (h *mockHandler) Handle(ctx context.Context, job *Job) error {
	h.calls.Add(1)
	if h.failNext.Load() {
		if h.failError != nil {
			return h.failError
		}
		return errors.New("mock handler error")
	}
	return nil
}

func TestWorkerStartStop(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)
	w := &Worker{
		DB:               db,
		Queue:            q,
		PollInterval:     5 * time.Millisecond,
		ReclaimInterval:  10 * time.Millisecond,
		ReclaimTimeoutMs: 100,
	}

	// Use a fresh context for the worker; the test's deferred cancel
	// signals the worker to stop via the context.
	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	// Give the worker a moment to start its goroutines
	time.Sleep(20 * time.Millisecond)

	// Signal stop and wait
	cancel()
	done := make(chan struct{})
	go func() {
		w.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Clean stop confirmed
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not stop within 5 seconds")
	}
}

func TestWorkerPicksUpJobAndCallsHandler(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)
	h := &mockHandler{}

	w := &Worker{
		DB:               db,
		Queue:            q,
		PollInterval:     10 * time.Millisecond,
		ReclaimInterval:   500 * time.Millisecond,
		ReclaimTimeoutMs: 5000,
	}
	w.RegisterHandler("test-job", h)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Enqueue a job before starting
	_, err = q.Enqueue(EnqueueInput{JobType: "test-job", MaxAttempts: 3})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	w.Start(ctx)

	// Poll until the handler is called
	for i := 0; i < 100; i++ {
		if h.calls.Load() > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if h.calls.Load() == 0 {
		t.Fatal("handler was not called within timeout")
	}

	// Verify job is completed
	jobs, err := q.List("completed", 10)
	if err != nil {
		t.Fatalf("list completed: %v", err)
	}
	if len(jobs) != 1 {
		t.Errorf("expected 1 completed job, got %d", len(jobs))
	}
}

func TestWorkerCallsFailOnHandlerError(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)
	h := &mockHandler{failError: errors.New("handler failed")}
	h.failNext.Store(true)

	w := &Worker{
		DB:               db,
		Queue:            q,
		PollInterval:     10 * time.Millisecond,
		ReclaimInterval:   500 * time.Millisecond,
		ReclaimTimeoutMs: 5000,
	}
	w.RegisterHandler("fail-job", h)

	ctx, cancel := context.WithCancel(context.Background())

	_, err = q.Enqueue(EnqueueInput{JobType: "fail-job", MaxAttempts: 3})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	w.Start(ctx)

	// Poll until job is processed (completed or failed)
	var processed bool
	for i := 0; i < 100; i++ {
		completed, _ := q.List("completed", 10)
		failed, _ := q.List("failed", 10)
		if len(completed)+len(failed) > 0 {
			processed = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !processed {
		t.Fatal("job was not processed within timeout")
	}

	jobs, _ := q.List("failed", 10)
	if len(jobs) != 1 {
		t.Errorf("expected 1 failed job, got %d", len(jobs))
	}

	cancel()
}

func TestWorkerCallsCompleteOnSuccess(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)
	h := &mockHandler{} // succeeds

	w := &Worker{
		DB:               db,
		Queue:            q,
		PollInterval:     10 * time.Millisecond,
		ReclaimInterval:   500 * time.Millisecond,
		ReclaimTimeoutMs: 5000,
	}
	w.RegisterHandler("success-job", h)

	ctx, cancel := context.WithCancel(context.Background())

	_, err = q.Enqueue(EnqueueInput{JobType: "success-job", MaxAttempts: 1})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	w.Start(ctx)

	// Poll until job is completed
	var completed bool
	for i := 0; i < 100; i++ {
		jobs, _ := q.List("completed", 10)
		if len(jobs) > 0 {
			completed = true
			break
		}
		time.Sleep(10 * time.Millisecond)
	}

	if !completed {
		t.Fatal("job was not completed within timeout")
	}

	cancel()
}

func TestReclaimLoopRuns(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)
	w := &Worker{
		DB:               db,
		Queue:            q,
		PollInterval:     100 * time.Millisecond,
		ReclaimInterval:  15 * time.Millisecond, // runs frequently
		ReclaimTimeoutMs: 10,
	}

	ctx, cancel := context.WithCancel(context.Background())
	w.Start(ctx)

	// Let the reclaim loop run a few times
	time.Sleep(60 * time.Millisecond)

	cancel()
	done := make(chan struct{})
	go func() {
		w.Stop()
		close(done)
	}()

	select {
	case <-done:
		// Clean stop confirmed
	case <-time.After(5 * time.Second):
		t.Fatal("worker did not stop within 5 seconds")
	}
}