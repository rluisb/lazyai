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
		nil,
		nil,
		nil,
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

	result := buildPhase2Result(types.SetupScopeGlobal, "", nil, "", "", false, nil, nil, nil)
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
	if result.CoverageThreshold != 80 {
		t.Fatalf("CoverageThreshold = %d, want 80", result.CoverageThreshold)
	}
}

func TestRunPhase2NonInteractivePreservesProjectProfileDefaults(t *testing.T) {
	t.Parallel()

	defaults := &Phase2Result{
		Preset:            types.PresetLevelStandard,
		ProjectOverview:   "Builds AI setup scaffolds",
		NamingConventions: "Use camelCase",
		ErrorHandling:     "Return wrapped errors",
		APIConventions:    "JSON responses",
		ImportOrder:       "stdlib, third-party, internal",
		ProtectedBranch:   "main",
		TestCommand:       "go test ./...",
		LintCommand:       "go vet ./...",
		BuildCommand:      "go build ./...",
		CoverageThreshold: 95,
	}

	result, action, err := RunPhase2(types.SetupScopeProject, defaults, true)
	if err != nil {
		t.Fatalf("RunPhase2: %v", err)
	}
	if action != PhaseContinue {
		t.Fatalf("action = %v, want %v", action, PhaseContinue)
	}

	if result.ProjectOverview != defaults.ProjectOverview || result.NamingConventions != defaults.NamingConventions || result.ErrorHandling != defaults.ErrorHandling || result.APIConventions != defaults.APIConventions {
		t.Fatalf("profile fields not preserved: %#v", result)
	}
	if result.ImportOrder != defaults.ImportOrder || result.ProtectedBranch != defaults.ProtectedBranch {
		t.Fatalf("convention fields not preserved: %#v", result)
	}
	if result.TestCommand != defaults.TestCommand || result.LintCommand != defaults.LintCommand || result.BuildCommand != defaults.BuildCommand {
		t.Fatalf("command fields not preserved: %#v", result)
	}
	if result.CoverageThreshold != 95 {
		t.Fatalf("CoverageThreshold = %d, want 95", result.CoverageThreshold)
	}
}

func TestRunPhase2NonInteractiveCoverageDefaultsAndBounds(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		in   int
		want int
	}{
		{name: "skipped defaults to 80", in: 0, want: 80},
		{name: "below range resets to 80", in: -1, want: 80},
		{name: "above range resets to 80", in: 101, want: 80},
		{name: "valid is preserved", in: 72, want: 72},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, _, err := RunPhase2(types.SetupScopeProject, &Phase2Result{CoverageThreshold: tc.in}, true)
			if err != nil {
				t.Fatalf("RunPhase2: %v", err)
			}
			if result.CoverageThreshold != tc.want {
				t.Fatalf("CoverageThreshold = %d, want %d", result.CoverageThreshold, tc.want)
			}
		})
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
	if got, want := info.Title(), "Features & Conventions — 3/8: Branch Pattern (previous: {ticket}/{description})"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}

	standardInfo := phase2StepInfoFor(3, types.PresetLevelStandard, defaults)
	if got, want := standardInfo.Title(), "Features & Conventions — 2/4: Branch Pattern (previous: {ticket}/{description})"; got != want {
		t.Fatalf("Title() = %q, want %q", got, want)
	}
}

