package eject

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRunRemovesMetadataKeepsNativeFiles(t *testing.T) {
	dir := t.TempDir()
	mustWrite := func(rel, content string) {
		path := filepath.Join(dir, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	mustWrite(".ai/lazyai.json", "{}")
	mustWrite(".ai/lock.json", "{}")
	mustWrite(".ai/migration-report.md", "# report\n")
	mustWrite(".ai/agents/reviewer.md", "# Reviewer\n")
	mustWrite(".ai-setup.json", "{}")
	mustWrite(".ai-setup.db", "db")

	plan := Inspect(dir)
	if len(plan.Existing) != 5 {
		t.Fatalf("Inspect existing = %d, want 5 (%+v)", len(plan.Existing), plan.Existing)
	}

	result, err := Run(dir)
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if len(result.Removed) != 5 {
		t.Fatalf("removed = %d, want 5", len(result.Removed))
	}

	for _, rel := range []string{".ai/lazyai.json", ".ai/lock.json", ".ai/migration-report.md", ".ai-setup.json", ".ai-setup.db"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); !os.IsNotExist(err) {
			t.Fatalf("expected %s removed, err=%v", rel, err)
		}
	}
	if _, err := os.Stat(filepath.Join(dir, ".ai", "agents", "reviewer.md")); err != nil {
		t.Fatalf("expected native file kept: %v", err)
	}
}
