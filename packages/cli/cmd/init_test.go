package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
	"github.com/rluisb/lazyai/packages/cli/tui/wizard"
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
	opencodeConfigPath := filepath.Join(dir, "opencode.json")
	opencodeConfig, err := os.ReadFile(opencodeConfigPath)
	if err != nil {
		t.Fatalf("read opencode config: %v", err)
	}
	var parsed map[string]any
	if err := json.Unmarshal(opencodeConfig, &parsed); err != nil {
		t.Fatalf("parse opencode config: %v", err)
	}
	if _, ok := parsed["default_agent"]; ok {
		t.Fatalf("did not expect default_agent in baseline OpenCode config")
	}
	if _, ok := parsed["mcp"].(map[string]any); ok {
		t.Fatalf("expected init to avoid baseline MCP in %s; got %s", opencodeConfigPath, string(opencodeConfig))
	}
	if !fileExists(filepath.Join(dir, "AGENTS.md")) {
		t.Fatal("expected AGENTS.md to exist")
	}
}

func TestInitNonInteractiveLeavesProjectProfileEmptyAndPreservesPlaceholders(t *testing.T) {
	dir := t.TempDir()
	ensureTestLibraryFS(t)
	withWorkingDir(t, dir)
	useReversa := false

	config := &wizard.WizardConfig{
		Interactive:   false,
		HomeDir:       testRepoRoot(t),
		TargetDir:     dir,
		CLIScope:      types.SetupScopeProject,
		CLITools:      []types.ToolId{types.ToolIdOpenCode},
		CLIPreset:     types.PresetLevelMinimal,
		CLIUseReversa: &useReversa,
	}

	captureOutput(t, func() {
		if err := runInitNonInteractive(config); err != nil {
			t.Fatalf("runInitNonInteractive: %v", err)
		}
	})

	storeData := readSeededStoreData(t, dir)
	if storeData.Config.ProjectOverview != "" || storeData.Config.NamingConventions != "" || storeData.Config.ErrorHandling != "" || storeData.Config.ApiConventions != "" {
		t.Fatalf("profile config should remain empty: %#v", storeData.Config)
	}
	if storeData.Config.ImportOrder != "" || storeData.Config.ProtectedBranch != "" {
		t.Fatalf("convention config should remain empty: %#v", storeData.Config)
	}
	if storeData.Config.TestCommand != "" || storeData.Config.LintCommand != "" || storeData.Config.BuildCommand != "" {
		t.Fatalf("command config should remain empty: %#v", storeData.Config)
	}
	if storeData.Config.CoverageThreshold != 80 {
		t.Fatalf("CoverageThreshold = %d, want 80", storeData.Config.CoverageThreshold)
	}

	agents, err := os.ReadFile(filepath.Join(dir, "AGENTS.md"))
	if err != nil {
		t.Fatalf("read AGENTS.md: %v", err)
	}
	content := string(agents)
	for _, want := range []string{
		"<!-- fill-in: project overview -->",
		"<!-- fill-in: naming convention -->",
		"<!-- fill-in: error handling pattern -->",
		"<!-- fill-in: API response convention -->",
		"<!-- fill-in: import order -->",
		"<!-- fill-in: protected branch -->",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("AGENTS.md missing %q\n%s", want, content)
		}
	}
}

func TestInitRemovedFlagsNotRegistered(t *testing.T) {
	removed := []string{
		"project-overview",
		"naming-conventions",
		"error-handling",
		"api-conventions",
		"import-order",
		"protected-branch",
		"test-command",
		"lint-command",
		"build-command",
		"coverage-threshold",
		"enable-obsidian",
		"obsidian-vault-path",
		"enable-qmd",
		"qmd-index-path",
		"enable-codegraph",
		"codegraph-data-path",
		"enable-graphify",
		"graphify-data-path",
	}
	for _, name := range removed {
		if initCmd.Flags().Lookup(name) != nil {
			t.Fatalf("removed flag %q is still registered", name)
		}
	}

	for _, name := range []string{"memory-path", "reversa", "no-reversa"} {
		if initCmd.Flags().Lookup(name) == nil {
			t.Fatalf("expected flag %q to remain registered", name)
		}
	}
}

