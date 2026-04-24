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

// TestCompileMCPForTool_ScopeParity asserts that OpenCode writes to the
// expected on-disk path via scope-aware placement.
// Project/workspace: <targetDir>/opencode.jsonc
// Global:           <home>/.config/opencode/opencode.jsonc
func TestCompileMCPForTool_ScopeParity(t *testing.T) {
	type expect struct {
		scope          types.SetupScope
		writePathUnder func(target, home string) string
	}
	expects := []expect{
		{scope: types.SetupScopeProject, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".opencode", "opencode.jsonc") }},
		{scope: types.SetupScopeWorkspace, writePathUnder: func(t, _ string) string { return filepath.Join(t, ".opencode", "opencode.jsonc") }},
		{scope: types.SetupScopeGlobal, writePathUnder: func(_, h string) string { return filepath.Join(h, ".config", "opencode", "opencode.jsonc") }},
	}

	for _, exp := range expects {
		exp := exp
		t.Run(string(exp.scope), func(t *testing.T) {
			target := t.TempDir()
			home := t.TempDir()
			writeCanonicalMcp(t, target)

			records, err := CompileMCPForTool(types.ToolIdOpenCode, CompileContext{
				TargetDir:  target,
				HomeDir:    home,
				SetupScope: exp.scope,
			})
			if err != nil {
				t.Fatalf("CompileMCPForTool: %v", err)
			}

			wantPath := exp.writePathUnder(target, home)
			if !files.FileExists(wantPath) {
				t.Errorf("expected file at %q, not found", wantPath)
			}

			for _, rec := range records {
				p := rec.Path
				if filepath.IsAbs(p) {
					switch exp.scope {
					case types.SetupScopeProject, types.SetupScopeWorkspace:
						if strings.HasPrefix(p, home) {
							t.Errorf("opencode × %s wrote under home: %q", exp.scope, p)
						}
					case types.SetupScopeGlobal:
						if strings.HasPrefix(p, target) {
							t.Errorf("opencode × %s wrote under target: %q", exp.scope, p)
						}
					}
				}
			}
		})
	}
}
