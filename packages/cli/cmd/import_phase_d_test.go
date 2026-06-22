package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
)

func TestImportWritesReportManifestAndRawPreservation(t *testing.T) {
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
	mustWrite(".opencode/agents/reviewer.md", "# Reviewer\n\nReview carefully.\n")
	mustWrite(".pi/agents/researcher.md", "# Researcher\n")
	mustWrite(".omp/skills/diagnose/SKILL.md", "# Diagnose\n")
	mustWrite(".github/copilot-instructions.md", "# Copilot\n")

	withWorkingDir(t, dir)
	cmd := newImportCommandForTest("", true, false)
	_, _ = captureOutput(t, func() {
		if err := runImport(cmd, nil); err != nil {
			t.Fatalf("runImport: %v", err)
		}
	})

	for _, rel := range []string{
		".ai/lazyai.json",
		".ai/migration-report.md",
		".ai/agents/reviewer.md",
		".ai/adapters/pi/raw/.pi/agents/researcher.md",
		".ai/adapters/omp/raw/.omp/skills/diagnose/SKILL.md",
		".ai/adapters/copilot/raw/.github/copilot-instructions.md",
		".ai/adapters/opencode/raw/.opencode/agents/reviewer.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected %s: %v", rel, err)
		}
	}
	for _, rel := range []string{
		".opencode/agents/reviewer.md",
		".pi/agents/researcher.md",
		".omp/skills/diagnose/SKILL.md",
		".github/copilot-instructions.md",
	} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected native source kept at %s: %v", rel, err)
		}
	}
}

func TestImportThenEjectPreservesNativeFilesAndRemovesMetadata(t *testing.T) {
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
	mustWrite(".opencode/agents/reviewer.md", "# Reviewer\n\nReview carefully.\n")
	mustWrite(".pi/agents/researcher.md", "# Researcher\n")

	withWorkingDir(t, dir)
	_, _ = captureOutput(t, func() {
		if err := runImport(newImportCommandForTest("", true, false), nil); err != nil {
			t.Fatalf("runImport: %v", err)
		}
	})
	_, _ = captureOutput(t, func() {
		if err := runEject(newEjectCommandForTest(true), nil); err != nil {
			t.Fatalf("runEject: %v", err)
		}
	})

	for _, rel := range []string{".opencode/agents/reviewer.md", ".pi/agents/researcher.md"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); err != nil {
			t.Fatalf("expected native file kept at %s: %v", rel, err)
		}
	}
	for _, rel := range []string{".ai/lazyai.json", ".ai/migration-report.md", ".ai-setup.json", ".ai-setup.db"} {
		if _, err := os.Stat(filepath.Join(dir, rel)); !os.IsNotExist(err) {
			t.Fatalf("expected metadata %s removed, err=%v", rel, err)
		}
	}
}

func newImportCommandForTest(tool string, noInteractive, preview bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("tool", "", "")
	cmd.Flags().Bool("no-interactive", false, "")
	cmd.Flags().Bool("preview", false, "")
	cmd.Flags().String("strategy", "smart", "")
	cmd.Flags().Bool("verbose", false, "")
	cmd.Flags().Bool("skip-backup", false, "")
	_ = cmd.Flags().Set("tool", tool)
	if noInteractive {
		_ = cmd.Flags().Set("no-interactive", "true")
	}
	if preview {
		_ = cmd.Flags().Set("preview", "true")
	}
	return cmd
}

func newEjectCommandForTest(noInteractive bool) *cobra.Command {
	cmd := &cobra.Command{}
	cmd.Flags().String("dir", "", "")
	cmd.Flags().Bool("no-interactive", false, "")
	if noInteractive {
		_ = cmd.Flags().Set("no-interactive", "true")
	}
	return cmd
}
