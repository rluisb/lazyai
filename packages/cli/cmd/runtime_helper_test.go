package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

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
	tmpDir := t.TempDir()
	origWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get wd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origWd)

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
