package session

import (
	"path/filepath"
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

	if _, err := db.Exec(runtime.SchemaV1); err != nil {
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
