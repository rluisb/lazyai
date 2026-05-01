package globalpaths

import (
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestResolveGlobalToolTargetDir(t *testing.T) {
	home := "/tmp/fakehome"
	cases := []struct {
		tool types.ToolId
		want string
	}{
		{types.ToolIdOpenCode, filepath.Join(home, ".config", "opencode")},
		{types.ToolIdClaudeCode, filepath.Join(home, ".claude")},
		{types.ToolIdCopilot, filepath.Join(home, ".copilot")},
	}
	for _, c := range cases {
		got, err := ResolveGlobalToolTargetDir(c.tool, home)
		if err != nil {
			t.Errorf("ResolveGlobalToolTargetDir(%s): %v", c.tool, err)
			continue
		}
		if got != c.want {
			t.Errorf("ResolveGlobalToolTargetDir(%s) = %q, want %q", c.tool, got, c.want)
		}
	}
}

func TestIsGlobalSupportedTool(t *testing.T) {
	cases := []struct {
		tool types.ToolId
		want bool
	}{
		{types.ToolIdClaudeCode, true},
		{types.ToolIdOpenCode, true},
		{types.ToolIdCopilot, true}, // now supported with probe gating
		{types.ToolId("gemini"), false},
		{types.ToolId("codex"), false},
		{types.ToolId("pi"), false},
	}
	for _, c := range cases {
		got := IsGlobalSupportedTool(c.tool)
		if got != c.want {
			t.Errorf("IsGlobalSupportedTool(%s) = %v, want %v", c.tool, got, c.want)
		}
	}
}
