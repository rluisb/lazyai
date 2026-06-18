package sqlite

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

func TestJobQueueStoreEnqueueClaimCompleteAndList(t *testing.T) {
	store := newJobQueueStoreTestAdapter(t)

	job, err := store.EnqueueJob(domain.EnqueueJobInput{
		JobType:     "test-job",
		Payload:     map[string]any{"key": "value"},
		Priority:    5,
		MaxAttempts: 0,
	})
	if err != nil {
		t.Fatalf("enqueue job: %v", err)
	}
	if job.Status != domain.QueueJobStatusPending {
		t.Fatalf("expected pending job, got %s", job.Status)
	}
	if job.MaxAttempts != 3 {
		t.Fatalf("expected default max attempts 3, got %d", job.MaxAttempts)
	}

	claimed, err := store.DequeueJob("test-job")
	if err != nil {
		t.Fatalf("dequeue job: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected claimed job, got nil")
	}
	if claimed.Status != domain.QueueJobStatusClaimed || claimed.Attempts != 1 || claimed.ClaimedAt == "" {
		t.Fatalf("unexpected claimed job: %+v", claimed)
	}

	if err := store.CompleteJob(claimed.ID); err != nil {
		t.Fatalf("complete job: %v", err)
	}
	completed, err := store.GetJob(claimed.ID)
	if err != nil {
		t.Fatalf("get completed job: %v", err)
	}
	if completed.Status != domain.QueueJobStatusCompleted || completed.CompletedAt == "" {
		t.Fatalf("unexpected completed job: %+v", completed)
	}

	jobs, err := store.ListJobs(string(domain.QueueJobStatusCompleted), 10)
	if err != nil {
		t.Fatalf("list completed jobs: %v", err)
	}
	if len(jobs) != 1 || jobs[0].ID != completed.ID {
		t.Fatalf("expected one completed job, got %+v", jobs)
	}
}

func TestJobQueueStoreFailRequeuesUntilAttemptsExhausted(t *testing.T) {
	store := newJobQueueStoreTestAdapter(t)

	job, err := store.EnqueueJob(domain.EnqueueJobInput{JobType: "fail-job", MaxAttempts: 2})
	if err != nil {
		t.Fatalf("enqueue job: %v", err)
	}

	claimed, err := store.DequeueJob("fail-job")
	if err != nil {
		t.Fatalf("first dequeue: %v", err)
	}
	if err := store.FailJob(claimed.ID, map[string]any{"error": "first"}); err != nil {
		t.Fatalf("first fail: %v", err)
	}
	requeued, err := store.GetJob(job.ID)
	if err != nil {
		t.Fatalf("get requeued job: %v", err)
	}
	if requeued.Status != domain.QueueJobStatusPending || requeued.ClaimedAt != "" {
		t.Fatalf("expected pending requeued job, got %+v", requeued)
	}

	claimed, err = store.DequeueJob("fail-job")
	if err != nil {
		t.Fatalf("second dequeue: %v", err)
	}
	if err := store.FailJob(claimed.ID, map[string]any{"error": "second"}); err != nil {
		t.Fatalf("second fail: %v", err)
	}
	failed, err := store.GetJob(job.ID)
	if err != nil {
		t.Fatalf("get failed job: %v", err)
	}
	if failed.Status != domain.QueueJobStatusFailed || failed.CompletedAt == "" {
		t.Fatalf("expected terminal failed job, got %+v", failed)
	}
	if failed.Error["error"] != "second" {
		t.Fatalf("expected stored error, got %+v", failed.Error)
	}
}

func TestJobQueueStoreDequeueReturnsNilWhenEmpty(t *testing.T) {
	store := newJobQueueStoreTestAdapter(t)

	job, err := store.DequeueJob("missing")
	if err != nil {
		t.Fatalf("dequeue empty queue: %v", err)
	}
	if job != nil {
		t.Fatalf("expected nil job, got %+v", job)
	}
}

func newJobQueueStoreTestAdapter(t *testing.T) *JobQueueStore {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewJobQueueStore(database)
}
