package wizard

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"

	"github.com/ricardoborges-teachable/ai-setup/internal/preset"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

const phase2BackValue = "__phase2_back__"

type phase2StepInfo struct {
	Current   int
	Total     int
	StepTitle string
	Previous  string
}

func (s phase2StepInfo) Title() string {
	title := fmt.Sprintf("Features & Conventions — %d/%d: %s", s.Current, s.Total, s.StepTitle)
	if s.Previous == "" {
		return title
	}
	return fmt.Sprintf("%s (previous: %s)", title, s.Previous)
}

// Phase2Result holds the feature selection results from the second wizard phase.
type Phase2Result struct {
	Preset   types.PresetLevel
	Features *types.FeatureFlags
	GitConv  *types.GitConventions
}

var branchPatternOptions = []huh.Option[string]{
	huh.NewOption("Conventional — feat/PROJ-123-add-login", "{type}/{ticket}-{description}"),
	huh.NewOption("Jira Style — task/PBG-35/creator-billing", "{type}/{ticket}/{description}"),
	huh.NewOption("Simple Type — feat/add-login", "{type}/{description}"),
	huh.NewOption("Ticket First — PROJ-123/add-login", "{ticket}/{description}"),
	huh.NewOption("Description Only — add-login", "{description}"),
	huh.NewOption("Custom Pattern", "custom"),
}

var commitPatternOptions = []huh.Option[string]{
	huh.NewOption("Conventional Commits — feat(auth): add login", "{type}({scope}): {description}"),
	huh.NewOption("Simple Type — feat: add login", "{type}: {description}"),
	huh.NewOption("Ticket Prefix — [PROJ-123] add login", "[{ticket}] {description}"),
	huh.NewOption("Description Only — add login", "{description}"),
	huh.NewOption("Custom Pattern", "custom"),
}

// featureOptions defines the available features for custom preset.
var featureOptions = []huh.Option[string]{
	huh.NewOption("Quality Gates", "qualityGates"),
	huh.NewOption("RPI Workflow", "rpiWorkflow"),
	huh.NewOption("Reasoning Protocol", "chainOfThought"),
	huh.NewOption("Bug Resolution", "bugResolution"),
	huh.NewOption("Context Discipline", "contextEngineering"),
	huh.NewOption("Decision Protocol", "treeOfThoughts"),
	huh.NewOption("ADR Enforcement", "adrEnforcement"),
	huh.NewOption("Agent Coordination", "agentHarness"),
	huh.NewOption("Pivot Handling", "pivotHandling"),
}

// RunPhase2 runs the features and conventions phase.
//
// In non-interactive mode, it applies CLI overrides or defaults derived from
// the given scope. In interactive mode, it presents preset, feature, and
// git convention prompts via huh forms.
func RunPhase2(scope types.SetupScope, defaults *Phase2Result, nonInteractive bool) (*Phase2Result, PhaseAction, error) {
	if nonInteractive {
		return runPhase2NonInteractive(scope, defaults)
	}
	return runPhase2Interactive(scope, defaults)
}

func runPhase2NonInteractive(scope types.SetupScope, defaults *Phase2Result) (*Phase2Result, PhaseAction, error) {
	var presetValue types.PresetLevel
	var features *types.FeatureFlags
	var branchPattern string
	var commitPattern string
	var requireTicket bool

	if defaults != nil {
		presetValue = defaults.Preset
		features = defaults.Features
		if defaults.GitConv != nil {
			branchPattern = defaults.GitConv.BranchPattern
			commitPattern = defaults.GitConv.CommitPattern
			requireTicket = defaults.GitConv.RequireTicket
		}
	}

	return buildPhase2Result(scope, presetValue, features, branchPattern, commitPattern, requireTicket), PhaseContinue, nil
}

