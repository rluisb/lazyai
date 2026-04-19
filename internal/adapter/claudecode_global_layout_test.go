package adapter

import (
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
	wantAgent := filepath.Join(claudeDir, "agents", "builder.md")
	if !files.FileExists(wantAgent) {
		t.Errorf("expected agent at canonical path %q, missing", wantAgent)
	}

	// Flat-layout file at the legacy path must not be created.
	flatLegacy := filepath.Join(claudeDir, "builder.md")
	if files.FileExists(flatLegacy) {
		t.Errorf("agent written at legacy flat path %q (regression of spec 012 Task 001)", flatLegacy)
	}

	// agents/CLAUDE.md (tool-context for the agents directory) must live next
	// to the agents — not at the personal-conventions path.
	agentsContext := filepath.Join(claudeDir, "agents", "CLAUDE.md")
	if !files.FileExists(agentsContext) {
		t.Errorf("agents/CLAUDE.md context file missing at %q", agentsContext)
	}
}

// TestClaudeCode_GlobalLegacyAgentsMigrated verifies that a pre-existing
// flat-layout agent file (from a pre-spec-012 install) is migrated into the
// canonical ~/.claude/agents/ directory by Install.
func TestClaudeCode_GlobalLegacyAgentsMigrated(t *testing.T) {
	ctx, _, home := newScopeTestContext(t, types.SetupScopeGlobal)

	// Pre-seed the legacy flat layout: ~/.claude/builder.md exists with
	// arbitrary user content (simulating the buggy pre-spec-012 install).
	claudeDir := filepath.Join(home, ".claude")
	if err := files.EnsureDir(claudeDir); err != nil {
		t.Fatal(err)
	}
	legacyContent := []byte("---\nname: Builder (legacy)\n---\nHand-edited by user\n")
	legacyPath := filepath.Join(claudeDir, "builder.md")
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
	canonical := filepath.Join(claudeDir, "agents", "builder.md")
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