func TestPhase2StepInfo_ChatModesAndOpenCode(t *testing.T) {
	t.Parallel()

	defaults := &Phase2Result{
		Preset:    types.PresetLevelCustom,
		ChatModes: []types.ChatModeId{types.ChatModeIdArchitect},
	}

	chatmodesInfo := phase2StepInfoFor(6, types.PresetLevelCustom, defaults)
	if got, want := chatmodesInfo.Title(), "Features & Conventions — 6/8: Copilot Chat Modes (previous: architect)"; got != want {
		t.Fatalf("chatmodes title = %q, want %q", got, want)
	}

	// Opencode steps (7, 8) appear only for the custom preset.
	occmdInfo := phase2StepInfoFor(7, types.PresetLevelCustom, &Phase2Result{
		Preset:           types.PresetLevelCustom,
		OpenCodeCommands: []types.OpenCodeCommandId{types.OpenCodeCommandIdReview},
	})
	if got, want := occmdInfo.Title(), "Features & Conventions — 7/8: OpenCode Commands (previous: review)"; got != want {
		t.Fatalf("opencode commands title = %q, want %q", got, want)
	}

	ocmodeInfo := phase2StepInfoFor(8, types.PresetLevelCustom, &Phase2Result{
		Preset:        types.PresetLevelCustom,
		OpenCodeModes: []types.OpenCodeModeId{types.OpenCodeModeIdPlan},
	})
	if got, want := ocmodeInfo.Title(), "Features & Conventions — 8/8: OpenCode Modes (previous: plan)"; got != want {
		t.Fatalf("opencode modes title = %q, want %q", got, want)
	}
}

// TestPhase2Stepping_NonCustomSkipsCommandsAndChatmodes verifies that after
// step 5 (Require Ticket), non-custom presets exit the loop (skip steps 6+7).
func TestPhase2Stepping_NonCustomSkipsCommandsAndChatmodes(t *testing.T) {
	t.Parallel()

	// For non-custom preset, next(5) must be >= 8 (exit).
	next := nextPhase2Step(5, types.PresetLevelStandard)
	if next <= 7 {
		t.Errorf("next(5, standard) = %d; expected >= 8 to skip commands/chatmodes", next)
	}

	// For custom preset, next(5) == 6 (show Commands step).
	if got := nextPhase2Step(5, types.PresetLevelCustom); got != 6 {
		t.Errorf("next(5, custom) = %d; want 6", got)
	}
	// next(6, custom) == 7.
	if got := nextPhase2Step(6, types.PresetLevelCustom); got != 7 {
		t.Errorf("next(6, custom) = %d; want 7", got)
	}
}

// TestBuildPhase2Result_CustomPreservesToolSelections confirms that custom preset
// carries ChatModes/OpenCodeCommands/OpenCodeModes through; non-
// custom presets strip them (preset defaults handle selection).
func TestBuildPhase2Result_CustomPreservesToolSelections(t *testing.T) {
	t.Parallel()

	modes := []types.ChatModeId{types.ChatModeIdReviewer}
	occmds := []types.OpenCodeCommandId{types.OpenCodeCommandIdReview, types.OpenCodeCommandIdTest}
	ocmodes := []types.OpenCodeModeId{types.OpenCodeModeIdPlan}

	custom := buildPhase2Result(types.SetupScopeProject, types.PresetLevelCustom, nil, "", "", false, modes, occmds, ocmodes)
	if !reflect.DeepEqual(custom.ChatModes, modes) {
		t.Errorf("custom.ChatModes = %v, want %v", custom.ChatModes, modes)
	}
	if !reflect.DeepEqual(custom.OpenCodeCommands, occmds) {
		t.Errorf("custom.OpenCodeCommands = %v, want %v", custom.OpenCodeCommands, occmds)
	}
	if !reflect.DeepEqual(custom.OpenCodeModes, ocmodes) {
		t.Errorf("custom.OpenCodeModes = %v, want %v", custom.OpenCodeModes, ocmodes)
	}

	standard := buildPhase2Result(types.SetupScopeProject, types.PresetLevelStandard, nil, "", "", false, modes, occmds, ocmodes)
	if len(standard.ChatModes) != 0 {
		t.Errorf("standard.ChatModes = %v; non-custom must drop explicit list", standard.ChatModes)
	}
	if len(standard.OpenCodeCommands) != 0 {
		t.Errorf("standard.OpenCodeCommands = %v; non-custom must drop explicit list", standard.OpenCodeCommands)
	}
	if len(standard.OpenCodeModes) != 0 {
		t.Errorf("standard.OpenCodeModes = %v; non-custom must drop explicit list", standard.OpenCodeModes)
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
