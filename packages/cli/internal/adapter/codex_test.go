package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestCodexAdapter_Install_EmitsSubagentsSkillsAndHooks(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	libFS, ok := ctx.LibraryFS.(fstest.MapFS)
	if !ok {
		t.Fatalf("expected test library fs")
	}
	libFS["codex/hooks.json"] = &fstest.MapFile{
		Data: []byte(`{"hooks":{"PreToolUse":[{"matcher":"Bash","hooks":[{"type":"command","command":"x"}]}]}}`),
	}
	libFS["codex/hooks/lazyai/block-destructive-shell.sh"] = &fstest.MapFile{
		Data: []byte("#!/usr/bin/env bash\nexit 0\n"),
	}
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdReviewer},
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	a := &CodexAdapter{}
	if _, err := a.Install(ctx); err != nil {
		t.Fatalf("Codex Install failed: %v", err)
	}

	for _, path := range []string{
		filepath.Join(targetDir, ".codex", "agents", "reviewer.toml"),
		filepath.Join(targetDir, ".agents", "skills", "diagnose", "SKILL.md"),
		filepath.Join(targetDir, ".codex", "hooks.json"),
		filepath.Join(targetDir, ".codex", "hooks", "lazyai", "block-destructive-shell.sh"),
	} {
		assertExists(t, path)
	}

	// The default "guide" agent is the session (AGENTS.md), never a subagent.
	assertMissing(t, filepath.Join(targetDir, ".codex", "agents", "guide.toml"))

	data, err := os.ReadFile(filepath.Join(targetDir, ".codex", "agents", "reviewer.toml"))
	if err != nil {
		t.Fatalf("read reviewer subagent: %v", err)
	}
	got := string(data)
	for _, key := range []string{`name = "reviewer"`, "description =", "developer_instructions ="} {
		if !strings.Contains(got, key) {
			t.Fatalf("codex subagent TOML missing %q:\n%s", key, got)
		}
	}
}

func TestCompileCodexMCP_WritesMcpServersToml(t *testing.T) {
	targetDir := t.TempDir()
	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	mcp := `{"servers":{"ctx7":{"command":"npx","args":["-y","@upstash/context7-mcp"]},"figma":{"url":"https://mcp.figma.com/mcp","headers":{"X-Region":"us"}}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(mcp), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	records, err := CompileMCPForTool(types.ToolIdCodex, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool(codex) failed: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 tracked record, got %d", len(records))
	}

	data, err := os.ReadFile(filepath.Join(targetDir, ".codex", "config.toml"))
	if err != nil {
		t.Fatalf("read config.toml: %v", err)
	}
	got := string(data)
	for _, want := range []string{
		"[mcp_servers.ctx7]",
		`command = "npx"`,
		"[mcp_servers.figma]",
		`url = "https://mcp.figma.com/mcp"`,
		"http_headers",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("config.toml missing %q:\n%s", want, got)
		}
	}
}

func TestCompileCodexMCP_PreservesUserConfig(t *testing.T) {
	targetDir := t.TempDir()
	codexDir := filepath.Join(targetDir, ".codex")
	_ = files.EnsureDir(codexDir)
	if err := os.WriteFile(filepath.Join(codexDir, "config.toml"),
		[]byte("model = \"gpt-5.5\"\napproval_policy = \"on-request\"\n"), 0o644); err != nil {
		t.Fatalf("seed config.toml: %v", err)
	}
	aiDir := filepath.Join(targetDir, ".ai")
	_ = files.EnsureDir(aiDir)
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"),
		[]byte(`{"servers":{"ctx7":{"command":"npx"}}}`), 0o644); err != nil {
		t.Fatalf("write mcp.json: %v", err)
	}

	if _, err := CompileMCPForTool(types.ToolIdCodex, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("CompileMCPForTool(codex) failed: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(codexDir, "config.toml"))
	if err != nil {
		t.Fatalf("read config.toml: %v", err)
	}
	got := string(data)
	if !strings.Contains(got, `model = "gpt-5.5"`) {
		t.Fatalf("user key model lost after MCP merge:\n%s", got)
	}
	if !strings.Contains(got, `approval_policy = "on-request"`) {
		t.Fatalf("user key approval_policy lost after MCP merge:\n%s", got)
	}
	if !strings.Contains(got, "[mcp_servers.ctx7]") {
		t.Fatalf("mcp server not merged in:\n%s", got)
	}
}
