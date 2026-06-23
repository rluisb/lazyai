package db

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestOpen_RestrictsFilePermissions(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows does not expose POSIX 0600 file mode semantics")
	}
	dir := t.TempDir()
	dbPath := filepath.Join(dir, "test.db")

	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open %s: %v", dbPath, err)
	}
	t.Cleanup(func() { db.Close() })

	info, err := os.Stat(dbPath)
	if err != nil {
		t.Fatalf("stat db file: %v", err)
	}

	mode := info.Mode().Perm()
	if mode != 0o600 {
		t.Errorf("db file permissions = %o, want 0600", mode)
	}
}

func TestOpen_MemoryDoesNotChmod(t *testing.T) {
	db, err := Open(":memory:")
	if err != nil {
		t.Fatalf("Open :memory:: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	// Should not error for in-memory databases.
	if db == nil {
		t.Fatal("expected non-nil DB for in-memory database")
	}
}
