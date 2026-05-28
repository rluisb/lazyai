package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// withTempDir changes to a temporary directory and restores the original
// working directory on test cleanup. Returns the temp directory path.
func withTempDir(t *testing.T) string {
	t.Helper()
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	t.Cleanup(func() {
		_ = os.Chdir(origWd)
	})
	return tmpDir
}

func TestGetRuntimeDBPath(t *testing.T) {
	// getRuntimeDBPath uses os.Getwd() and joins .specify/session.db
	path := getRuntimeDBPath()

	if !strings.Contains(path, ".specify") {
		t.Errorf("expected path to contain '.specify', got: %s", path)
	}

	if !strings.HasSuffix(path, "session.db") {
		t.Errorf("expected path to end with 'session.db', got: %s", path)
	}
}

func TestOpenRuntimeDB(t *testing.T) {
	// Use a temporary directory so we don't interfere with the real DB
	tmpDir := withTempDir(t)

	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	// Verify the DB file was created
	dbPath := filepath.Join(tmpDir, ".specify", "session.db")
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Errorf("expected database file to exist at %s", dbPath)
	}

	// Verify we can query the runtime tables exist by doing a simple SELECT
	var count int
	if err := db.QueryRow("SELECT count(*) FROM sqlite_master WHERE type='table' AND name='sessions'").Scan(&count); err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	if count != 1 {
		t.Errorf("expected sessions table to exist, got count=%d", count)
	}
}

func TestEnsureSession(t *testing.T) {
	withTempDir(t)

	db, err := openRuntimeDB()
	if err != nil {
		t.Fatalf("openRuntimeDB failed: %v", err)
	}
	defer db.Close()

	sessionID := "test-session"

	// First call should create the session
	if err := ensureSession(db, sessionID); err != nil {
		t.Fatalf("ensureSession first call failed: %v", err)
	}

	// Verify exactly one row exists
	var count int
	if err := db.QueryRow("SELECT count(*) FROM sessions WHERE id = ?", sessionID).Scan(&count); err != nil {
		t.Fatalf("failed to count sessions: %v", err)
	}
	if count != 1 {
		t.Errorf("expected 1 session row after first ensureSession, got %d", count)
	}

	// Verify row has expected values
	var status, agent string
	if err := db.QueryRow("SELECT status, agent FROM sessions WHERE id = ?", sessionID).Scan(&status, &agent); err != nil {
		t.Fatalf("failed to query session: %v", err)
	}
	if status != "active" {
		t.Errorf("expected status='active', got %q", status)
	}
	if agent != "cli" {
		t.Errorf("expected agent='cli', got %q", agent)
	}

	// Second call should be idempotent (no error, no new row)
	if err := ensureSession(db, sessionID); err != nil {
		t.Fatalf("ensureSession second call failed: %v", err)
	}

	if err := db.QueryRow("SELECT count(*) FROM sessions WHERE id = ?", sessionID).Scan(&count); err != nil {
		t.Fatalf("failed to count sessions after second call: %v", err)
	}
	if count != 1 {
		t.Errorf("expected still 1 session row after idempotent ensureSession, got %d", count)
	}
}
