package globalpaths

import (
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestResolveGlobalToolTargetDir(t *testing.T) {
	home := t.TempDir()
	// Ensure PI_CODING_AGENT_DIR never leaks into the fallback test.
	t.Setenv("PI_CODING_AGENT_DIR", "")
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

func TestResolveGlobalToolTargetDir_OmpEnvOverride(t *testing.T) {
	home := t.TempDir()
	custom := filepath.Join(home, "custom-omp-agent")
	t.Setenv("PI_CODING_AGENT_DIR", custom)

	got, err := ResolveGlobalToolTargetDir(types.ToolIdOmp, home)
	if err != nil {
		t.Fatalf("ResolveGlobalToolTargetDir(omp): %v", err)
	}
	if got != custom {
		t.Errorf("ResolveGlobalToolTargetDir(omp) = %q, want %q (env override)", got, custom)
	}
}

func TestResolveGlobalToolTargetDir_OmpFallback(t *testing.T) {
	home := t.TempDir()
	t.Setenv("PI_CODING_AGENT_DIR", "")

	got, err := ResolveGlobalToolTargetDir(types.ToolIdOmp, home)
	if err != nil {
		t.Fatalf("ResolveGlobalToolTargetDir(omp): %v", err)
	}
	want := filepath.Join(home, ".omp", "agent")
	if got != want {
		t.Errorf("ResolveGlobalToolTargetDir(omp) = %q, want %q (fallback)", got, want)
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
