package wizard

import (
	"reflect"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/preset"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestRunPhase2NonInteractiveDefaults(t *testing.T) {
	t.Parallel()

	scopes := []types.SetupScope{
		types.SetupScopeGlobal,
		types.SetupScopeProject,
		types.SetupScopeWorkspace,
	}
	presets := []types.PresetLevel{
		types.PresetLevelMinimal,
		types.PresetLevelStandard,
		types.PresetLevelFull,
		types.PresetLevelCustom,
	}

	for _, scope := range scopes {
		scope := scope
		for _, presetLevel := range presets {
			presetLevel := presetLevel
			t.Run(string(scope)+"/"+string(presetLevel), func(t *testing.T) {
				t.Parallel()

				result, action, err := RunPhase2(scope, &Phase2Result{Preset: presetLevel}, true)
				if err != nil {
					t.Fatalf("RunPhase2: %v", err)
				}
				if action != PhaseContinue {
					t.Fatalf("action = %v, want %v", action, PhaseContinue)
				}
				if result.Preset != presetLevel {
					t.Fatalf("Preset = %q, want %q", result.Preset, presetLevel)
				}

				wantFeatures := preset.ResolvePreset(presetLevel)
				if wantFeatures == nil {
					defaults := types.DefaultFeatureFlags()
					wantFeatures = &defaults
				}
				if !reflect.DeepEqual(result.Features, wantFeatures) {
					t.Fatalf("Features = %#v, want %#v", result.Features, wantFeatures)
				}

				wantGit := types.DefaultGitConventions()
				if !reflect.DeepEqual(result.GitConv, &wantGit) {
					t.Fatalf("GitConv = %#v, want %#v", result.GitConv, &wantGit)
				}
			})
		}
	}
}

func TestBuildPhase2Result(t *testing.T) {
	t.Parallel()

	customFeatures := &types.FeatureFlags{
		QualityGates:       true,
		RPIWorkflow:        true,
		ChainOfThought:     false,
		BugResolution:      true,
		ContextEngineering: false,
		TreeOfThoughts:     true,
		ADREnforcement:     false,
		AgentHarness:       true,
		PivotHandling:      false,
	}

	result := buildPhase2Result(
		types.SetupScopeProject,
		types.PresetLevelCustom,
		customFeatures,
		"branch/{description}",
		"[{ticket}] {description}",
		true,
	)

	if result.Preset != types.PresetLevelCustom {
		t.Fatalf("Preset = %q, want %q", result.Preset, types.PresetLevelCustom)
	}
	if !reflect.DeepEqual(result.Features, customFeatures) {
		t.Fatalf("Features = %#v, want %#v", result.Features, customFeatures)
	}
	if got, want := result.GitConv.BranchPattern, "branch/{description}"; got != want {
		t.Fatalf("BranchPattern = %q, want %q", got, want)
	}
	if got, want := result.GitConv.CommitPattern, "[{ticket}] {description}"; got != want {
		t.Fatalf("CommitPattern = %q, want %q", got, want)
	}
	if !result.GitConv.RequireTicket {
		t.Fatalf("RequireTicket = false, want true")
	}

	customFeatures.QualityGates = false
	if !result.Features.QualityGates {
		t.Fatalf("result.Features was not copied")
	}
}

func TestBuildPhase2ResultGitConventionDefaults(t *testing.T) {
	t.Parallel()

	result := buildPhase2Result(types.SetupScopeGlobal, "", nil, "", "", false)
	wantPreset := preset.DefaultPresetForScope(types.SetupScopeGlobal)
	wantFeatures := preset.ResolvePreset(wantPreset)
	wantGit := types.DefaultGitConventions()

	if result.Preset != wantPreset {
		t.Fatalf("Preset = %q, want %q", result.Preset, wantPreset)
	}
	if !reflect.DeepEqual(result.Features, wantFeatures) {
		t.Fatalf("Features = %#v, want %#v", result.Features, wantFeatures)
	}
	if !reflect.DeepEqual(result.GitConv, &wantGit) {
		t.Fatalf("GitConv = %#v, want %#v", result.GitConv, &wantGit)
	}
}

func TestPreviousPhase2Step(t *testing.T) {
	t.Parallel()

	if got := previousPhase2Step(3, types.PresetLevelStandard); got != 1 {
		t.Fatalf("previousPhase2Step(standard branch) = %d, want 1", got)
	}
	if got := previousPhase2Step(3, types.PresetLevelCustom); got != 2 {
		t.Fatalf("previousPhase2Step(custom branch) = %d, want 2", got)
	}
	if got := previousPhase2Step(5, types.PresetLevelCustom); got != 4 {
		t.Fatalf("previousPhase2Step(custom require ticket) = %d, want 4", got)
	}
}

func TestPhase2StepInfoFor(t *testing.T) {
	t.Parallel()

	defaults := &Phase2Result{
		Preset: types.PresetLevelCustom,
		Features: &types.FeatureFlags{
			QualityGates: true,
			RPIWorkflow:  true,
		},
		GitConv: &types.GitConventions{
			BranchPattern: "{ticket}/{description}",
			CommitPattern: "[{ticket}] {description}",
		},
	}

	info := phase2StepInfoFor(3, types.PresetLevelCustom, defaults)
	if got, want := info.Title(), "Features & Conventions — 3/5: Branch Pattern (previous: {ticket}/{description})"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}

	standardInfo := phase2StepInfoFor(3, types.PresetLevelStandard, defaults)
	if got, want := standardInfo.Title(), "Features & Conventions — 2/4: Branch Pattern (previous: {ticket}/{description})"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}
}

func TestFeatureSelectionFromFlags(t *testing.T) {
	t.Parallel()

	flags := &types.FeatureFlags{
		QualityGates:       true,
		RPIWorkflow:        true,
		ChainOfThought:     true,
		BugResolution:      true,
		ContextEngineering: true,
		PivotHandling:      true,
	}

	got := featureSelectionFromFlags(flags)
	want := []string{"qualityGates", "rpiWorkflow", "chainOfThought", "bugResolution", "contextEngineering", "pivotHandling"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("featureSelectionFromFlags() = %#v, want %#v", got, want)
	}
}
