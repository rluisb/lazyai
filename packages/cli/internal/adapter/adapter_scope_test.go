package adapter

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// newScopeTestContext returns an AdapterContext rooted at two separate temp
// dirs for TargetDir and HomeDir. Tests use this to assert paths land under
// the correct root per scope (R-3 mitigation: never read os.UserHomeDir()).
func newScopeTestContext(t *testing.T, scope types.SetupScope) (*AdapterContext, string, string) {
	t.Helper()
	target := t.TempDir()
	home := t.TempDir()
	ctx := &AdapterContext{
		TargetDir:  target,
		HomeDir:    home,
		SetupScope: scope,
		LibraryFS:  createTestFS(),
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"researcher"},
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}
	return ctx, target, home
}

// TestAdapter_ScopeParity asserts that each adapter writes its primary tree
// under the scope-correct root (research §2). It does not exhaustively list
// every emitted file — just the scope-defining directories — so that future
// content changes don't churn this test.
func TestAdapter_ScopeParity(t *testing.T) {
	type caseRow struct {
		name    string
		adapter ToolAdapter
		scope   types.SetupScope
		root    func(target, home string) string
	}

	rows := []caseRow{
		{"claude_project", &ClaudeCodeAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"claude_workspace", &ClaudeCodeAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"claude_global", &ClaudeCodeAdapter{}, types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".claude") }},

		{"opencode_project", &OpenCodeAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"opencode_workspace", &OpenCodeAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"opencode_global", &OpenCodeAdapter{}, types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".config", "opencode") }},

		{"copilot_project", &CopilotAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".github") }},
		{"copilot_workspace", &CopilotAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".github") }},
		// copilot_global is exercised separately below.

		{"pi_project", &PiAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".pi") }},
		{"pi_workspace", &PiAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".pi") }},
		{"pi_global", &PiAdapter{}, types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".pi") }},

		{"antigravity_project", &AntigravityAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".gemini") }},
		{"antigravity_workspace", &AntigravityAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".gemini") }},
		{"antigravity_global", &AntigravityAdapter{}, types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".gemini") }},
	}

	for _, row := range rows {
		row := row
		t.Run(row.name, func(t *testing.T) {
			ctx, target, home := newScopeTestContext(t, row.scope)
			records, err := row.adapter.Install(ctx)
			if err != nil {
				t.Fatalf("Install: %v", err)
			}
			if len(records) == 0 {
				t.Fatal("no file records produced")
			}
			wantRoot := row.root(target, home)
			if !files.DirExists(wantRoot) {
				t.Errorf("expected directory %q to exist after Install", wantRoot)
			}

			// Leak check: when scope is project/workspace, no path should
			// contain the home dir; when scope is global, no path should
			// contain the target dir.
			for _, rec := range records {
				p := rec.Path
				switch row.scope {
				case types.SetupScopeProject, types.SetupScopeWorkspace:
					if strings.HasPrefix(p, home) {
						t.Errorf("project/workspace scope wrote under home dir: %q", p)
					}
				case types.SetupScopeGlobal:
					// records may be stored as relative paths rooted at
					// target (adapter writes under home for global). Assert
					// the absolute path on disk lives under home.
					if filepath.IsAbs(p) && strings.HasPrefix(p, target) {
						t.Errorf("global scope wrote under target dir: %q", p)
					}
				}
			}
		})
	}
}

// TestCopilotAdapter_GlobalScope_Skips verifies the adapter early-returns
// (no error, no records, no writes) when scope=global and probes fail.
func TestCopilotAdapter_GlobalScope_Skips(t *testing.T) {
	ctx, target, home := newScopeTestContext(t, types.SetupScopeGlobal)
	t.Setenv("PATH", t.TempDir())
	adapter := &CopilotAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("Install at scope=global must not error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records at scope=global, got %d", len(records))
	}
	if files.DirExists(filepath.Join(home, ".github")) {
		t.Error("copilot must not create ~/.github at scope=global")
	}
	if files.DirExists(filepath.Join(target, ".github")) {
		t.Error("copilot must not create <target>/.github at scope=global")
	}
}

// TestCopilotAdapter_GlobalScope_Emits verifies the adapter correctly emits
// agents, instructions, and chatmodes under ~/.copilot/ at global scope.
func TestCopilotAdapter_GlobalScope_Emits(t *testing.T) {
	ctx, _, home := newScopeTestContext(t, types.SetupScopeGlobal)
	// Create ~/.copilot/ so the probe passes
	copilotHome := filepath.Join(home, ".copilot")
	if err := files.EnsureDir(copilotHome); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	adapter := &CopilotAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("Install at scope=global: %v", err)
	}
	if len(records) == 0 {
		t.Errorf("expected records at scope=global with ~/.copilot/, got 0")
	}
	// Verify agents directory was created
	if !files.DirExists(filepath.Join(copilotHome, "agents")) {
		t.Error("agents directory not created under ~/.copilot/")
	}
	// Verify instructions directory was created
	if !files.DirExists(filepath.Join(copilotHome, "instructions")) {
		t.Error("instructions directory not created under ~/.copilot/")
	}
}
