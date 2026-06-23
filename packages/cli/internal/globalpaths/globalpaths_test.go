package globalpaths

import (
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
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
		{types.ToolIdOmp, filepath.Join(home, ".omp", "agent")},
		{types.ToolIdKiro, filepath.Join(home, ".kiro")},
		{types.ToolIdPi, filepath.Join(home, ".pi")},
		{types.ToolIdAntigravity, filepath.Join(home, ".gemini")},
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
		{types.ToolIdCopilot, true},
		{types.ToolIdOmp, true},
		{types.ToolIdKiro, true},
		{types.ToolIdPi, true},
		{types.ToolIdAntigravity, true},
		{types.ToolId("gemini"), false},
		{types.ToolId("codex"), false},
	}
	for _, c := range cases {
		got := IsGlobalSupportedTool(c.tool)
		if got != c.want {
			t.Errorf("IsGlobalSupportedTool(%s) = %v, want %v", c.tool, got, c.want)
		}
	}
}
