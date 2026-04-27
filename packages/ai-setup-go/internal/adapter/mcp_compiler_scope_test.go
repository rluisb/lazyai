package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// writeCanonicalMcp writes a minimal .ai/mcp.json at targetDir for tests.
func writeCanonicalMcp(t *testing.T, targetDir string) {
	t.Helper()
	aiDir := filepath.Join(targetDir, ".ai")
	if err := files.EnsureDir(aiDir); err != nil {
		t.Fatal(err)
	}
	const mcp = `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem"]}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcp), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestCompileMCPForTool_ScopeParity asserts that each (tool, scope) pair writes
// to the expected on-disk path via ResolveToolRoot — not a project-relative
// fallback. Copilot × global and Claude × global skip cleanly.
func TestCompileMCPForTool_ScopeParity(t *testing.T) {
	type expect struct {
		scope          types.SetupScope
		skipped        bool // if true, expect 0 records and no writes
		writePathUnder func(target, home string) string
	}
	cases := []struct {
		name    string
		tool    types.ToolId
		expects []expect
	}{
		{
			name: "opencode",
			tool: types.ToolIdOpenCode,
			expects: []expect{
				{scope: types.SetupScopeProject, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".opencode", "opencode.jsonc") }},
				{scope: types.SetupScopeWorkspace, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".opencode", "opencode.jsonc") }},
				{scope: types.SetupScopeGlobal, writePathUnder: func(_, h string) string { return filepath.Join(h, ".config", "opencode", "opencode.jsonc") }},
			},
		},
		{
			name: "gemini",
			tool: types.ToolIdGemini,
			expects: []expect{
				{scope: types.SetupScopeProject, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".gemini", "settings.json") }},
				{scope: types.SetupScopeWorkspace, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".gemini", "settings.json") }},
				{scope: types.SetupScopeGlobal, writePathUnder: func(_, h string) string { return filepath.Join(h, ".gemini", "settings.json") }},
			},
		},
		{
			name: "codex",
			tool: types.ToolIdCodex,
			expects: []expect{
				{scope: types.SetupScopeProject, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".codex", "config.toml") }},
				{scope: types.SetupScopeWorkspace, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".codex", "config.toml") }},
				{scope: types.SetupScopeGlobal, writePathUnder: func(_, h string) string { return filepath.Join(h, ".codex", "config.toml") }},
			},
		},
		{
			name: "claude-code",
			tool: types.ToolIdClaudeCode,
			expects: []expect{
				{scope: types.SetupScopeProject, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".mcp.json") }},
				{scope: types.SetupScopeWorkspace, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".mcp.json") }},
				{scope: types.SetupScopeGlobal, skipped: true},
			},
		},
		{
			name: "copilot",
			tool: types.ToolIdCopilot,
			expects: []expect{
				{scope: types.SetupScopeProject, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".vscode", "mcp.json") }},
				{scope: types.SetupScopeWorkspace, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".vscode", "mcp.json") }},
				{scope: types.SetupScopeGlobal, skipped: true},
			},
		},
	}

	for _, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			for _, exp := range tc.expects {
				exp := exp
				t.Run(string(exp.scope), func(t *testing.T) {
					target := t.TempDir()
					home := t.TempDir()
					if tc.tool == types.ToolIdCopilot && exp.scope == types.SetupScopeGlobal {
						t.Setenv("PATH", t.TempDir())
					}
					writeCanonicalMcp(t, target)

					records, err := CompileMCPForTool(tc.tool, CompileContext{
						TargetDir:  target,
						HomeDir:    home,
						SetupScope: exp.scope,
					})
					if err != nil {
						t.Fatalf("CompileMCPForTool: %v", err)
					}

					if exp.skipped {
						if len(records) != 0 {
							t.Errorf("expected 0 records for %s × %s (skipped), got %d", tc.tool, exp.scope, len(records))
						}
						return
					}

					wantPath := exp.writePathUnder(target, home)
					if !files.FileExists(wantPath) {
						t.Errorf("expected file at %q, not found", wantPath)
					}

					// Leak check: project/workspace must not write under home; global must not write under target.
					for _, rec := range records {
						p := rec.Path
						if filepath.IsAbs(p) {
							switch exp.scope {
							case types.SetupScopeProject, types.SetupScopeWorkspace:
								if strings.HasPrefix(p, home) {
									t.Errorf("%s × %s wrote under home: %q", tc.tool, exp.scope, p)
								}
							case types.SetupScopeGlobal:
								if strings.HasPrefix(p, target) {
									t.Errorf("%s × %s wrote under target: %q", tc.tool, exp.scope, p)
								}
							}
						}
					}
				})
			}
		})
	}
}
