package setup

import (
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestBuildAddScaffoldContextMergesSelectionsWithoutDuplicates(t *testing.T) {
	storeData := types.DefaultStoreData()
	storeData.Config.SetupScope = types.SetupScopeProject
	storeData.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
	storeData.Config.CLITools = []string{string(types.ToolIdOpenCode)}
	storeData.Config.ProjectName = "lazy-app"
	storeData.Config.PlanningDir = "specs"
	storeData.Selections.Agents = []types.AgentId{types.AgentIdBuilder}
	storeData.Selections.Skills = []types.SkillId{types.SkillIdPlan}

	ctx, presetLevel, err := BuildAddScaffoldContext("/work/app", Library{
		Dir: "library",
		FS:  fstest.MapFS{},
	}, &storeData, AddSelections{
		Tools:  []types.ToolId{types.ToolIdOpenCode, types.ToolIdClaudeCode},
		Agents: []string{string(types.AgentIdBuilder), string(types.AgentIdReviewer)},
		Skills: []string{string(types.SkillIdPlan), string(types.SkillIdResearch)},
	})
	if err != nil {
		t.Fatalf("BuildAddScaffoldContext: %v", err)
	}

	if presetLevel != types.PresetLevelStandard {
		t.Fatalf("presetLevel = %q, want %q", presetLevel, types.PresetLevelStandard)
	}
	assertToolIDs(t, ctx.Tools, types.ToolIdOpenCode, types.ToolIdClaudeCode)
	assertStrings(t, ctx.CLITools, string(types.ToolIdOpenCode), string(types.ToolIdClaudeCode))
	assertAgentIDs(t, ctx.Agents, types.AgentIdBuilder, types.AgentIdReviewer)
	assertSkillIDs(t, ctx.Skills, types.SkillIdPlan, types.SkillIdResearch)
	if ctx.Strategy != types.ConflictStrategyAlign {
		t.Fatalf("Strategy = %q, want %q", ctx.Strategy, types.ConflictStrategyAlign)
	}
	if ctx.TargetDir != "/work/app" || ctx.LibraryDir != "library" || ctx.ProjectName != "lazy-app" {
		t.Fatalf("context was not populated from store data: %#v", ctx)
	}
}

func TestBuildUpdateScaffoldContextAppliesForceDryRunAndStoreData(t *testing.T) {
	storeData := types.DefaultStoreData()
	storeData.Config.SetupScope = types.SetupScopeProject
	storeData.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
	storeData.Config.ProjectName = "lazy-app"
	storeData.Selections.Agents = []types.AgentId{types.AgentIdBuilder}

	ctx, presetLevel, err := BuildUpdateScaffoldContext("/work/app", Library{FS: fstest.MapFS{}}, &storeData, UpdateOptions{
		Force:  true,
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("BuildUpdateScaffoldContext: %v", err)
	}

	if presetLevel != types.PresetLevelStandard {
		t.Fatalf("presetLevel = %q, want %q", presetLevel, types.PresetLevelStandard)
	}
	if ctx.Strategy != types.ConflictStrategyBackupAndReplace {
		t.Fatalf("Strategy = %q, want %q", ctx.Strategy, types.ConflictStrategyBackupAndReplace)
	}
	if !ctx.Force || !ctx.DryRun {
		t.Fatalf("Force/DryRun = %v/%v, want true/true", ctx.Force, ctx.DryRun)
	}
	if ctx.StoreData != &storeData {
		t.Fatal("expected update context to retain original store data pointer")
	}
}

func TestBuildScaffoldContextRejectsNilStoreData(t *testing.T) {
	if _, _, err := BuildAddScaffoldContext("/work/app", Library{}, nil, AddSelections{}); err == nil {
		t.Fatal("expected BuildAddScaffoldContext to reject nil store data")
	}
	if _, _, err := BuildUpdateScaffoldContext("/work/app", Library{}, nil, UpdateOptions{}); err == nil {
		t.Fatal("expected BuildUpdateScaffoldContext to reject nil store data")
	}
}

func assertToolIDs(t *testing.T, got []types.ToolId, want ...types.ToolId) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("tool count = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("tool[%d] = %q, want %q (all: %v)", i, got[i], want[i], got)
		}
	}
}

func assertAgentIDs(t *testing.T, got []types.AgentId, want ...types.AgentId) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("agent count = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("agent[%d] = %q, want %q (all: %v)", i, got[i], want[i], got)
		}
	}
}

func assertSkillIDs(t *testing.T, got []types.SkillId, want ...types.SkillId) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("skill count = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("skill[%d] = %q, want %q (all: %v)", i, got[i], want[i], got)
		}
	}
}

func assertStrings(t *testing.T, got []string, want ...string) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("string count = %d, want %d (%v)", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("string[%d] = %q, want %q (all: %v)", i, got[i], want[i], got)
		}
	}
}
