package adapter

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// TestClaudeCode_GlobalAgentsInSubdir is the regression guard for spec 012,
// task 001: at global scope, agent files must live under ~/.claude/agents/,
// not flat at ~/.claude/. Reverting the fix to the pre-spec-012 layout will
// fail this test.
func TestClaudeCode_GlobalAgentsInSubdir(t *testing.T) {
	ctx, _, home := newScopeTestContext(t, types.SetupScopeGlobal)

	if _, err := (&ClaudeCodeAdapter{}).Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	claudeDir := filepath.Join(home, ".claude")
	wantAgent := filepath.Join(claudeDir, "agents", "guide.md")
	if !files.FileExists(wantAgent) {
		t.Errorf("expected agent at canonical path %q, missing", wantAgent)
	}

	// Flat-layout file at the legacy path must not be created.
	flatLegacy := filepath.Join(claudeDir, "guide.md")
	if files.FileExists(flatLegacy) {
		t.Errorf("agent written at legacy flat path %q (regression of spec 012 Task 001)", flatLegacy)
	}

	// Reserved context docs inside the Claude tool directory are handled
	// elsewhere and must not be created by adapter install.
	for _, path := range []string{
		filepath.Join(claudeDir, "CLAUDE.md"),
		filepath.Join(claudeDir, "agents", "CLAUDE.md"),
		filepath.Join(claudeDir, "skills", "CLAUDE.md"),
	} {
		if files.FileExists(path) {
			t.Errorf("reserved context doc should not be created at %q", path)
		}
	}
}

// TestClaudeCode_GlobalLegacyAgentsMigrated verifies that a pre-existing
// flat-layout agent file (from a pre-spec-012 install) is migrated into the
// canonical ~/.claude/agents/ directory by Install.
func TestClaudeCode_GlobalLegacyAgentsMigrated(t *testing.T) {
	ctx, _, home := newScopeTestContext(t, types.SetupScopeGlobal)

	// Pre-seed the legacy flat layout: ~/.claude/implementer.md exists with
	// arbitrary user content (simulating the buggy pre-spec-012 install).
	claudeDir := filepath.Join(home, ".claude")
	if err := files.EnsureDir(claudeDir); err != nil {
		t.Fatal(err)
	}
	legacyContent := []byte("---\nname: Implementer (legacy)\n---\nHand-edited by user\n")
	legacyPath := filepath.Join(claudeDir, "implementer.md")
	if err := files.WriteFile(legacyPath, legacyContent, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := (&ClaudeCodeAdapter{}).Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// After install: legacy path is gone; canonical path holds either the
	// migrated content or the freshly installed library content.
	if files.FileExists(legacyPath) {
		t.Errorf("legacy flat agent %q still present after Install", legacyPath)
	}
	canonical := filepath.Join(claudeDir, "agents", "implementer.md")
	if !files.FileExists(canonical) {
		t.Errorf("canonical agent %q missing after Install (migration failed)", canonical)
	}
}

// TestClaudeCode_GlobalPersonalCLAUDEMDPreserved verifies that a pre-existing
// ~/.claude/CLAUDE.md (the user's personal-conventions file per Claude Code
// docs) is not overwritten by Install at global scope.
func TestClaudeCode_GlobalPersonalCLAUDEMDPreserved(t *testing.T) {
	ctx, _, home := newScopeTestContext(t, types.SetupScopeGlobal)

	claudeDir := filepath.Join(home, ".claude")
	if err := files.EnsureDir(claudeDir); err != nil {
		t.Fatal(err)
	}
	personal := []byte("# My personal Claude conventions\n\n- Always use tabs.\n")
	personalPath := filepath.Join(claudeDir, "CLAUDE.md")
	if err := files.WriteFile(personalPath, personal, 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := (&ClaudeCodeAdapter{}).Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	got, err := files.ReadFile(personalPath)
	if err != nil {
		t.Fatalf("read personal CLAUDE.md: %v", err)
	}
	if string(got) != string(personal) {
		t.Errorf("~/.claude/CLAUDE.md was overwritten\nwant: %q\n got: %q", string(personal), string(got))
	}
}

// TestClaudeCode_SampleRuleSourced verifies that the TypeScript sample rule
// matches the library source byte-for-byte (spec 012 task 003).
func TestClaudeCode_SampleRuleSourced(t *testing.T) {
	ctx, _, home := newScopeTestContext(t, types.SetupScopeGlobal)

	if _, err := (&ClaudeCodeAdapter{}).Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	libContent, err := ReadSampleRuleContent(ctx)
	if err != nil {
		t.Fatalf("read library rule: %v", err)
	}

	installed := filepath.Join(home, ".claude", "rules", "typescript.md")
	installedContent, err := files.ReadFile(installed)
	if err != nil {
		t.Fatalf("read installed rule: %v", err)
	}

	if !bytes.Equal(libContent, installedContent) {
		t.Errorf("installed rule does not match library source\nlibrary:\n%q\n\ninstalled:\n%q",
			string(libContent), string(installedContent))
	}
}

// TestClaudeCode_CommandsAndOutputStylesScopeParity verifies that commands and
// output-styles are copied at every scope (spec 012 task 007).
func TestClaudeCode_CommandsAndOutputStylesScopeParity(t *testing.T) {
	cases := []struct {
		name  string
		scope types.SetupScope
		root  func(target, home string) string
	}{
		{"project", types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"workspace", types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"global", types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".claude") }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, target, home := newScopeTestContext(t, c.scope)

			if _, err := (&ClaudeCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install: %v", err)
			}

			root := c.root(target, home)

			// Check commands directory exists
			commandsDir := filepath.Join(root, "commands")
			if !files.DirExists(commandsDir) {
				t.Errorf("commands directory missing at %q", commandsDir)
			}

			// Check output-styles directory exists
			stylesDir := filepath.Join(root, "output-styles")
			if !files.DirExists(stylesDir) {
				t.Errorf("output-styles directory missing at %q", stylesDir)
			}

			// Check a few sample files exist
			expectedFiles := []string{
				filepath.Join(commandsDir, "review.md"),
				filepath.Join(commandsDir, "test.md"),
				filepath.Join(commandsDir, "commit.md"),
				filepath.Join(stylesDir, "terse.md"),
				filepath.Join(stylesDir, "explanatory.md"),
			}
			for _, f := range expectedFiles {
				if !files.FileExists(f) {
					t.Errorf("expected file missing: %q", f)
				}
			}
		})
	}
}

