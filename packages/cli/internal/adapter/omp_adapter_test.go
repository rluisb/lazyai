package adapter

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestOmpAdapter_Install_SkillsOnly(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Skills: []types.SkillId{types.SkillIdDiagnose, types.SkillIdIssueTriage},
	}

	adapter := &OmpAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OMP Install failed: %v", err)
	}

	for _, rel := range []string{
		".omp/skills/diagnose/SKILL.md",
		".omp/skills/issue-triage/SKILL.md",
	} {
		assertExists(t, filepath.Join(targetDir, rel))
	}

	assertMissing(t, filepath.Join(targetDir, ".omp", "agents"))
}

func TestOmpAdapter_GlobalScope_InstallsSkillsNoAgentsDir(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.SetupScope = types.SetupScopeGlobal
	homeDir := t.TempDir()
	ctx.HomeDir = homeDir
	ctx.Selections = AdapterSelections{
		Skills: []types.SkillId{types.SkillIdDiagnose},
	}

	adapter := &OmpAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("OMP Install (global) failed: %v", err)
	}

	assertExists(t, filepath.Join(homeDir, ".omp", "agent", "skills", "diagnose", "SKILL.md"))
	assertMissing(t, filepath.Join(homeDir, ".omp", "agent", "agents"))
	if _, err := os.Stat(filepath.Join(targetDir, ".omp")); !os.IsNotExist(err) {
		t.Fatalf("expected no .omp under target dir for global scope")
	}
}

func TestOmpOutputMapping_AgentsNotEmitted(t *testing.T) {
	target, ok := LookupOutputTarget(types.ToolIdOmp, AssetKindAgents)
	if !ok {
		t.Fatalf("no output target for OMP %q", AssetKindAgents)
	}
	if target.Shape != ShapeNone {
		t.Fatalf("OMP agent target shape=%q, want %q", target.Shape, ShapeNone)
	}
}