func TestBuildScaffoldContextDefaultsSkippedCoverageTo80(t *testing.T) {
	dir := t.TempDir()
	config := &wizard.WizardConfig{
		Interactive: false,
		HomeDir:     testRepoRoot(t),
		TargetDir:   dir,
		CLIScope:    types.SetupScopeProject,
		CLITools:    []types.ToolId{types.ToolIdOpenCode},
		CLIPreset:   types.PresetLevelMinimal,
		CLIName:     "profile-test",
	}
	result := &wizard.WizardResult{
		Phase1: &wizard.Phase1Result{Scope: config.CLIScope, Tools: config.CLITools, ProjectName: config.CLIName},
		Phase2: &wizard.Phase2Result{Preset: config.CLIPreset},
	}

	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		t.Fatalf("buildScaffoldContext: %v", err)
	}
	if ctx.CoverageThreshold != 80 {
		t.Fatalf("CoverageThreshold = %d, want 80", ctx.CoverageThreshold)
	}
}

func TestBuildScaffoldContextWorkspaceUsesTargetAsPlanningRepoAndFlagAsWorkspaceRoot(t *testing.T) {
	planningRepo := t.TempDir()
	workspaceRoot := t.TempDir()
	useReversa := false
	config := &wizard.WizardConfig{
		Interactive:      false,
		HomeDir:          testRepoRoot(t),
		TargetDir:        planningRepo,
		CLIScope:         types.SetupScopeWorkspace,
		CLITools:         []types.ToolId{types.ToolIdOpenCode},
		CLIPreset:        types.PresetLevelMinimal,
		CLIName:          "workspace-test",
		CLIWorkspaceRoot: workspaceRoot,
		CLIUseReversa:    &useReversa,
	}
	result := &wizard.WizardResult{
		Phase1: &wizard.Phase1Result{Scope: config.CLIScope, Tools: config.CLITools, ProjectName: config.CLIName},
		Phase2: &wizard.Phase2Result{Preset: config.CLIPreset, UseReversa: &useReversa},
	}

	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		t.Fatalf("buildScaffoldContext: %v", err)
	}
	if ctx.PlanningRepoPath != planningRepo {
		t.Fatalf("PlanningRepoPath = %q, want target planning repo %q", ctx.PlanningRepoPath, planningRepo)
	}
	if ctx.WorkspaceRoot != workspaceRoot {
		t.Fatalf("WorkspaceRoot = %q, want CLI workspace root %q", ctx.WorkspaceRoot, workspaceRoot)
	}
}

func TestBuildScaffoldContextGlobalClearsPlanningAndWorkspaceRoots(t *testing.T) {
	dir := t.TempDir()
	config := &wizard.WizardConfig{
		Interactive: false,
		HomeDir:     testRepoRoot(t),
		TargetDir:   dir,
		CLIScope:    types.SetupScopeGlobal,
		CLITools:    []types.ToolId{types.ToolIdClaudeCode},
		CLIPreset:   types.PresetLevelMinimal,
		CLIName:     "global-test",
	}
	result := &wizard.WizardResult{
		Phase1: &wizard.Phase1Result{Scope: config.CLIScope, Tools: config.CLITools, ProjectName: config.CLIName},
		Phase2: &wizard.Phase2Result{Preset: config.CLIPreset},
	}

	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		t.Fatalf("buildScaffoldContext: %v", err)
	}
	if ctx.PlanningRepoPath != "" || ctx.WorkspaceRoot != "" {
		t.Fatalf("global roots = planning %q workspace %q, want both empty", ctx.PlanningRepoPath, ctx.WorkspaceRoot)
	}
}

func TestBuildScaffoldContextNoReversaSkipsScoutPopulation(t *testing.T) {
	dir := t.TempDir()
	writeGoProjectForReversaTest(t, dir)
	useReversa := false
	config := &wizard.WizardConfig{
		Interactive:   false,
		HomeDir:       testRepoRoot(t),
		TargetDir:     dir,
		CLIScope:      types.SetupScopeProject,
		CLITools:      []types.ToolId{types.ToolIdOpenCode},
		CLIPreset:     types.PresetLevelMinimal,
		CLIUseReversa: &useReversa,
	}
	result := &wizard.WizardResult{
		Phase1: &wizard.Phase1Result{Scope: config.CLIScope, Tools: config.CLITools, ProjectName: "reversa-off"},
		Phase2: &wizard.Phase2Result{Preset: config.CLIPreset, UseReversa: &useReversa},
	}

	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		t.Fatalf("buildScaffoldContext: %v", err)
	}
	if ctx.SurfaceData != nil || ctx.PrimaryLanguage != "" || ctx.PackageManager != "" {
		t.Fatalf("Scout populated context despite UseReversa=false: language=%q packageManager=%q surface=%#v", ctx.PrimaryLanguage, ctx.PackageManager, ctx.SurfaceData)
	}
}

