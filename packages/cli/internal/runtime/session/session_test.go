package session

import (
	"fmt"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

func setupSessionTestDB(t *testing.T) *runtime.DB {
	t.Helper()

	dbPath := filepath.Join(t.TempDir(), "session.db")
	db, err := runtime.Open(dbPath)
	if err != nil {
		t.Fatalf("runtime.Open failed: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec(runtime.SchemaCurrent); err != nil {
		t.Fatalf("apply schema failed: %v", err)
	}

	return db
}

func TestSessionManagerStartParsesStartedAt(t *testing.T) {
	db := setupSessionTestDB(t)
	mgr := NewManager(db)

	s, err := mgr.Start("timestamp test", StartOptions{Agent: "test-agent"})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	if s.StartedAt.IsZero() {
		t.Fatal("expected StartedAt to be parsed")
	}
	if s.Goal != "timestamp test" {
		t.Errorf("Goal = %q, want timestamp test", s.Goal)
	}
	if s.Agent != "test-agent" {
		t.Errorf("Agent = %q, want test-agent", s.Agent)
	}
}

func TestSessionManagerGetParsesStartedAndEndedAt(t *testing.T) {
	db := setupSessionTestDB(t)
	mgr := NewManager(db)

	startedAt := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	endedAt := startedAt.Add(5 * time.Minute)
	_, err := db.Exec(
		"INSERT INTO sessions (id, started_at, ended_at, agent, status, goal) VALUES (?, ?, ?, ?, ?, ?)",
		"ses_test_get", startedAt.Format(time.RFC3339), endedAt.Format(time.RFC3339), "cli", "ended", "get test",
	)
	if err != nil {
		t.Fatalf("insert session failed: %v", err)
	}

	s, err := mgr.Get("ses_test_get")
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}

	if !s.StartedAt.Equal(startedAt) {
		t.Errorf("StartedAt = %s, want %s", s.StartedAt, startedAt)
	}
	if s.EndedAt == nil || !s.EndedAt.Equal(endedAt) {
		t.Fatalf("EndedAt = %v, want %s", s.EndedAt, endedAt)
	}
}

func TestSessionManagerListParsesStartedAt(t *testing.T) {
	db := setupSessionTestDB(t)
	mgr := NewManager(db)

	startedAt := time.Date(2026, 5, 28, 12, 0, 0, 0, time.UTC)
	for _, sessionID := range []string{"ses_test_list_1", "ses_test_list_2"} {
		_, err := db.Exec(
			"INSERT INTO sessions (id, started_at, agent, status, goal) VALUES (?, ?, ?, ?, ?)",
			sessionID, startedAt.Format(time.RFC3339), "cli", "active", sessionID,
		)
		if err != nil {
			t.Fatalf("insert session %s failed: %v", sessionID, err)
		}
	}

	sessions, err := mgr.List()
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}
	if len(sessions) != 2 {
		t.Fatalf("len(sessions) = %d, want 2", len(sessions))
	}
	for _, s := range sessions {
		if s.StartedAt.IsZero() {
			t.Fatalf("session %s has zero StartedAt", s.ID)
		}
	}
}

func TestAcquireLock_ConcurrentOnlyOneSucceeds(t *testing.T) {
	db := setupSessionTestDB(t)
	mgr := NewManager(db)

	// Seed a session so the lock has a valid session_id.
	s, err := mgr.Start("lock race test", StartOptions{Agent: "test-agent"})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	const goroutines = 20
	var successCount, errorCount int
	var wg sync.WaitGroup
	var mu sync.Mutex
	var firstErr error

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			_, err := mgr.AcquireLock(s.ID, "resource-A", fmt.Sprintf("worker-%d", n))
			mu.Lock()
			if err != nil {
				errorCount++
				if firstErr == nil {
					firstErr = err
				}
			} else {
				successCount++
			}
			mu.Unlock()
		}(i)
	}
	wg.Wait()

	if successCount != 1 {
		t.Fatalf("expected exactly 1 successful lock acquisition, got %d (errors: %d, firstErr: %v)",
			successCount, errorCount, firstErr)
	}
	if errorCount != goroutines-1 {
		t.Fatalf("expected %d errors, got %d", goroutines-1, errorCount)
	}
}

func TestAddTags_ConcurrentNoLostUpdates(t *testing.T) {
	db := setupSessionTestDB(t)
	mgr := NewManager(db)

	s, err := mgr.Start("tag race test", StartOptions{Agent: "test-agent"})
	if err != nil {
		t.Fatalf("Start failed: %v", err)
	}

	// Each goroutine adds a distinct tag concurrently.
	const goroutines = 20
	var wg sync.WaitGroup
	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(n int) {
			defer wg.Done()
			tag := fmt.Sprintf("tag-%d", n)
			if err := mgr.AddTags(s.ID, []string{tag}); err != nil {
				t.Errorf("AddTags %d failed: %v", n, err)
			}
		}(i)
	}
	wg.Wait()

	// After the dust settles, all 20 tags must be present — no lost updates.
	got, err := mgr.Get(s.ID)
	if err != nil {
		t.Fatalf("Get failed: %v", err)
	}
	if len(got.Tags) != goroutines {
		t.Fatalf("expected %d tags, got %d: %v", goroutines, len(got.Tags), got.Tags)
	}

	seen := make(map[string]bool)
	for _, tag := range got.Tags {
		seen[tag] = true
	}
	for i := 0; i < goroutines; i++ {
		tag := fmt.Sprintf("tag-%d", i)
		if !seen[tag] {
			t.Errorf("missing tag %s", tag)
		}
	}
}
