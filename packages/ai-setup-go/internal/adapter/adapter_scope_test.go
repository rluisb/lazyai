package adapter

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// newScopeTestContext returns an AdapterContext rooted at two separate temp
// dirs for TargetDir and HomeDir. Tests use this to assert paths land under
// the correct root per scope (R-3 mitigation: never read os.UserHomeDir()).
func newScopeTestContext(t *testing.T, scope types.SetupScope) (*AdapterContext, string, string) {
	t.Helper()
	target := t.TempDir()
	home := t.TempDir()
	ctx := &AdapterContext{
		TargetDir:  target,
		HomeDir:    home,
		SetupScope: scope,
		LibraryFS:  createTestFS(),
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Agents: []types.AgentId{"builder"},
			Skills: []types.SkillId{"implement"},
		},
	}
	return ctx, target, home
}

// TestAdapter_ScopeParity asserts that each adapter writes its primary tree
// under the scope-correct root (research §2). It does not exhaustively list
// every emitted file — just the scope-defining directories — so that future
// content changes don't churn this test.
func TestAdapter_ScopeParity(t *testing.T) {
	type expect struct {
		mustExistUnder string // directory that must exist after Install
		mustNotContain string // substring that must not appear in any created path (leak check)
	}
	type caseRow struct {
		name    string
		adapter ToolAdapter
		scope   types.SetupScope
		root    func(target, home string) string
	}

	rows := []caseRow{
		{"claude_project", &ClaudeCodeAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"claude_workspace", &ClaudeCodeAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".claude") }},
		{"claude_global", &ClaudeCodeAdapter{}, types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".claude") }},

		{"opencode_project", &OpenCodeAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"opencode_workspace", &OpenCodeAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".opencode") }},
		{"opencode_global", &OpenCodeAdapter{}, types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".config", "opencode") }},

		{"gemini_project", &GeminiAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".gemini") }},
		{"gemini_workspace", &GeminiAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".gemini") }},
		{"gemini_global", &GeminiAdapter{}, types.SetupScopeGlobal, func(_, h string) string { return filepath.Join(h, ".gemini") }},

		{"copilot_project", &CopilotAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".github") }},
		{"copilot_workspace", &CopilotAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".github") }},
		{"pi_project", &PiAdapter{}, types.SetupScopeProject, func(t, _ string) string { return filepath.Join(t, ".pi") }},
		{"pi_workspace", &PiAdapter{}, types.SetupScopeWorkspace, func(t, _ string) string { return filepath.Join(t, ".pi") }},
		// copilot_global is exercised separately below.
	}

	for _, row := range rows {
		row := row
		t.Run(row.name, func(t *testing.T) {
			ctx, target, home := newScopeTestContext(t, row.scope)
			records, err := row.adapter.Install(ctx)
			if err != nil {
				t.Fatalf("Install: %v", err)
			}
			if len(records) == 0 {
				t.Fatal("no file records produced")
			}
			wantRoot := row.root(target, home)
			if !files.DirExists(wantRoot) {
				t.Errorf("expected directory %q to exist after Install", wantRoot)
			}

			// Leak check: when scope is project/workspace, no path should
			// contain the home dir; when scope is global, no path should
			// contain the target dir.
			for _, rec := range records {
				p := rec.Path
				switch row.scope {
				case types.SetupScopeProject, types.SetupScopeWorkspace:
					if strings.HasPrefix(p, home) {
						t.Errorf("project/workspace scope wrote under home dir: %q", p)
					}
				case types.SetupScopeGlobal:
					// records may be stored as relative paths rooted at
					// target (adapter writes under home for global). Assert
					// the absolute path on disk lives under home.
					if filepath.IsAbs(p) && strings.HasPrefix(p, target) {
						t.Errorf("global scope wrote under target dir: %q", p)
					}
				}
			}
		})
	}
}

// TestCopilotAdapter_GlobalScope_Skips verifies the adapter early-returns
// (no error, no records, no writes) when scope=global and probes fail.
func TestCopilotAdapter_GlobalScope_Skips(t *testing.T) {
	ctx, target, home := newScopeTestContext(t, types.SetupScopeGlobal)
	t.Setenv("PATH", t.TempDir())
	adapter := &CopilotAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("Install at scope=global must not error: %v", err)
	}
	if len(records) != 0 {
		t.Errorf("expected 0 records at scope=global, got %d", len(records))
	}
	if files.DirExists(filepath.Join(home, ".github")) {
		t.Error("copilot must not create ~/.github at scope=global")
	}
	if files.DirExists(filepath.Join(target, ".github")) {
		t.Error("copilot must not create <target>/.github at scope=global")
	}
}

