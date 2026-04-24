package adapter

import (
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestIsScopeSupported(t *testing.T) {
	tests := []struct {
		tool  types.ToolId
		scope types.SetupScope
		want  bool
	}{
		{types.ToolIdOpenCode, types.SetupScopeProject, true},
		{types.ToolIdOpenCode, types.SetupScopeGlobal, true},
		{types.ToolIdOpenCode, types.SetupScopeWorkspace, true},
		{types.ToolId("claude-code"), types.SetupScopeProject, false},
		{types.ToolId("gemini"), types.SetupScopeProject, false},
		{types.ToolId("copilot"), types.SetupScopeProject, false},
		{types.ToolId("codex"), types.SetupScopeProject, false},
	}
	for _, tc := range tests {
		got := IsScopeSupported(tc.tool, tc.scope)
		if got != tc.want {
			t.Errorf("IsScopeSupported(%s, %s) = %v, want %v", tc.tool, tc.scope, got, tc.want)
		}
	}
}

func TestResolveToolRoot_AllPairs(t *testing.T) {
	target := t.TempDir()
	home := t.TempDir()
	ctx := &AdapterContext{TargetDir: target, HomeDir: home}

	cases := []struct {
		tool  types.ToolId
		scope types.SetupScope
		want  string
	}{
		{types.ToolIdOpenCode, types.SetupScopeProject, filepath.Join(target, ".opencode")},
		{types.ToolIdOpenCode, types.SetupScopeWorkspace, filepath.Join(target, ".opencode")},
		{types.ToolIdOpenCode, types.SetupScopeGlobal, filepath.Join(home, ".config", "opencode")},
	}

	for _, c := range cases {
		got, err := ResolveToolRoot(c.tool, c.scope, ctx)
		if err != nil {
			t.Errorf("ResolveToolRoot(%s, %s): unexpected err: %v", c.tool, c.scope, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveToolRoot(%s, %s) = %q, want %q", c.tool, c.scope, got, c.want)
		}
	}
}

func TestResolveToolRoot_NilCtx(t *testing.T) {
	_, err := ResolveToolRoot(types.ToolIdOpenCode, types.SetupScopeProject, nil)
	if err == nil {
		t.Fatal("expected error on nil ctx")
	}
}
