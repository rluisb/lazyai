package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"
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

func TestRunBuildPluginSupportsCopilotCliTarget(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "copilot-bundle")
	cmd := &cobra.Command{}
	cmd.Flags().String("out", "", "")
	cmd.Flags().String("target", "claude", "")
	cmd.Flags().Bool("force", false, "")
	_ = cmd.Flags().Set("out", outDir)
	_ = cmd.Flags().Set("target", "copilot-cli")

	_, _ = captureOutput(t, func() {
		if err := runBuildPlugin(cmd, nil); err != nil {
			t.Fatalf("runBuildPlugin: %v", err)
		}
	})

	for _, rel := range []string{"plugin.json", "agents/guide.agent.md", ".mcp.json"} {
		if _, err := os.Stat(filepath.Join(outDir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
}

func TestRunBuildPluginRejectsUnknownTarget(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().String("out", "", "")
	cmd.Flags().String("target", "claude", "")
	cmd.Flags().Bool("force", false, "")
	_ = cmd.Flags().Set("out", filepath.Join(t.TempDir(), "bundle"))
	_ = cmd.Flags().Set("target", "gemini")

	err := runBuildPlugin(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "unsupported bundle target") {
		t.Fatalf("expected unsupported-target error, got %v", err)
	}
}
