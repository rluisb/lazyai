package tokenrent

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestCheckPassesUnderBudget(t *testing.T) {
	root := t.TempDir()
	canonicalDir := filepath.Join(root, CanonicalSubdir)
	mustWrite(t, filepath.Join(canonicalDir, "agents", "builder.md"), []byte("builder"))
	mustWrite(t, filepath.Join(canonicalDir, "skills", "diagnose.md"), []byte("diagnose"))

	result, err := Check(canonicalDir, filepath.Join(root, OverrideRelPath), DefaultBudgetBytes)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if result.TotalBytes != len("builder")+len("diagnose") {
		t.Fatalf("TotalBytes = %d", result.TotalBytes)
	}
	if result.OverrideUsed {
		t.Fatal("OverrideUsed should be false")
	}
}

func TestCheckFailsOverBudgetWithoutOverride(t *testing.T) {
	root := t.TempDir()
	canonicalDir := filepath.Join(root, CanonicalSubdir)
	mustWrite(t, filepath.Join(canonicalDir, "agents", "builder.md"), make([]byte, DefaultBudgetBytes+1))

	_, err := Check(canonicalDir, filepath.Join(root, OverrideRelPath), DefaultBudgetBytes)
	var budgetErr *BudgetError
	if !errors.As(err, &budgetErr) {
		t.Fatalf("expected BudgetError, got %v", err)
	}
	want := "Library budget exceeded: 50001 / 50000 bytes. Override: add .lazyai/token-rent-override with justification."
	if budgetErr.Error() != want {
		t.Fatalf("error = %q, want %q", budgetErr.Error(), want)
	}
}

func TestCheckPassesWithValidOverride(t *testing.T) {
	root := t.TempDir()
	canonicalDir := filepath.Join(root, CanonicalSubdir)
	overridePath := filepath.Join(root, OverrideRelPath)
	mustWrite(t, filepath.Join(canonicalDir, "agents", "builder.md"), make([]byte, DefaultBudgetBytes+200))
	mustWrite(t, overridePath, []byte("budget: 60000\nreason: approved expansion\napproved_by: test\n"))

	result, err := Check(canonicalDir, overridePath, DefaultBudgetBytes)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if !result.OverrideUsed {
		t.Fatal("OverrideUsed should be true")
	}
	if result.Override == nil || result.Override.Reason != "approved expansion" {
		t.Fatalf("Override = %#v", result.Override)
	}
}

func TestCheckFailsWithInvalidOverrideReason(t *testing.T) {
	root := t.TempDir()
	canonicalDir := filepath.Join(root, CanonicalSubdir)
	overridePath := filepath.Join(root, OverrideRelPath)
	mustWrite(t, filepath.Join(canonicalDir, "agents", "builder.md"), make([]byte, DefaultBudgetBytes+100))
	mustWrite(t, overridePath, []byte("budget: 60000\nreason: \n"))

	_, err := Check(canonicalDir, overridePath, DefaultBudgetBytes)
	var overrideErr *OverrideError
	if !errors.As(err, &overrideErr) {
		t.Fatalf("expected OverrideError, got %v", err)
	}
	if overrideErr.Reason != "reason must be non-empty" {
		t.Fatalf("OverrideError.Reason = %q", overrideErr.Reason)
	}
}

func TestCheckExcludesGitkeepFromBudget(t *testing.T) {
	root := t.TempDir()
	canonicalDir := filepath.Join(root, CanonicalSubdir)
	mustWrite(t, filepath.Join(canonicalDir, "agents", ".gitkeep"), make([]byte, 64000))
	mustWrite(t, filepath.Join(canonicalDir, "agents", "builder.md"), []byte("builder"))

	result, err := Check(canonicalDir, filepath.Join(root, OverrideRelPath), DefaultBudgetBytes)
	if err != nil {
		t.Fatalf("Check returned error: %v", err)
	}
	if result.TotalBytes != len("builder") {
		t.Fatalf("TotalBytes = %d, want %d", result.TotalBytes, len("builder"))
	}
	if len(result.Files) != 1 || result.Files[0].Path != "agents/builder.md" {
		t.Fatalf("Files = %#v", result.Files)
	}
}

func TestCurrentCanonicalLibraryWithinBudget(t *testing.T) {
	root := repoRootFromCaller(t)
	result, err := CheckProject(root)
	if err != nil {
		t.Fatalf("CheckProject returned error: %v", err)
	}
	if result.TotalBytes > DefaultBudgetBytes {
		t.Fatalf("current canonical library = %d bytes, budget = %d", result.TotalBytes, DefaultBudgetBytes)
	}
}

func mustWrite(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("MkdirAll(%s): %v", path, err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("WriteFile(%s): %v", path, err)
	}
}

func repoRootFromCaller(t *testing.T) string {
	t.Helper()
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller failed")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "..", ".."))
}
