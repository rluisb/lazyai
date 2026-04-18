package scaffold

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestMemoryDocDestPath(t *testing.T) {
	target := "/tmp/target"
	home := "/tmp/home"

	cases := []struct {
		name  string
		tool  types.ToolId
		scope types.SetupScope
		want  string
		unsup bool
	}{
		// claude-code
		{"claude_project", types.ToolIdClaudeCode, types.SetupScopeProject, filepath.Join(target, "CLAUDE.md"), false},
		{"claude_workspace", types.ToolIdClaudeCode, types.SetupScopeWorkspace, filepath.Join(target, "CLAUDE.md"), false},
		{"claude_global", types.ToolIdClaudeCode, types.SetupScopeGlobal, filepath.Join(home, ".claude", "CLAUDE.md"), false},
		// opencode
		{"opencode_project", types.ToolIdOpenCode, types.SetupScopeProject, filepath.Join(target, "AGENTS.md"), false},
		{"opencode_global", types.ToolIdOpenCode, types.SetupScopeGlobal, filepath.Join(home, ".config", "opencode", "AGENTS.md"), false},
		// gemini
		{"gemini_project", types.ToolIdGemini, types.SetupScopeProject, filepath.Join(target, "GEMINI.md"), false},
		{"gemini_global", types.ToolIdGemini, types.SetupScopeGlobal, filepath.Join(home, ".gemini", "GEMINI.md"), false},
		// codex
		{"codex_project", types.ToolIdCodex, types.SetupScopeProject, filepath.Join(target, "AGENTS.md"), false},
		{"codex_global", types.ToolIdCodex, types.SetupScopeGlobal, filepath.Join(home, ".codex", "AGENTS.md"), false},
		// copilot
		{"copilot_project", types.ToolIdCopilot, types.SetupScopeProject, filepath.Join(target, ".github", "copilot-instructions.md"), false},
		{"copilot_workspace", types.ToolIdCopilot, types.SetupScopeWorkspace, filepath.Join(target, ".github", "copilot-instructions.md"), false},
		{"copilot_global", types.ToolIdCopilot, types.SetupScopeGlobal, "", true},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			outputFile := RootFileByTool[c.tool]
			got, err := memoryDocDestPath(c.tool, c.scope, target, home, outputFile)
			if c.unsup {
				if !errors.Is(err, errMemoryDocScopeUnsupported) {
					t.Fatalf("want errMemoryDocScopeUnsupported, got err=%v path=%q", err, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected err: %v", err)
			}
			if got != c.want {
				t.Errorf("got %q, want %q", got, c.want)
			}
		})
	}
}
