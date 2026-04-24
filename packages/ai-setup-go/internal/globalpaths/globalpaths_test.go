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

func TestResolveGlobalToolTargetDir_UnknownTool(t *testing.T) {
	home := "/tmp/fakehome"
	got, err := ResolveGlobalToolTargetDir("unknown-tool", home)
	if err != nil {
		t.Errorf("ResolveGlobalToolTargetDir(unknown): unexpected error: %v", err)
	}
	if got != "" {
		t.Errorf("ResolveGlobalToolTargetDir(unknown) = %q, want empty", got)
	}
}

func TestIsGlobalSupportedTool(t *testing.T) {
	cases := []struct {
		tool types.ToolId
		want bool
	}{
		{types.ToolIdOpenCode, true},
		{types.ToolId("claude-code"), false},
		{types.ToolId("gemini"), false},
		{types.ToolId("codex"), false},
		{types.ToolId("copilot"), false},
	}
	for _, c := range cases {
		got := IsGlobalSupportedTool(c.tool)
		if got != c.want {
			t.Errorf("IsGlobalSupportedTool(%s) = %v, want %v", c.tool, got, c.want)
		}
	}
}
