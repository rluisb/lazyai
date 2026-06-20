package cmd

import (
	"path/filepath"
	"runtime"
	"testing"
)

func TestSafeRestorePathRejectsTraversal(t *testing.T) {
	root := t.TempDir()

	cases := []string{
		"../escape.txt",
		"../../etc/passwd",
		"sub/../../escape.txt",
		"a/b/../../../escape.txt",
	}
	for _, name := range cases {
		if _, err := safeRestorePath(root, name); err == nil {
			t.Errorf("expected traversal entry %q to be rejected, got nil error", name)
		}
	}
}

func TestSafeRestorePathRejectsAbsolute(t *testing.T) {
	root := t.TempDir()

	abs := "/etc/evil"
	if runtime.GOOS == "windows" {
		abs = `C:\Windows\evil`
	}
	if _, err := safeRestorePath(root, abs); err == nil {
		t.Errorf("expected absolute entry %q to be rejected, got nil error", abs)
	}
}

func TestSafeRestorePathAllowsLegitimateEntries(t *testing.T) {
	root := t.TempDir()

	cases := []string{
		".ai-setup.db",
		".specify/ledger.jsonl",
		".opencode/config.yaml",
		".specify/memory/handoff.md",
	}
	for _, name := range cases {
		dest, err := safeRestorePath(root, name)
		if err != nil {
			t.Errorf("expected legitimate entry %q to be allowed, got: %v", name, err)
			continue
		}
		want := filepath.Join(root, name)
		if dest != want {
			t.Errorf("entry %q resolved to %q, want %q", name, dest, want)
		}
	}
}