func TestBuildScaffoldContextReversaPopulatesScoutValues(t *testing.T) {
	dir := t.TempDir()
	writeGoProjectForReversaTest(t, dir)
	useReversa := true
	config := &wizard.WizardConfig{
		Interactive:   false,
		HomeDir:       testRepoRoot(t),
		TargetDir:     dir,
		CLIScope:      types.SetupScopeProject,
		CLITools:      []types.ToolId{types.ToolIdOpenCode},
		CLIPreset:     types.PresetLevelMinimal,
		CLIUseReversa: &useReversa,
	}
	result := &wizard.WizardResult{
		Phase1: &wizard.Phase1Result{Scope: config.CLIScope, Tools: config.CLITools, ProjectName: "reversa-on"},
		Phase2: &wizard.Phase2Result{Preset: config.CLIPreset, UseReversa: &useReversa},
	}

	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		t.Fatalf("buildScaffoldContext: %v", err)
	}
	if ctx.SurfaceData == nil {
		t.Fatal("expected Scout surface data to be populated")
	}
	if ctx.PrimaryLanguage != "Go" {
		t.Fatalf("PrimaryLanguage = %q, want Go", ctx.PrimaryLanguage)
	}
	if ctx.PackageManager != "go modules" {
		t.Fatalf("PackageManager = %q, want go modules", ctx.PackageManager)
	}
}

func TestBuildScaffoldContextDefaultsReversaEnabled(t *testing.T) {
	dir := t.TempDir()
	writeGoProjectForReversaTest(t, dir)
	config := &wizard.WizardConfig{
		Interactive: false,
		HomeDir:     testRepoRoot(t),
		TargetDir:   dir,
		CLIScope:    types.SetupScopeProject,
		CLITools:    []types.ToolId{types.ToolIdOpenCode},
		CLIPreset:   types.PresetLevelMinimal,
	}
	result := &wizard.WizardResult{
		Phase1: &wizard.Phase1Result{Scope: config.CLIScope, Tools: config.CLITools, ProjectName: "reversa-default"},
		Phase2: &wizard.Phase2Result{Preset: config.CLIPreset},
	}

	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		t.Fatalf("buildScaffoldContext: %v", err)
	}
	if ctx.PrimaryLanguage != "Go" {
		t.Fatalf("PrimaryLanguage = %q, want Go", ctx.PrimaryLanguage)
	}
}

func TestRunInitRejectsConflictingReversaFlags(t *testing.T) {
	cmd := &cobra.Command{}
	cmd.Flags().Bool("reversa", false, "")
	cmd.Flags().Bool("no-reversa", false, "")
	if err := cmd.Flags().Set("reversa", "true"); err != nil {
		t.Fatalf("set reversa: %v", err)
	}
	if err := cmd.Flags().Set("no-reversa", "true"); err != nil {
		t.Fatalf("set no-reversa: %v", err)
	}

	err := runInit(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "--reversa and --no-reversa cannot be used together") {
		t.Fatalf("runInit conflicting flags error = %v", err)
	}
}

func TestInitNonInteractiveDryRunWritesNoFiles(t *testing.T) {
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
		DryRun:      true,
	}

	captureOutput(t, func() {
		if err := runInitNonInteractive(config); err != nil {
			t.Fatalf("runInitNonInteractive: %v", err)
		}
	})

	for _, rel := range []string{".ai-setup.db", "AGENTS.md", ".opencode", ".ai", ".specify"} {
		if fileExists(filepath.Join(dir, rel)) {
			t.Fatalf("dry-run created %s", rel)
		}
	}
}

func writeGoProjectForReversaTest(t *testing.T, dir string) {
	t.Helper()
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/reversa-test\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "main.go"), []byte("package main\n\nfunc main() {}\n"), 0o644); err != nil {
		t.Fatalf("write main.go: %v", err)
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
