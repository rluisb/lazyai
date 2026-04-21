package cmd

import (
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
	"github.com/ricardoborges-teachable/ai-setup/tui/wizard"
)

func TestInitNonInteractiveHappyPath(t *testing.T) {
	dir := t.TempDir()
	ensureTestLibraryFS(t)
	withWorkingDir(t, dir)

	config := &wizard.WizardConfig{
		Interactive: false,
		HomeDir:     testRepoRoot(t),
		TargetDir:   dir,
		CLIScope:    types.SetupScopeProject,
		CLITools:    []types.ToolId{types.ToolIdOpenCode},
		CLIPreset:   types.PresetLevelMinimal,
	}

	captureOutput(t, func() {
		if err := runInitNonInteractive(config); err != nil {
			t.Fatalf("runInitNonInteractive: %v", err)
		}
	})

	storeData := readSeededStoreData(t, dir)
	if storeData.Config.SetupScope != types.SetupScopeProject {
		t.Fatalf("SetupScope = %q, want %q", storeData.Config.SetupScope, types.SetupScopeProject)
	}
	if storeData.Config.ProjectName != filepath.Base(dir) {
		t.Fatalf("ProjectName = %q, want %q", storeData.Config.ProjectName, filepath.Base(dir))
	}
	if len(storeData.Config.Tools) != 1 || storeData.Config.Tools[0] != types.ToolIdOpenCode {
		t.Fatalf("Tools = %v, want [%s]", storeData.Config.Tools, types.ToolIdOpenCode)
	}
	if !fileExists(filepath.Join(dir, ".ai-setup.db")) {
		t.Fatal("expected .ai-setup.db to exist")
	}
	if !fileExists(filepath.Join(dir, ".ai", "mcp.json")) {
		t.Fatal("expected .ai/mcp.json to exist")
	}
	if !fileExists(filepath.Join(dir, "AGENTS.md")) {
		t.Fatal("expected AGENTS.md to exist")
	}
}

// TestInitNonInteractiveScopeFilter_MixedList verifies that when copilot
// is requested at global scope but the probes fail (no CLI or ~/.copilot/),
// the install proceeds for the remaining tools.
func TestInitNonInteractiveScopeFilter_MixedList(t *testing.T) {
	dir := t.TempDir()
	ensureTestLibraryFS(t)
	withWorkingDir(t, dir)

	home := t.TempDir()

	config := &wizard.WizardConfig{
		Interactive: false,
		HomeDir:     home,
		TargetDir:   dir,
		CLIScope:    types.SetupScopeGlobal,
		CLITools:    []types.ToolId{types.ToolIdClaudeCode, types.ToolIdCopilot},
		CLIPreset:   types.PresetLevelMinimal,
		CLIName:     "global",
	}

	_, _ = captureOutput(t, func() {
		if err := runInitNonInteractive(config); err != nil {
			t.Fatalf("runInitNonInteractive: %v", err)
		}
	})
	// Copilot is now supported at scope level (filtered at adapter level when probes fail)
	// The install should proceed with both tools; Copilot skips silently
}

// TestInitNonInteractiveScopeFilter_AllUnsupported verifies that when copilot
// is the only tool at global scope, the install still proceeds (even though
// the adapter will skip due to failed probes).
func TestInitNonInteractiveScopeFilter_AllUnsupported(t *testing.T) {
	dir := t.TempDir()
	ensureTestLibraryFS(t)
	withWorkingDir(t, dir)

	config := &wizard.WizardConfig{
		Interactive: false,
		HomeDir:     t.TempDir(),
		TargetDir:   dir,
		CLIScope:    types.SetupScopeGlobal,
		CLITools:    []types.ToolId{types.ToolIdCopilot},
		CLIPreset:   types.PresetLevelMinimal,
		CLIName:     "global",
	}

	// Copilot is now supported at scope level; install should proceed
	// (adapter skips silently when probes fail)
	err := runInitNonInteractive(config)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}
