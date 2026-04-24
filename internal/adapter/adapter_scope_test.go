package adapter

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// newScopeTestContext returns an AdapterContext rooted at two separate temp
// dirs for TargetDir and HomeDir.
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
			Agents: []types.AgentId{"builder"},
			Skills: []types.SkillId{"implement"},
		},
	}
	return ctx, target, home
}

// TestAdapter_ScopeParity asserts that OpenCode writes its primary tree under
// the scope-correct root. For project/workspace the .opencode/ dir under
// target; for global the ~/.config/opencode dir under home.
func TestAdapter_ScopeParity(t *testing.T) {
	type caseRow struct {
		name    string
		adapter ToolAdapter
		scope   types.SetupScope
		root    func(target, home string) string
	}

	rows := []caseRow{
		{"opencode_project", &OpenCodeAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"opencode_workspace", &OpenCodeAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"opencode_global", &OpenCodeAdapter{}, types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".config", "opencode") }},
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

			for _, rec := range records {
				p := rec.Path
				switch row.scope {
				case types.SetupScopeProject, types.SetupScopeWorkspace:
					if strings.HasPrefix(p, home) {
						t.Errorf("project/workspace scope wrote under home dir: %q", p)
					}
				case types.SetupScopeGlobal:
					if filepath.IsAbs(p) && strings.HasPrefix(p, target) {
						t.Errorf("global scope wrote under target dir: %q", p)
					}
				}
			}
		})
	}
}
