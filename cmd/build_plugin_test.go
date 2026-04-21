package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreflightOutDir_MissingIsOK(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "does-not-exist")
	if err := preflightOutDir(outDir, false); err != nil {
		t.Errorf("missing outDir must be OK: %v", err)
	}
}

func TestPreflightOutDir_EmptyIsOK(t *testing.T) {
	outDir := t.TempDir() // empty by default
	if err := preflightOutDir(outDir, false); err != nil {
		t.Errorf("empty outDir must be OK: %v", err)
	}
}

func TestPreflightOutDir_NonEmptyRejectedWithoutForce(t *testing.T) {
	outDir := t.TempDir()
	if err := os.WriteFile(filepath.Join(outDir, "existing.txt"), []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	err := preflightOutDir(outDir, false)
	if err == nil {
		t.Error("expected error for non-empty outDir without --force")
	}
	if !strings.Contains(err.Error(), "pass --force") {
		t.Errorf("error message should mention --force: %v", err)
	}
}

func TestPreflightOutDir_ForceWipesExisting(t *testing.T) {
	outDir := t.TempDir()
	seed := filepath.Join(outDir, "existing.txt")
	if err := os.WriteFile(seed, []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	if err := preflightOutDir(outDir, true); err != nil {
		t.Errorf("--force must wipe cleanly: %v", err)
	}
	if _, err := os.Stat(seed); !os.IsNotExist(err) {
		t.Errorf("seed file should be gone after --force; err=%v", err)
	}
}

func TestPreflightOutDir_RejectsFileAtPath(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "file-not-dir")
	if err := os.WriteFile(outDir, []byte("x"), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}
	err := preflightOutDir(outDir, true)
	if err == nil {
		t.Error("expected error when outDir path is a regular file")
	}
}
