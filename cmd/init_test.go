package cmd

import (
	"path/filepath"
	"strings"
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

// TestInitNonInteractiveScopeFilter_MixedList verifies that tools not
// supported at the chosen scope (copilot × global) are skipped with a WARN
// and the install proceeds for the remaining tools.
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

	_, stderr := captureOutput(t, func() {
		if err := runInitNonInteractive(config); err != nil {
			t.Fatalf("runInitNonInteractive: %v", err)
		}
	})
	if !strings.Contains(stderr, "skipping tool \"copilot\"") {
		t.Errorf("expected WARN about skipping copilot, got stderr:\n%s", stderr)
	}
	// Copilot is dropped; claude-code remains. config.CLITools should no
	// longer contain copilot after runInitNonInteractive returns.
	for _, tool := range config.CLITools {
		if tool == types.ToolIdCopilot {
			t.Errorf("copilot was not filtered out of config.CLITools")
		}
	}
}

// TestInitNonInteractiveScopeFilter_AllUnsupported verifies that a tool list
// containing only incompatible tools returns a non-nil error.
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

	err := runInitNonInteractive(config)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "no tools remain") {
		t.Errorf("expected 'no tools remain' error, got: %v", err)
	}
}
