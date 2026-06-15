package cmd

import (
	"reflect"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
	"github.com/rluisb/lazyai/packages/cli/tui/wizard"
)

func TestBuildScaffoldContextHonorsExplicitPhase1Selections(t *testing.T) {
	t.Parallel()
	ensureTestLibraryFS(t)

	result := &wizard.WizardResult{
		Phase1: &wizard.Phase1Result{
			Scope:         types.SetupScopeProject,
			Tools:         []types.ToolId{types.ToolIdOpenCode},
			Skills:        []types.SkillId{types.SkillIdDiagnose, types.SkillIdPrReview},
			Agents:        []types.AgentId{types.AgentIdBuilder},
			EnableServers: []string{"filesystem", "memory"},
			ProjectName:   "demo-app",
		},
		Phase2: &wizard.Phase2Result{
			Preset: types.PresetLevelMinimal,
		},
	}

	config := &wizard.WizardConfig{TargetDir: t.TempDir(), HomeDir: t.TempDir()}
	ctx, err := buildScaffoldContext(result, config)
	if err != nil {
		t.Fatalf("buildScaffoldContext: %v", err)
	}

	if want := []types.AgentId{types.AgentIdBuilder}; !reflect.DeepEqual(ctx.Agents, want) {
		t.Fatalf("Agents = %#v, want %#v", ctx.Agents, want)
	}
	if want := []types.SkillId{types.SkillIdDiagnose, types.SkillIdPrReview}; !reflect.DeepEqual(ctx.Skills, want) {
		t.Fatalf("Skills = %#v, want %#v", ctx.Skills, want)
	}
	if want := []string{"filesystem", "memory"}; !reflect.DeepEqual(ctx.EnableServers, want) {
		t.Fatalf("EnableServers = %#v, want %#v", ctx.EnableServers, want)
	}
}
