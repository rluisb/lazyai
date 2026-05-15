package agentmemory

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestEnsureDataDirCreatesPrivateDirectoryAndGitignore(t *testing.T) {
	dir := filepath.Join(t.TempDir(), "agent-memory")
	if err := EnsureDataDir(dir); err != nil {
		t.Fatalf("EnsureDataDir() error = %v", err)
	}

	info, err := os.Stat(dir)
	if err != nil {
		t.Fatalf("stat dir: %v", err)
	}
	if !info.IsDir() {
		t.Fatalf("%s is not a directory", dir)
	}
	if runtime.GOOS != "windows" && info.Mode().Perm() != 0o700 {
		t.Fatalf("dir mode = %o, want 0700", info.Mode().Perm())
	}

	content, err := os.ReadFile(filepath.Join(dir, ".gitignore"))
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if string(content) != "*.sqlite\n*.sqlite-*\n" {
		t.Fatalf(".gitignore = %q", string(content))
	}
}