// TestCopilotAdapter_GlobalScope_Emits verifies the adapter correctly emits
// agents, instructions, and chatmodes under ~/.copilot/ at global scope.
func TestCopilotAdapter_GlobalScope_Emits(t *testing.T) {
	ctx, _, home := newScopeTestContext(t, types.SetupScopeGlobal)
	// Create ~/.copilot/ so the probe passes
	copilotHome := filepath.Join(home, ".copilot")
	if err := files.EnsureDir(copilotHome); err != nil {
		t.Fatalf("EnsureDir: %v", err)
	}
	adapter := &CopilotAdapter{}
	records, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("Install at scope=global: %v", err)
	}
	if len(records) == 0 {
		t.Errorf("expected records at scope=global with ~/.copilot/, got 0")
	}
	// Verify agents directory was created
	if !files.DirExists(filepath.Join(copilotHome, "agents")) {
		t.Error("agents directory not created under ~/.copilot/")
	}
	// Verify instructions directory was created
	if !files.DirExists(filepath.Join(copilotHome, "instructions")) {
		t.Error("instructions directory not created under ~/.copilot/")
	}
}

// TestCodexAdapter_WritesConfigAndSplitSkills verifies Codex now emits both
// config.toml under configRoot and skills under skillsRoot, at every scope.
func TestCodexAdapter_WritesConfigAndSplitSkills(t *testing.T) {
	cases := []struct {
		name  string
		scope types.SetupScope
	}{
		{"project", types.SetupScopeProject},
		{"workspace", types.SetupScopeWorkspace},
		{"global", types.SetupScopeGlobal},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			ctx, target, home := newScopeTestContext(t, c.scope)
			adapter := &CodexAdapter{}
			if _, err := adapter.Install(ctx); err != nil {
				t.Fatalf("Install: %v", err)
			}
			cfg, skills, err := ResolveCodexRoots(c.scope, ctx)
			if err != nil {
				t.Fatalf("ResolveCodexRoots: %v", err)
			}
			if !files.FileExists(filepath.Join(cfg, "config.toml")) {
				t.Errorf("config.toml missing at %q", cfg)
			}
			if !files.DirExists(skills) {
				t.Errorf("skills dir missing at %q", skills)
			}
			// Regression: codex must never write under filepath.Dir(target)
			// (the old buggy behaviour — parent of the project dir).
			badGuess := filepath.Join(filepath.Dir(target), ".agents")
			if files.DirExists(badGuess) && c.scope != types.SetupScopeGlobal {
				t.Errorf("codex wrote under parent-of-target: %q", badGuess)
			}
			// Global scope must also emit AGENTS.override.md.
			if c.scope == types.SetupScopeGlobal {
				if !files.FileExists(filepath.Join(cfg, "AGENTS.override.md")) {
					t.Errorf("AGENTS.override.md missing at global configRoot %q", cfg)
				}
			}
			_ = home
		})
	}
}

