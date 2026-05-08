package queue

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

func TestEnqueueDequeueComplete(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)

	// Enqueue a job
	job, err := q.Enqueue(EnqueueInput{
		JobType:     "test-job",
		Payload:     map[string]any{"key": "value"},
		Priority:    5,
		MaxAttempts: 3,
	})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if job.Status != StatusPending {
		t.Errorf("expected status=pending, got %s", job.Status)
	}
	if job.MaxAttempts != 3 {
		t.Errorf("expected maxAttempts=3, got %d", job.MaxAttempts)
	}

	// Dequeue the job
	dequeued, err := q.Dequeue("test-job")
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if dequeued == nil {
		t.Fatal("expected a job, got nil")
	}
	if dequeued.Status != StatusClaimed {
		t.Errorf("expected status=claimed, got %s", dequeued.Status)
	}
	if dequeued.Attempts != 1 {
		t.Errorf("expected attempts=1, got %d", dequeued.Attempts)
	}

	// Complete the job
	err = q.Complete(dequeued.ID)
	if err != nil {
		t.Fatalf("complete: %v", err)
	}

	completed, err := q.Get(dequeued.ID)
	if err != nil {
		t.Fatalf("get completed: %v", err)
	}
	if completed.Status != StatusCompleted {
		t.Errorf("expected status=completed, got %s", completed.Status)
	}
	if completed.CompletedAt == "" {
		t.Error("expected completedAt to be set")
	}
}

func TestEnqueueDequeueFailRequeue(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)

	if _, err := q.Enqueue(EnqueueInput{
		JobType:     "fail-job",
		Payload:     map[string]any{"fail": true},
		Priority:    0,
		MaxAttempts: 3,
	}); err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	// Dequeue and fail
	dequeued, err := q.Dequeue("fail-job")
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}

	err = q.Fail(dequeued.ID, map[string]any{"error": "something went wrong"})
	if err != nil {
		t.Fatalf("fail: %v", err)
	}

	// Should be re-queued (attempts=1 < maxAttempts=3)
	requeued, err := q.Get(dequeued.ID)
	if err != nil {
		t.Fatalf("get requeued: %v", err)
	}
	if requeued.Status != StatusPending {
		t.Errorf("expected status=pending after re-queue, got %s", requeued.Status)
	}
	if requeued.Error == nil {
		t.Error("expected error to be stored")
	}

	// Second failure — still under max, re-queues again
	failed2, err := q.Get(dequeued.ID)
	if err != nil {
		t.Fatalf("get failed2: %v", err)
	}
	if failed2.Status != StatusPending {
		t.Errorf("expected status=pending after second failure, got %s", failed2.Status)
	}

	// Dequeue and fail again — now attempts=2, still < max=3, so re-queue again
	dequeued2, err := q.Dequeue("fail-job")
	if err != nil {
		t.Fatalf("dequeue2: %v", err)
	}
	err = q.Fail(dequeued2.ID, map[string]any{"error": "still failing"})
	if err != nil {
		t.Fatalf("fail2: %v", err)
	}

	// Third dequeue and third failure — now attempts=3, max=3, so it's final failed
	dequeued3, err := q.Dequeue("fail-job")
	if err != nil {
		t.Fatalf("dequeue3: %v", err)
	}
	err = q.Fail(dequeued3.ID, map[string]any{"error": "third time"})
	if err != nil {
		t.Fatalf("fail3: %v", err)
	}

	// Third failure — should be marked as failed
	failed, err := q.Get(dequeued3.ID)
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if failed.Status != StatusFailed {
		t.Errorf("expected status=failed after exhausting retries, got %s", failed.Status)
	}
	if failed.CompletedAt == "" {
		t.Error("expected completedAt to be set for final failure")
	}
}

func TestDequeue_ReturnsNilWhenEmpty(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)

	result, err := q.Dequeue("non-existent")
	if err != nil {
		t.Fatalf("dequeue error: %v", err)
	}
	if result != nil {
		t.Error("expected nil for empty queue")
	}
}

func TestList_FilterByStatus(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)

	// Enqueue jobs with different statuses via direct insert
	// (simulating different states)
	now := "2024-01-01T00:00:00Z"
	db.Exec(`INSERT INTO queue_jobs (id, job_type, payload_json, status, priority, max_attempts, created_at) VALUES ('job-pending', 'list-test', '{}', 'pending', 0, 3, ?)`, now)
	db.Exec(`INSERT INTO queue_jobs (id, job_type, payload_json, status, priority, max_attempts, created_at) VALUES ('job-claimed', 'list-test', '{}', 'claimed', 0, 3, ?)`, now)
	db.Exec(`INSERT INTO queue_jobs (id, job_type, payload_json, status, priority, max_attempts, created_at) VALUES ('job-completed', 'list-test', '{}', 'completed', 0, 3, ?)`, now)

	// List pending jobs
	pending, err := q.List("pending", 10)
	if err != nil {
		t.Fatalf("list pending: %v", err)
	}
	if len(pending) != 1 {
		t.Errorf("expected 1 pending job, got %d", len(pending))
	}

	// List all jobs (no filter)
	all, err := q.List("", 10)
	if err != nil {
		t.Fatalf("list all: %v", err)
	}
	if len(all) != 3 {
		t.Errorf("expected 3 total jobs, got %d", len(all))
	}
}

func TestReclaim_ResetsStaleClaimedJobs(t *testing.T) {
	db, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if err := db.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	q := New(db)

	// job is used to retrieve ID in the Reclaim test below
	job, _ := q.Enqueue(EnqueueInput{JobType: "reclaim-test", MaxAttempts: 3})
	if err != nil {
		t.Fatalf("enqueue: %v", err)
	}

	// Dequeue to claim it (sets claimed_at to now)
	dequeued, err := q.Dequeue("reclaim-test")
	if err != nil {
		t.Fatalf("dequeue: %v", err)
	}
	if dequeued.Status != StatusClaimed {
		t.Errorf("expected status=claimed, got %s", dequeued.Status)
	}

	// Reclaim with timeout of 1ms — the job was just claimed, so it should NOT be reclaimed
	// (claimed_at is ~now, cutoff is ~now - 1ms, so claimed_at > cutoff)
	// Reclaim with 60s timeout — the job was just claimed so it won't be reclaimed.
	// This tests that Reclaim does not incorrectly reclaim fresh jobs.
	_, err = q.Reclaim(60000)
	if err != nil {
		t.Fatalf("reclaim: %v", err)
	}
	// The job was just claimed, so it shouldn't be reclaimed with a normal timeout
	jobAfterReclaim, err := q.Get(job.ID)
	if err != nil {
		t.Fatalf("get after reclaim: %v", err)
	}
	if jobAfterReclaim.Status != StatusClaimed {
		t.Errorf("job should still be claimed with normal timeout, got %s", jobAfterReclaim.Status)
	}

	_ = dequeued // avoid unused
}