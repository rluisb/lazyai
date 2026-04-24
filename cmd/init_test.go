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

// TestInitNonInteractiveGlobalScope verifies that OpenCode at global scope
// installs correctly when HomeDir is provided.
func TestInitNonInteractiveGlobalScope(t *testing.T) {
	dir := t.TempDir()
	ensureTestLibraryFS(t)
	withWorkingDir(t, dir)

	home := t.TempDir()

	config := &wizard.WizardConfig{
		Interactive: false,
		HomeDir:     home,
		TargetDir:   dir,
		CLIScope:    types.SetupScopeGlobal,
		CLITools:    []types.ToolId{types.ToolIdOpenCode},
		CLIPreset:   types.PresetLevelMinimal,
		CLIName:     "global",
	}

	_, _ = captureOutput(t, func() {
		if err := runInitNonInteractive(config); err != nil {
			t.Fatalf("runInitNonInteractive: %v", err)
		}
	})
}