// TestCodexAdapter_CompileMCP_WritesServers verifies that CompileMCP reads the
// canonical .ai/mcp.json and merges the enabled servers into .codex/config.toml.
func TestCodexAdapter_CompileMCP_WritesServers(t *testing.T) {
	dir := t.TempDir()

	// Write a minimal .ai/mcp.json with one enabled server.
	aiDir := filepath.Join(dir, ".ai")
	if err := files.EnsureDir(aiDir); err != nil {
		t.Fatal(err)
	}
	mcpJSON := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem"],"enabled":true}}}`
	if err := files.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	adapter := &CodexAdapter{}
	records, err := adapter.CompileMCP(CompileContext{
		TargetDir:  dir,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("CompileMCP: %v", err)
	}

	configPath := filepath.Join(dir, ".codex", "config.toml")
	if !files.FileExists(configPath) {
		t.Fatalf("config.toml not created at %q", configPath)
	}

	data, _ := files.ReadFile(configPath)
	content := string(data)
	if !strings.Contains(content, "mcp_servers") {
		t.Errorf("mcp_servers table missing from config.toml:\n%s", content)
	}
	if !strings.Contains(content, "filesystem") {
		t.Errorf("filesystem server missing from config.toml:\n%s", content)
	}
	if !strings.Contains(content, "npx") {
		t.Errorf("command 'npx' missing from config.toml:\n%s", content)
	}

	if len(records) == 0 {
		t.Error("expected at least one TrackedFile record")
	}
}

// TestCodexAdapter_ConfigMergePreservesUserKeys verifies that running Install
// against a pre-existing config.toml with user-authored tables preserves them
// and creates a .bak sidecar.
func TestCodexAdapter_ConfigMergePreservesUserKeys(t *testing.T) {
	ctx, _, _ := newScopeTestContext(t, types.SetupScopeProject)

	configRoot, _, err := ResolveCodexRoots(types.SetupScopeProject, ctx)
	if err != nil {
		t.Fatalf("ResolveCodexRoots: %v", err)
	}
	if err := files.EnsureDir(configRoot); err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(configRoot, "config.toml")
	userTOML := "[profile]\nname = \"me\"\n"
	if err := files.WriteFile(configPath, []byte(userTOML), 0o644); err != nil {
		t.Fatal(err)
	}

	if _, err := (&CodexAdapter{}).Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	if !files.FileExists(configPath + ".bak") {
		t.Error("config.toml.bak not created on first touch")
	}
	data, err := files.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	content := string(data)
	if !strings.Contains(content, "[profile]") {
		t.Errorf("user [profile] lost after merge:\n%s", content)
	}
	if !strings.Contains(content, "[mcp_servers]") {
		t.Errorf("ai-setup [mcp_servers] not merged in:\n%s", content)
	}
}

// TestGeminiAdapter_DriveCLI_FallsBackWhenBinaryAbsent verifies that when
// DriveCLI=true but the gemini binary is not on PATH, Install succeeds by
// falling back to direct-write (settings.json is still created).
func TestGeminiAdapter_DriveCLI_FallsBackWhenBinaryAbsent(t *testing.T) {
	ctx, _, _ := newScopeTestContext(t, types.SetupScopeProject)
	ctx.DriveCLI = true

	// Guarantee gemini binary is not found by prepending an empty tmpdir to PATH.
	emptyDir := t.TempDir()
	origPATH := os.Getenv("PATH")
	t.Setenv("PATH", emptyDir+string(os.PathListSeparator)+origPATH)

	adapter := &GeminiAdapter{}
	_, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("Install with DriveCLI=true and absent binary must not error: %v", err)
	}

	geminiDir := filepath.Join(ctx.TargetDir, ".gemini")
	settingsPath := filepath.Join(geminiDir, "settings.json")
	if !files.FileExists(settingsPath) {
		t.Errorf("settings.json not created via direct-write fallback")
	}
}

// TestGeminiAdapter_DriveCLI_CallsGeminiBinary verifies that when DriveCLI=true
// and a stub gemini binary is on PATH, the adapter invokes `gemini mcp add`.
func TestGeminiAdapter_DriveCLI_CallsGeminiBinary(t *testing.T) {
	ctx, _, _ := newScopeTestContext(t, types.SetupScopeProject)
	ctx.DriveCLI = true

	// Write a stub gemini binary that records the args it receives.
	stubDir := t.TempDir()
	recordFile := filepath.Join(stubDir, "gemini-args.txt")
	stubScript := fmt.Sprintf("#!/bin/sh\necho \"$@\" >> %s\n", recordFile)
	stubPath := filepath.Join(stubDir, "gemini")
	if err := os.WriteFile(stubPath, []byte(stubScript), 0o755); err != nil {
		t.Fatal(err)
	}

	// Write a canonical .ai/mcp.json so the adapter has a server to register.
	aiDir := filepath.Join(ctx.TargetDir, ".ai")
	if err := files.EnsureDir(aiDir); err != nil {
		t.Fatal(err)
	}
	mcpJSON := `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem"]}}}`
	if err := files.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcpJSON), 0o644); err != nil {
		t.Fatal(err)
	}

	origPATH := os.Getenv("PATH")
	t.Setenv("PATH", stubDir+string(os.PathListSeparator)+origPATH)

	adapter := &GeminiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	// Stub should have been called — record file will exist if gemini was invoked.
	if !files.FileExists(recordFile) {
		t.Error("stub gemini binary was not called when DriveCLI=true and binary is present")
	}
	data, _ := files.ReadFile(recordFile)
	if !strings.Contains(string(data), "mcp add") {
		t.Errorf("expected 'mcp add' in stub args, got: %s", string(data))
	}
}
