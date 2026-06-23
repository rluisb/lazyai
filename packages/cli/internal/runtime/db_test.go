package runtime

import (
	"os"
	"path/filepath"
	"testing"
)

func TestBackupTo_WorksWithQuoteInPath(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "source.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec("CREATE TABLE t (v INTEGER)"); err != nil {
		t.Fatalf("create table: %v", err)
	}
	if _, err := db.Exec("INSERT INTO t (v) VALUES (1)"); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Path containing a single quote — must not panic or cause a SQL
	// injection / syntax error. The sanitization doubles the quote so
	// SQLite treats it as a literal character inside the string.
	destDir := t.TempDir()
	destPath := filepath.Join(destDir, "it's a backup.db")

	if err := db.BackupTo(destPath); err != nil {
		t.Fatalf("BackupTo with quoted path failed: %v", err)
	}

	if _, err := os.Stat(destPath); err != nil {
		t.Fatalf("backup file not created: %v", err)
	}
}

func TestBackupTo_NormalPath(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "source.db")
	db, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	if _, err := db.Exec("CREATE TABLE t (v INTEGER)"); err != nil {
		t.Fatalf("create table: %v", err)
	}

	destPath := filepath.Join(t.TempDir(), "backup.db")
	if err := db.BackupTo(destPath); err != nil {
		t.Fatalf("BackupTo failed: %v", err)
	}
	if _, err := os.Stat(destPath); err != nil {
		t.Fatalf("backup file not created: %v", err)
	}
}
