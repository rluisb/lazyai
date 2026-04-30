package cmd

import (
	"encoding/json"
	"os"
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
	opencodeConfigPath := filepath.Join(dir, ".opencode", "opencode.jsonc")
	opencodeConfig, err := os.ReadFile(opencodeConfigPath)
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(opencodeConfig, &parsed); err != nil {
		t.Fatalf("parse opencode config: %v", err)
	}
	if _, ok := parsed["mcp"].(map[string]any); !ok {
		t.Fatalf("expected init to compile MCP servers into %s; got %s", opencodeConfigPath, string(opencodeConfig))
	}
	if !fileExists(filepath.Join(dir, "AGENTS.md")) {
		t.Fatal("expected AGENTS.md to exist")
	}
}

func TestInitNonInteractiveDefaultsExistingSetupPolicy(t *testing.T) {
	dir := t.TempDir()
	config := &wizard.WizardConfig{
		Interactive: false,
		HomeDir:     testRepoRoot(t),
		TargetDir:   dir,
		CLIScope:    types.SetupScopeProject,
		CLITools:    []types.ToolId{types.ToolIdOpenCode},
		CLIPreset:   types.PresetLevelMinimal,
		CLIName:     "policy-test",
	}
	result := &wizard.WizardResult{
		Phase1: &wizard.Phase1Result{Scope: config.CLIScope, Tools: config.CLITools, ProjectName: config.CLIName},
		Phase2: &wizard.Phase2Result{Preset: config.CLIPreset},
	}

	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		t.Fatalf("buildScaffoldContext: %v", err)
	}
	if config.CLIExistingSetupPolicy != types.SetupPolicyAbsorb {
		t.Fatalf("CLIExistingSetupPolicy = %q, want %q", config.CLIExistingSetupPolicy, types.SetupPolicyAbsorb)
	}
	if ctx.ExistingSetupPolicy != types.SetupPolicyAbsorb {
		t.Fatalf("ExistingSetupPolicy = %q, want %q", ctx.ExistingSetupPolicy, types.SetupPolicyAbsorb)
	}
	if ctx.Strategy != types.ConflictStrategySkip {
		t.Fatalf("Strategy = %q, want %q", ctx.Strategy, types.ConflictStrategySkip)
	}
}

func TestBuildScaffoldContextMapsExistingSetupPolicyToConflictStrategy(t *testing.T) {
	for _, tc := range []struct {
		name   string
		policy types.SetupPolicy
		want   types.ConflictStrategy
	}{
		{name: "absorb preserves existing files", policy: types.SetupPolicyAbsorb, want: types.ConflictStrategySkip},
		{name: "adapt mvp falls back to absorb preservation", policy: types.SetupPolicyAdapt, want: types.ConflictStrategySkip},
		{name: "backup-only keeps backup replace behavior", policy: types.SetupPolicyBackupOnly, want: types.ConflictStrategyAlign},
	} {
		t.Run(tc.name, func(t *testing.T) {
			dir := t.TempDir()
			config := &wizard.WizardConfig{
				Interactive:            false,
				HomeDir:                testRepoRoot(t),
				TargetDir:              dir,
				CLIScope:               types.SetupScopeProject,
				CLITools:               []types.ToolId{types.ToolIdOpenCode},
				CLIPreset:              types.PresetLevelMinimal,
				CLIName:                "policy-test",
				CLIExistingSetupPolicy: tc.policy,
			}
			result := &wizard.WizardResult{
				Phase1: &wizard.Phase1Result{Scope: config.CLIScope, Tools: config.CLITools, ProjectName: config.CLIName},
				Phase2: &wizard.Phase2Result{Preset: config.CLIPreset},
			}

			ctx, err := buildScaffoldContext(result, config)
			if err != nil {
				t.Fatalf("buildScaffoldContext: %v", err)
			}
			if ctx.Strategy != tc.want {
				t.Fatalf("Strategy = %q, want %q", ctx.Strategy, tc.want)
			}
		})
	}
}

func TestParseExistingSetupPolicy(t *testing.T) {
	for _, raw := range []string{"", "absorb", "adapt", "backup-only"} {
		if _, err := parseExistingSetupPolicy(raw); err != nil {
			t.Fatalf("parseExistingSetupPolicy(%q): %v", raw, err)
		}
	}
	if _, err := parseExistingSetupPolicy("invalid"); err == nil {
		t.Fatal("parseExistingSetupPolicy invalid returned nil error")
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
