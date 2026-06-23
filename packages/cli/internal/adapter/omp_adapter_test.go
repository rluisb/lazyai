package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestOmpAdapter_Install_AgentsAndSkills(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdResearcher, types.AgentIdReviewer},
		Skills: []types.SkillId{types.SkillIdDiagnose, types.SkillIdIssueTriage},
	}

	adapter := &OmpAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OMP Install failed: %v", err)
	}

	for _, rel := range []string{
		".omp/agents/researcher.md",
		".omp/agents/reviewer.md",
		".omp/skills/diagnose/SKILL.md",
		".omp/skills/issue-triage/SKILL.md",
	} {
		assertExists(t, filepath.Join(targetDir, rel))
	}
}

func TestOmpAdapter_Install_CommandsAndPrompts(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	libDir, err := library.FindLibraryDir()
	if err != nil {
		t.Fatalf("FindLibraryDir: %v", err)
	}
	ctx.LibraryDir = libDir
	ctx.LibraryFS = nil
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdResearcher, types.AgentIdReviewer},
		Skills: []types.SkillId{types.SkillIdDiagnose, types.SkillIdIssueTriage},
	}

	adapter := &OmpAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OMP Install failed: %v", err)
	}

	for _, rel := range []string{
		".omp/commands/graphify.md",
		".omp/commands/handoff.md",
		".omp/prompts/plan.md",
	} {
		assertExists(t, filepath.Join(targetDir, rel))
	}
}

func TestOmpAdapter_Install_Hooks(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	libDir, err := library.FindLibraryDir()
	if err != nil {
		t.Fatalf("FindLibraryDir: %v", err)
	}
	ctx.LibraryDir = libDir
	ctx.LibraryFS = nil
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdResearcher},
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	adapter := &OmpAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OMP Install failed: %v", err)
	}

	hookPath := filepath.Join(targetDir, ".omp", "hooks", "pre", "block-destructive-shell.ts")
	assertExists(t, hookPath)

	data, err := os.ReadFile(hookPath)
	if err != nil {
		t.Fatalf("read hook file: %v", err)
	}
	content := string(data)
	if !strings.Contains(content, "export default") {
		t.Fatalf("hook file missing export default: %q", hookPath)
	}
	if !strings.Contains(content, "tool_call") {
		t.Fatalf("hook file missing tool_call: %q", hookPath)
	}
}

func TestOmpAdapter_GlobalScope_InstallsAgentsAndSkills(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.SetupScope = types.SetupScopeGlobal
	homeDir := t.TempDir()
	ctx.HomeDir = homeDir
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdResearcher},
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	adapter := &OmpAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OMP Install (global) failed: %v", err)
	}

	assertExists(t, filepath.Join(homeDir, ".omp", "agent", "agents", "researcher.md"))
	assertExists(t, filepath.Join(homeDir, ".omp", "agent", "skills", "diagnose", "SKILL.md"))
	if _, err := os.Stat(filepath.Join(targetDir, ".omp")); !os.IsNotExist(err) {
		t.Fatalf("expected no .omp under target dir for global scope")
	}
}

func TestOmpAdapter_CompileMCP_ProjectScope(t *testing.T) {
	targetDir := t.TempDir()
	setupCanonicalMcp(t, targetDir)

	records, err := CompileMCPForTool(types.ToolIdOmp, CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("CompileMCPForTool failed: %v", err)
	}

	configPath := filepath.Join(targetDir, ".omp", "mcp.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected %q: %v", configPath, err)
	}
	if !strings.Contains(string(data), `"mcpServers"`) {
		t.Fatalf("omp mcp config missing mcpServers: %s", string(data))
	}

	if len(records) != 1 {
		t.Fatalf("records len = %d, want 1", len(records))
	}
	if got := records[0].Path; got != configPath {
		t.Fatalf("record path = %q, want %q", got, configPath)
	}
}

// TestOmpAdapter_CompileMCP_AdapterMethod exercises the adapter method directly
// (not CompileMCPForTool) to catch regressions where the adapter returns
// ctx.FileRecords instead of delegating to the compiler.
func TestOmpAdapter_CompileMCP_AdapterMethod(t *testing.T) {
	targetDir := t.TempDir()
	setupCanonicalMcp(t, targetDir)

	adapter := &OmpAdapter{}
	records, err := adapter.CompileMCP(CompileContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("OmpAdapter.CompileMCP failed: %v", err)
	}

	configPath := filepath.Join(targetDir, ".omp", "mcp.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected %q: %v", configPath, err)
	}
	if !strings.Contains(string(data), `"mcpServers"`) {
		t.Fatalf("omp mcp config missing mcpServers: %s", string(data))
	}

	if len(records) != 1 {
		t.Fatalf("records len = %d, want 1", len(records))
	}
	if got := records[0].Path; got != configPath {
		t.Fatalf("record path = %q, want %q", got, configPath)
	}
}

func TestOmpOutputMapping_AgentsEmitted(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdOmp, AssetKindAgents)
	if !ok {
		t.Fatalf("no output target for OMP %q", AssetKindAgents)
	}
	if target.Shape != ShapeFlat {
		t.Fatalf("OMP agent target shape=%q, want %q", target.Shape, ShapeFlat)
	}
}