func TestClaudeCode_DefaultAgentScopeParity(t *testing.T) {
	cases := []struct {
		name  string
		scope types.SetupScope
		root  func(target, home string) string
	}{
		{"project", types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"workspace", types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"global", types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".claude") }},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, target, home := newScopeTestContext(t, c.scope)

			if _, err := (&ClaudeCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install: %v", err)
			}

			root := c.root(target, home)
			defaultAgent := filepath.Join(root, "agents", "guide.md")
			if !files.FileExists(defaultAgent) {
				t.Errorf("expected guide at %q, missing", defaultAgent)
			}
			orch := filepath.Join(root, "agents", "orchestrator.md")
			if files.FileExists(orch) {
				t.Errorf("orchestrator agent present at %q", orch)
			}
		})
	}
}

// TestClaudeCode_HookCommandScopeAware is the regression guard for issue #547:
// at global scope the hook command paths must reference the actual install
// location (~/.claude/hooks/<x>.sh, an absolute path), not the project-relative
// ${CLAUDE_PROJECT_DIR:-$PWD} form. At project scope the env-var form is
// unchanged — the scripts live beside settings.json in .claude/hooks/.
func TestClaudeCode_HookCommandScopeAware(t *testing.T) {
	cases := []struct {
		name       string
		scope      types.SetupScope
		claudeRoot func(target, home string) string
		checkCmd   func(t *testing.T, cmd, claudeRoot string)
	}{
		{
			name:       "global",
			scope:      types.SetupScopeGlobal,
			claudeRoot: func(_, h string) string { return filepath.Join(h, ".claude") },
			checkCmd: func(t *testing.T, cmd, root string) {
				t.Helper()
				wantPrefix := filepath.Join(root, "hooks") + string(filepath.Separator)
				if !strings.HasPrefix(cmd, wantPrefix) {
					t.Errorf("global hook command %q does not start with install dir %q", cmd, wantPrefix)
				}
				if strings.Contains(cmd, "${CLAUDE_PROJECT_DIR") {
					t.Errorf("global hook command %q must not use ${CLAUDE_PROJECT_DIR} (resolves to project, not home)", cmd)
				}
			},
		},
		{
			name:       "project",
			scope:      types.SetupScopeProject,
			claudeRoot: func(target, _ string) string { return filepath.Join(target, ".claude") },
			checkCmd: func(t *testing.T, cmd, _ string) {
				t.Helper()
				const wantPrefix = "${CLAUDE_PROJECT_DIR:-$PWD}/.claude/hooks/"
				if !strings.HasPrefix(cmd, wantPrefix) {
					t.Errorf("project hook command %q does not start with %q", cmd, wantPrefix)
				}
			},
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			ctx, target, home := newScopeTestContext(t, c.scope)

			if _, err := (&ClaudeCodeAdapter{}).Install(ctx); err != nil {
				t.Fatalf("Install: %v", err)
			}

			root := c.claudeRoot(target, home)
			settingsPath := filepath.Join(root, "settings.json")

			raw, err := os.ReadFile(settingsPath)
			if err != nil {
				t.Fatalf("read settings.json: %v", err)
			}

			var settings map[string]any
			if err := json.Unmarshal(raw, &settings); err != nil {
				t.Fatalf("parse settings.json: %v", err)
			}

			hooks, ok := settings["hooks"].(map[string]any)
			if !ok {
				t.Fatal("settings.json missing 'hooks' key")
			}

			// Extract the command from each hook event entry.
			for _, event := range []string{"PreToolUse", "Stop"} {
				entries, ok := hooks[event].([]any)
				if !ok || len(entries) == 0 {
					t.Errorf("hooks[%q] missing or empty", event)
					continue
				}
				group, ok := entries[0].(map[string]any)
				if !ok {
					t.Errorf("hooks[%q][0] not an object", event)
					continue
				}
				innerHooks, ok := group["hooks"].([]any)
				if !ok || len(innerHooks) == 0 {
					t.Errorf("hooks[%q][0].hooks missing or empty", event)
					continue
				}
				h, ok := innerHooks[0].(map[string]any)
				if !ok {
					t.Errorf("hooks[%q][0].hooks[0] not an object", event)
					continue
				}
				cmd, ok := h["command"].(string)
				if !ok {
					t.Errorf("hooks[%q][0].hooks[0].command not a string", event)
					continue
				}
				c.checkCmd(t, cmd, root)
			}
		})
	}
}
