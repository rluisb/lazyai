package taskqueue

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

func setupTestDB(t *testing.T) (*runtime.DB, func()) {
	t.Helper()
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}

	dbPath := filepath.Join(tmpDir, ".specify", "session.db")
	db, err := runtime.Open(dbPath)
	if err != nil {
		t.Fatalf("open runtime database: %v", err)
	}
	if _, err := db.Exec(runtime.SchemaV1); err != nil {
		t.Fatalf("apply schema: %v", err)
	}

	// Ensure a session row exists for FK constraints
	_, err = db.Exec(
		"INSERT INTO sessions (id, started_at, agent, status) VALUES (?, ?, ?, ?)",
		"cli", time.Now().UTC().Format(time.RFC3339), "cli", "active",
	)
	if err != nil {
		t.Fatalf("insert session: %v", err)
	}

	cleanup := func() {
		_ = db.Close()
		_ = os.Chdir(origWd)
	}
	return db, cleanup
}

func TestEnqueueAndGetTask(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)
	task, err := mgr.Enqueue("cli", "test", "do something", EnqueueOptions{})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}
	if task == nil {
		t.Fatal("expected task, got nil")
	}
	if task.Task != "do something" {
		t.Errorf("expected task 'do something', got %q", task.Task)
	}
	if task.Status != TaskOpen {
		t.Errorf("expected status open, got %q", task.Status)
	}
	if task.CreatedAt.IsZero() {
		t.Error("expected CreatedAt to be set")
	}

	// GetTask should scan created_at successfully (TEXT -> time.Time)
	got, err := mgr.GetTask(task.ID)
	if err != nil {
		t.Fatalf("get task failed: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("expected ID %d, got %d", task.ID, got.ID)
	}
	if got.CreatedAt.IsZero() {
		t.Error("expected CreatedAt parsed from TEXT")
	}
}

func TestListTasks(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)
	if _, err := mgr.Enqueue("cli", "test", "task one", EnqueueOptions{}); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}
	if _, err := mgr.Enqueue("cli", "test", "task two", EnqueueOptions{}); err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	tasks, err := mgr.List("cli", "")
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}
	if len(tasks) != 2 {
		t.Errorf("expected 2 tasks, got %d", len(tasks))
	}
	for _, task := range tasks {
		if task.CreatedAt.IsZero() {
			t.Errorf("task %d has zero CreatedAt", task.ID)
		}
	}
}

func TestClaimNext(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)
	_, err := mgr.Enqueue("cli", "test", "claimable task", EnqueueOptions{})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	claimed, err := mgr.Claim("cli", "test", "agent-a")
	if err != nil {
		t.Fatalf("claim failed: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected claimed task, got nil")
	}
	if claimed.Status != TaskClaimed {
		t.Errorf("expected status claimed, got %q", claimed.Status)
	}
	if len(claimed.Claims) != 1 {
		t.Errorf("expected 1 claim, got %d", len(claimed.Claims))
	}
	if claimed.Claims[0].Agent != "agent-a" {
		t.Errorf("expected agent-a, got %q", claimed.Claims[0].Agent)
	}
	if claimed.Claims[0].ClaimedAt.IsZero() {
		t.Error("expected ClaimedAt parsed from TEXT")
	}

	// Second claim should return nil (no tasks available)
	second, err := mgr.Claim("cli", "test", "agent-b")
	if err != nil {
		t.Fatalf("second claim failed: %v", err)
	}
	if second != nil {
		t.Error("expected no task for second claim")
	}
}

func TestClaimByID(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)
	task, err := mgr.Enqueue("cli", "test", "specific task", EnqueueOptions{})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}

	claimed, err := mgr.ClaimByID("cli", "test", "agent-a", task.ID)
	if err != nil {
		t.Fatalf("claim by id failed: %v", err)
	}
	if claimed == nil {
		t.Fatal("expected claimed task, got nil")
	}
	if claimed.ID != task.ID {
		t.Errorf("expected ID %d, got %d", task.ID, claimed.ID)
	}

	// Claim same task again should return nil
	again, err := mgr.ClaimByID("cli", "test", "agent-a", task.ID)
	if err != nil {
		t.Fatalf("second claim by id failed: %v", err)
	}
	if again != nil {
		t.Error("expected nil for duplicate claim")
	}
}

func TestCompleteTask(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)
	task, err := mgr.Enqueue("cli", "test", "completable task", EnqueueOptions{})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}
	if _, err := mgr.Claim("cli", "test", "agent-a"); err != nil {
		t.Fatalf("claim failed: %v", err)
	}

	if err := mgr.Complete(task.ID, "agent-a"); err != nil {
		t.Fatalf("complete failed: %v", err)
	}

	completed, err := mgr.GetTask(task.ID)
	if err != nil {
		t.Fatalf("get task failed: %v", err)
	}
	if completed.Status != TaskCompleted {
		t.Errorf("expected status completed, got %q", completed.Status)
	}
}

func TestFailTask(t *testing.T) {
	db, cleanup := setupTestDB(t)
	defer cleanup()

	mgr := NewManager(db)
	task, err := mgr.Enqueue("cli", "test", "failable task", EnqueueOptions{})
	if err != nil {
		t.Fatalf("enqueue failed: %v", err)
	}
	if _, err := mgr.Claim("cli", "test", "agent-a"); err != nil {
		t.Fatalf("claim failed: %v", err)
	}

	if err := mgr.Fail(task.ID, "agent-a", "something broke", "context"); err != nil {
		t.Fatalf("fail failed: %v", err)
	}

	failed, err := mgr.GetTask(task.ID)
	if err != nil {
		t.Fatalf("get task failed: %v", err)
	}
	if failed.Status != TaskFailed {
		t.Errorf("expected status failed, got %q", failed.Status)
	}

	dlq, err := mgr.GetDLQ("cli")
	if err != nil {
		t.Fatalf("get dlq failed: %v", err)
	}
	if len(dlq) != 1 {
		t.Errorf("expected 1 dlq entry, got %d", len(dlq))
	}
	if len(dlq) > 0 && dlq[0].FailedAt.IsZero() {
		t.Error("expected DLQ FailedAt parsed from TEXT")
	}
}
