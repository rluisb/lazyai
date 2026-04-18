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
		{types.ToolIdGemini, filepath.Join(home, ".gemini")},
		{types.ToolIdCodex, filepath.Join(home, ".codex")},
		{types.ToolIdCopilot, ""},
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

func TestResolveCodexSkillsGlobalDir(t *testing.T) {
	home := "/tmp/fakehome"
	got := ResolveCodexSkillsGlobalDir(home)
	want := filepath.Join(home, ".agents", "skills")
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestIsGlobalSupportedTool(t *testing.T) {
	cases := []struct {
		tool types.ToolId
		want bool
	}{
		{types.ToolIdClaudeCode, true},
		{types.ToolIdOpenCode, true},
		{types.ToolIdGemini, true},
		{types.ToolIdCodex, true},
		{types.ToolIdCopilot, false},
	}
	for _, c := range cases {
		got := IsGlobalSupportedTool(c.tool)
		if got != c.want {
			t.Errorf("IsGlobalSupportedTool(%s) = %v, want %v", c.tool, got, c.want)
		}
	}
}
