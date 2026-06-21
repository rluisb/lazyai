package adapter

import (
	"os"
	"path/filepath"
	"testing"

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

func TestOmpOutputMapping_AgentsEmitted(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdOmp, AssetKindAgents)
	if !ok {
		t.Fatalf("no output target for OMP %q", AssetKindAgents)
	}
	if target.Shape != ShapeFlat {
		t.Fatalf("OMP agent target shape=%q, want %q", target.Shape, ShapeFlat)
	}
}
