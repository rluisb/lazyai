package adapter

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestIsScopeSupported(t *testing.T) {
	tests := []struct {
		tool  types.ToolId
		scope types.SetupScope
		want  bool
	}{
		{types.ToolIdClaudeCode, types.SetupScopeProject, true},
		{types.ToolIdClaudeCode, types.SetupScopeGlobal, true},
		{types.ToolIdClaudeCode, types.SetupScopeWorkspace, true},
		{types.ToolIdOpenCode, types.SetupScopeProject, true},
		{types.ToolIdOpenCode, types.SetupScopeGlobal, true},
		{types.ToolIdOpenCode, types.SetupScopeWorkspace, true},
		{types.ToolIdCopilot, types.SetupScopeProject, true},
		{types.ToolIdCopilot, types.SetupScopeGlobal, true}, // now supported with probe gating
		{types.ToolIdCopilot, types.SetupScopeWorkspace, true},
		{types.ToolId("gemini"), types.SetupScopeProject, false},
		{types.ToolId("codex"), types.SetupScopeGlobal, false},
		{types.ToolId("pi"), types.SetupScopeWorkspace, false},
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

	type want struct {
		path     string
		errIsUns bool
	}
	cases := []struct {
		tool  types.ToolId
		scope types.SetupScope
		want  want
	}{
		// claude-code
		{types.ToolIdClaudeCode, types.SetupScopeProject, want{filepath.Join(target, ".claude"), false}},
		{types.ToolIdClaudeCode, types.SetupScopeWorkspace, want{filepath.Join(target, ".claude"), false}},
		{types.ToolIdClaudeCode, types.SetupScopeGlobal, want{filepath.Join(home, ".claude"), false}},
		// opencode
		{types.ToolIdOpenCode, types.SetupScopeProject, want{filepath.Join(target, ".opencode"), false}},
		{types.ToolIdOpenCode, types.SetupScopeWorkspace, want{filepath.Join(target, ".opencode"), false}},
		{types.ToolIdOpenCode, types.SetupScopeGlobal, want{filepath.Join(home, ".config", "opencode"), false}},
		// copilot
		{types.ToolIdCopilot, types.SetupScopeProject, want{filepath.Join(target, ".github"), false}},
		{types.ToolIdCopilot, types.SetupScopeWorkspace, want{filepath.Join(target, ".github"), false}},
		{types.ToolIdCopilot, types.SetupScopeGlobal, want{filepath.Join(home, ".copilot"), false}}, // now supported
		// removed tools
		{types.ToolId("gemini"), types.SetupScopeProject, want{"", true}},
		{types.ToolId("codex"), types.SetupScopeGlobal, want{"", true}},
		{types.ToolId("pi"), types.SetupScopeWorkspace, want{"", true}},
	}

	for _, c := range cases {
		got, err := ResolveToolRoot(c.tool, c.scope, ctx)
		if c.want.errIsUns {
			if !errors.Is(err, ErrScopeUnsupported) {
				t.Errorf("ResolveToolRoot(%s, %s): want ErrScopeUnsupported, got err=%v path=%q", c.tool, c.scope, err, got)
			}
			continue
		}
		if err != nil {
			t.Errorf("ResolveToolRoot(%s, %s): unexpected err: %v", c.tool, c.scope, err)
			continue
		}
		if got != c.want.path {
			t.Errorf("ResolveToolRoot(%s, %s) = %q, want %q", c.tool, c.scope, got, c.want.path)
		}
	}
}

func TestResolveToolRoot_NilCtx(t *testing.T) {
	_, err := ResolveToolRoot(types.ToolIdClaudeCode, types.SetupScopeProject, nil)
	if err == nil {
		t.Fatal("expected error on nil ctx")
	}
}
