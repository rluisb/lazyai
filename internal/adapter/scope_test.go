package adapter

import (
	"errors"
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
		{types.ToolIdClaudeCode, types.SetupScopeProject, true},
		{types.ToolIdClaudeCode, types.SetupScopeGlobal, true},
		{types.ToolIdClaudeCode, types.SetupScopeWorkspace, true},
		{types.ToolIdOpenCode, types.SetupScopeProject, true},
		{types.ToolIdOpenCode, types.SetupScopeGlobal, true},
		{types.ToolIdOpenCode, types.SetupScopeWorkspace, true},
		{types.ToolIdGemini, types.SetupScopeProject, true},
		{types.ToolIdGemini, types.SetupScopeGlobal, true},
		{types.ToolIdGemini, types.SetupScopeWorkspace, true},
		{types.ToolIdCodex, types.SetupScopeProject, true},
		{types.ToolIdCodex, types.SetupScopeGlobal, true},
		{types.ToolIdCodex, types.SetupScopeWorkspace, true},
		{types.ToolIdCopilot, types.SetupScopeProject, true},
		{types.ToolIdCopilot, types.SetupScopeGlobal, true}, // now supported with probe gating
		{types.ToolIdCopilot, types.SetupScopeWorkspace, true},
		{types.ToolIdPi, types.SetupScopeProject, true},
		{types.ToolIdPi, types.SetupScopeWorkspace, true},
		{types.ToolIdPi, types.SetupScopeGlobal, false},
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
		// gemini
		{types.ToolIdGemini, types.SetupScopeProject, want{filepath.Join(target, ".gemini"), false}},
		{types.ToolIdGemini, types.SetupScopeWorkspace, want{filepath.Join(target, ".gemini"), false}},
		{types.ToolIdGemini, types.SetupScopeGlobal, want{filepath.Join(home, ".gemini"), false}},
		// codex (single-root view)
		{types.ToolIdCodex, types.SetupScopeProject, want{filepath.Join(target, ".codex"), false}},
		{types.ToolIdCodex, types.SetupScopeWorkspace, want{filepath.Join(target, ".codex"), false}},
		{types.ToolIdCodex, types.SetupScopeGlobal, want{filepath.Join(home, ".codex"), false}},
		// copilot
		{types.ToolIdCopilot, types.SetupScopeProject, want{filepath.Join(target, ".github"), false}},
		{types.ToolIdCopilot, types.SetupScopeWorkspace, want{filepath.Join(target, ".github"), false}},
		{types.ToolIdCopilot, types.SetupScopeGlobal, want{filepath.Join(home, ".copilot"), false}}, // now supported
		// pi
		{types.ToolIdPi, types.SetupScopeProject, want{filepath.Join(target, ".pi"), false}},
		{types.ToolIdPi, types.SetupScopeWorkspace, want{filepath.Join(target, ".pi"), false}},
		{types.ToolIdPi, types.SetupScopeGlobal, want{"", true}},

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

func TestResolveCodexRoots(t *testing.T) {
	target := t.TempDir()
	home := t.TempDir()
	ctx := &AdapterContext{TargetDir: target, HomeDir: home}

	type pair struct {
		configRoot string
		skillsRoot string
	}
	cases := []struct {
		scope types.SetupScope
		want  pair
	}{
		{types.SetupScopeProject, pair{filepath.Join(target, ".codex"), filepath.Join(target, ".agents", "skills")}},
		{types.SetupScopeWorkspace, pair{filepath.Join(target, ".codex"), filepath.Join(target, ".agents", "skills")}},
		{types.SetupScopeGlobal, pair{filepath.Join(home, ".codex"), filepath.Join(home, ".agents", "skills")}},
	}

	for _, c := range cases {
		cfg, sk, err := ResolveCodexRoots(c.scope, ctx)
		if err != nil {
			t.Errorf("ResolveCodexRoots(%s): err=%v", c.scope, err)
			continue
		}
		if cfg != c.want.configRoot || sk != c.want.skillsRoot {
			t.Errorf("ResolveCodexRoots(%s) = (%q,%q), want (%q,%q)", c.scope, cfg, sk, c.want.configRoot, c.want.skillsRoot)
		}
	}
}

func TestResolveToolRoot_NilCtx(t *testing.T) {
	_, err := ResolveToolRoot(types.ToolIdClaudeCode, types.SetupScopeProject, nil)
	if err == nil {
		t.Fatal("expected error on nil ctx")
	}
}