func runPhase2Interactive(scope types.SetupScope, defaults *Phase2Result) (*Phase2Result, PhaseAction, error) {
	state := phase2InteractiveState{
		Preset:        defaultPhase2Preset(scope, defaults),
		Features:      defaultPhase2Features(defaults),
		BranchPattern: defaultPhase2BranchPattern(defaults),
		CommitPattern: defaultPhase2CommitPattern(defaults),
		RequireTicket: defaultPhase2RequireTicket(defaults),
	}

	currentStep := 1
	for currentStep >= 1 && currentStep <= 5 {
		presetForStep := state.Preset
		if presetForStep == "" {
			presetForStep = defaultPhase2Preset(scope, defaults)
		}

		switch currentStep {
		case 1:
			presetValue, action, err := askPreset(state.Preset, phase2StepInfoFor(1, presetForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				return nil, PhaseBack, nil
			}
			state.Preset = presetValue
			if state.Preset != types.PresetLevelCustom {
				state.Features = nil
			}
			currentStep = nextPhase2Step(currentStep, state.Preset)
		case 2:
			selectedFeatures, action, err := askFeatures(state.Features, phase2StepInfoFor(2, presetForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase2Step(currentStep, state.Preset)
				continue
			}
			state.Features = buildFeaturesFromSelection(selectedFeatures)
			currentStep = nextPhase2Step(currentStep, state.Preset)
		case 3:
			branchPattern, action, err := askBranchPattern(state.BranchPattern, phase2StepInfoFor(3, presetForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase2Step(currentStep, state.Preset)
				continue
			}
			state.BranchPattern = branchPattern
			currentStep++
		case 4:
			commitPattern, action, err := askCommitPattern(state.CommitPattern, phase2StepInfoFor(4, presetForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase2Step(currentStep, state.Preset)
				continue
			}
			state.CommitPattern = commitPattern
			currentStep++
		case 5:
			requireTicket, action, err := askRequireTicket(state.RequireTicket, phase2StepInfoFor(5, presetForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase2Step(currentStep, state.Preset)
				continue
			}
			state.RequireTicket = requireTicket
			currentStep++
		}
	}

	return buildPhase2Result(scope, state.Preset, state.Features, state.BranchPattern, state.CommitPattern, state.RequireTicket), PhaseContinue, nil
}

type phase2InteractiveState struct {
	Preset        types.PresetLevel
	Features      *types.FeatureFlags
	BranchPattern string
	CommitPattern string
	RequireTicket bool
}

func buildPhase2Result(scope types.SetupScope, presetValue types.PresetLevel, features *types.FeatureFlags, branch, commit string, requireTicket bool) *Phase2Result {
	resolvedPreset := presetValue
	if resolvedPreset == "" {
		resolvedPreset = preset.DefaultPresetForScope(scope)
	}

	resolvedFeatures := cloneFeatureFlags(features)
	if resolvedFeatures == nil {
		resolvedFeatures = preset.ResolvePreset(resolvedPreset)
		if resolvedFeatures == nil {
			defaults := types.DefaultFeatureFlags()
			resolvedFeatures = &defaults
		}
	}

	gitDefaults := types.DefaultGitConventions()
	if branch == "" {
		branch = gitDefaults.BranchPattern
	}
	if commit == "" {
		commit = gitDefaults.CommitPattern
	}

	return &Phase2Result{
		Preset:   resolvedPreset,
		Features: resolvedFeatures,
		GitConv: &types.GitConventions{
			BranchPattern: branch,
			CommitPattern: commit,
			Types:         cloneStrings(gitDefaults.Types),
			RequireTicket: requireTicket,
			TicketPattern: gitDefaults.TicketPattern,
		},
	}
}

func askPreset(current types.PresetLevel, info phase2StepInfo) (types.PresetLevel, PhaseAction, error) {
	presetValue := string(current)
	field := huh.NewSelect[string]().
		Title(info.Title()).
		Options(
			huh.NewOption("Minimal — Quality gates + git only", "minimal"),
			huh.NewOption("Standard (recommended) — +RPI, reasoning, bug resolution", "standard"),
			huh.NewOption("Full — All features enabled", "full"),
			huh.NewOption("Custom — Pick features individually", "custom"),
		).
		Value(&presetValue)

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	return types.PresetLevel(presetValue), PhaseContinue, nil
}

func askFeatures(current *types.FeatureFlags, info phase2StepInfo) ([]string, PhaseAction, error) {
	selectedFeatures := featureSelectionFromFlags(current)
	if len(selectedFeatures) == 0 {
		selectedFeatures = defaultPhase2FeatureSelection()
	}

	field := huh.NewMultiSelect[string]().
		Title(info.Title()).
		Options(appendPhase2BackOption(featureOptions)...).
		Value(&selectedFeatures)

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}
	if containsString(selectedFeatures, phase2BackValue) {
		return nil, PhaseBack, nil
	}

	return selectedFeatures, PhaseContinue, nil
}

func askBranchPattern(current string, info phase2StepInfo) (string, PhaseAction, error) {
	branchPattern := current
	if branchPattern == "" {
		branchPattern = types.DefaultGitConventions().BranchPattern
	}

	field := huh.NewSelect[string]().
		Title(info.Title()).
		Options(appendPhase2BackOption(branchPatternOptions)...).
		Value(&branchPattern)

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}
	if branchPattern == phase2BackValue {
		return "", PhaseBack, nil
	}
	if branchPattern != "custom" {
		return branchPattern, PhaseContinue, nil
	}

	customBranch := current
	if customBranch == "" || customBranch == "custom" {
		customBranch = types.DefaultGitConventions().BranchPattern
	}
	input := huh.NewInput().
		Title("Custom branch pattern (use {type}, {ticket}, {description}):").
		Placeholder(types.DefaultGitConventions().BranchPattern).
		Value(&customBranch)

	if err := huh.NewForm(huh.NewGroup(input).Title(info.Title())).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	if customBranch == "" {
		customBranch = types.DefaultGitConventions().BranchPattern
	}
	return customBranch, PhaseContinue, nil
}

func askCommitPattern(current string, info phase2StepInfo) (string, PhaseAction, error) {
	commitPattern := current
	if commitPattern == "" {
		commitPattern = types.DefaultGitConventions().CommitPattern
	}

	field := huh.NewSelect[string]().
		Title(info.Title()).
		Options(appendPhase2BackOption(commitPatternOptions)...).
		Value(&commitPattern)

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}
	if commitPattern == phase2BackValue {
		return "", PhaseBack, nil
	}
	if commitPattern != "custom" {
		return commitPattern, PhaseContinue, nil
	}

	customCommit := current
	if customCommit == "" || customCommit == "custom" {
		customCommit = types.DefaultGitConventions().CommitPattern
	}
	input := huh.NewInput().
		Title("Custom commit pattern (use {type}, {scope}, {ticket}, {description}):").
		Placeholder(types.DefaultGitConventions().CommitPattern).
		Value(&customCommit)

	if err := huh.NewForm(huh.NewGroup(input).Title(info.Title())).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	if customCommit == "" {
		customCommit = types.DefaultGitConventions().CommitPattern
	}
	return customCommit, PhaseContinue, nil
}

func askRequireTicket(current bool, info phase2StepInfo) (bool, PhaseAction, error) {
	requireTicket := current
	field := huh.NewConfirm().
		Title(info.Title()).
		Value(&requireTicket)

	if err := huh.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return false, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	return requireTicket, PhaseContinue, nil
}

func defaultPhase2Preset(scope types.SetupScope, defaults *Phase2Result) types.PresetLevel {
	if defaults != nil && defaults.Preset != "" {
		return defaults.Preset
	}
	return preset.DefaultPresetForScope(scope)
}

func defaultPhase2Features(defaults *Phase2Result) *types.FeatureFlags {
	if defaults == nil {
		return nil
	}
	return cloneFeatureFlags(defaults.Features)
}

func defaultPhase2BranchPattern(defaults *Phase2Result) string {
	if defaults != nil && defaults.GitConv != nil && defaults.GitConv.BranchPattern != "" {
		return defaults.GitConv.BranchPattern
	}
	return types.DefaultGitConventions().BranchPattern
}

func defaultPhase2CommitPattern(defaults *Phase2Result) string {
	if defaults != nil && defaults.GitConv != nil && defaults.GitConv.CommitPattern != "" {
		return defaults.GitConv.CommitPattern
	}
	return types.DefaultGitConventions().CommitPattern
}

func defaultPhase2RequireTicket(defaults *Phase2Result) bool {
	if defaults != nil && defaults.GitConv != nil {
		return defaults.GitConv.RequireTicket
	}
	return types.DefaultGitConventions().RequireTicket
}

func defaultPhase2FeatureSelection() []string {
	return []string{
		"qualityGates", "rpiWorkflow", "chainOfThought",
		"bugResolution", "contextEngineering", "pivotHandling",
	}
}

func featureSelectionFromFlags(flags *types.FeatureFlags) []string {
	if flags == nil {
		return nil
	}

	selected := make([]string, 0, len(featureOptions))
	if flags.QualityGates {
		selected = append(selected, "qualityGates")
	}
	if flags.RPIWorkflow {
		selected = append(selected, "rpiWorkflow")
	}
	if flags.ChainOfThought {
		selected = append(selected, "chainOfThought")
	}
	if flags.BugResolution {
		selected = append(selected, "bugResolution")
	}
	if flags.ContextEngineering {
		selected = append(selected, "contextEngineering")
	}
	if flags.TreeOfThoughts {
		selected = append(selected, "treeOfThoughts")
	}
	if flags.ADREnforcement {
		selected = append(selected, "adrEnforcement")
	}
	if flags.AgentHarness {
		selected = append(selected, "agentHarness")
	}
	if flags.PivotHandling {
		selected = append(selected, "pivotHandling")
	}
	return selected
}

func phase2StepInfoFor(step int, presetValue types.PresetLevel, defaults *Phase2Result) phase2StepInfo {
	total := phase2Total(presetValue)
	current := step
	if presetValue != types.PresetLevelCustom && step >= 3 {
		current--
	}

	info := phase2StepInfo{Current: current, Total: total}
	switch step {
	case 1:
		info.StepTitle = "Preset"
		if defaults != nil && defaults.Preset != "" {
			info.Previous = string(defaults.Preset)
		}
	case 2:
		info.StepTitle = "Features"
		if defaults != nil && defaults.Features != nil {
			info.Previous = strings.Join(featureSelectionFromFlags(defaults.Features), ", ")
		}
	case 3:
		info.StepTitle = "Branch Pattern"
		if defaults != nil && defaults.GitConv != nil && defaults.GitConv.BranchPattern != "" {
			info.Previous = defaults.GitConv.BranchPattern
		}
	case 4:
		info.StepTitle = "Commit Pattern"
		if defaults != nil && defaults.GitConv != nil && defaults.GitConv.CommitPattern != "" {
			info.Previous = defaults.GitConv.CommitPattern
		}
	case 5:
		info.StepTitle = "Require Ticket"
		if defaults != nil && defaults.GitConv != nil {
			info.Previous = fmt.Sprintf("%t", defaults.GitConv.RequireTicket)
		}
	}
	return info
}

func phase2Total(presetValue types.PresetLevel) int {
	if presetValue == types.PresetLevelCustom {
		return 5
	}
	return 4
}

func previousPhase2Step(current int, presetValue types.PresetLevel) int {
	previous := current - 1
	if presetValue != types.PresetLevelCustom && previous == 2 {
		previous--
	}
	if previous < 1 {
		return 1
	}
	return previous
}

func nextPhase2Step(current int, presetValue types.PresetLevel) int {
	next := current + 1
	if presetValue != types.PresetLevelCustom && next == 2 {
		next++
	}
	return next
}

func appendPhase2BackOption(options []huh.Option[string]) []huh.Option[string] {
	result := append([]huh.Option[string]{}, options...)
	return append(result, huh.NewOption("↩ Back", phase2BackValue))
}

func cloneFeatureFlags(flags *types.FeatureFlags) *types.FeatureFlags {
	if flags == nil {
		return nil
	}
	cp := *flags
	return &cp
}

// buildFeaturesFromSelection converts a slice of feature flag name strings
// to a FeatureFlags struct.
func buildFeaturesFromSelection(selected []string) *types.FeatureFlags {
	f := &types.FeatureFlags{}
	for _, s := range selected {
		switch s {
		case "contextEngineering":
			f.ContextEngineering = true
		case "rpiWorkflow":
			f.RPIWorkflow = true
		case "chainOfThought":
			f.ChainOfThought = true
		case "treeOfThoughts":
			f.TreeOfThoughts = true
		case "adrEnforcement":
			f.ADREnforcement = true
		case "qualityGates":
			f.QualityGates = true
		case "agentHarness":
			f.AgentHarness = true
		case "bugResolution":
			f.BugResolution = true
		case "pivotHandling":
			f.PivotHandling = true
		}
	}
	return f
}
