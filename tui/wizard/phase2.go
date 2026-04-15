package wizard

import (
	"fmt"

	"charm.land/huh/v2"

	"github.com/ricardoborges-teachable/ai-setup/internal/preset"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

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
	result := &Phase2Result{}
	if defaults != nil {
		*result = *defaults
	}

	// Determine preset.
	if result.Preset == "" {
		result.Preset = preset.DefaultPresetForScope(scope)
	}

	// Resolve features from preset if not already set.
	if result.Features == nil {
		resolved := preset.ResolvePreset(result.Preset)
		if resolved != nil {
			result.Features = resolved
		} else {
			defaults := types.DefaultFeatureFlags()
			result.Features = &defaults
		}
	}

	// Resolve git conventions if not set.
	if result.GitConv == nil {
		defaults := types.DefaultGitConventions()
		result.GitConv = &defaults
	}

	return result, PhaseContinue, nil
}

func runPhase2Interactive(scope types.SetupScope, defaults *Phase2Result) (*Phase2Result, PhaseAction, error) {
	// Default feature selection (all enabled).
	featureDefaultSelected := []string{
		"qualityGates", "rpiWorkflow", "chainOfThought",
		"bugResolution", "contextEngineering", "pivotHandling",
	}

	// --- Preset selection ---
	var presetValue string
	presetSelect := huh.NewSelect[string]().
		Title("Feature preset:").
		Options(
			huh.NewOption("Minimal — Quality gates + git only", "minimal"),
			huh.NewOption("Standard (recommended) — +RPI, reasoning, bug resolution", "standard"),
			huh.NewOption("Full — All features enabled", "full"),
			huh.NewOption("Custom — Pick features individually", "custom"),
		).
		Value(&presetValue)

	// We'll run preset first, then conditionally ask about features.
	form := huh.NewForm(huh.NewGroup(presetSelect).Title("Phase 2/4: Features & Conventions"))
	if err := form.Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	result := &Phase2Result{
		Preset: types.PresetLevel(presetValue),
	}

	// Resolve features based on preset.
	if result.Preset != types.PresetLevelCustom {
		resolved := preset.ResolvePreset(result.Preset)
		if resolved != nil {
			result.Features = resolved
		} else {
			f := types.DefaultFeatureFlags()
			result.Features = &f
		}
	}

	// --- Feature toggle (only for Custom preset) ---
	if result.Preset == types.PresetLevelCustom {
		var selectedFeatures []string
		selectedFeatures = append(selectedFeatures, featureDefaultSelected...)

		featureMulti := huh.NewMultiSelect[string]().
			Title("Which features should be enabled?").
			Options(featureOptions...).
			Value(&selectedFeatures)

		featureForm := huh.NewForm(huh.NewGroup(featureMulti))
		if err := featureForm.Run(); err != nil {
			return nil, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
		}

		result.Features = buildFeaturesFromSelection(selectedFeatures)
	}

	// --- Git conventions ---
	var branchPattern string
	var commitPattern string
	var requireTicket bool

	branchSelect := huh.NewSelect[string]().
		Title("Branch naming pattern:").
		Options(branchPatternOptions...).
		Value(&branchPattern)

	commitSelect := huh.NewSelect[string]().
		Title("Commit message pattern:").
		Options(commitPatternOptions...).
		Value(&commitPattern)

	ticketConfirm := huh.NewConfirm().
		Title("Require ticket ID in branches/commits?").
		Value(&requireTicket)

	gitForm := huh.NewForm(
		huh.NewGroup(branchSelect, commitSelect, ticketConfirm).
			Title("Git Conventions"),
	)
	if err := gitForm.Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	// Handle custom patterns.
	if branchPattern == "custom" {
		var customBranch string
		input := huh.NewInput().
			Title("Custom branch pattern (use {type}, {ticket}, {description}):").
			Placeholder("{type}/{ticket}-{description}").
			Value(&customBranch)
		if err := huh.NewForm(huh.NewGroup(input)).Run(); err != nil {
			return nil, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
		}
		branchPattern = customBranch
		if branchPattern == "" {
			branchPattern = "{type}/{ticket}-{description}"
		}
	}

	if commitPattern == "custom" {
		var customCommit string
		input := huh.NewInput().
			Title("Custom commit pattern (use {type}, {scope}, {ticket}, {description}):").
			Placeholder("{type}({scope}): {description}").
			Value(&customCommit)
		if err := huh.NewForm(huh.NewGroup(input)).Run(); err != nil {
			return nil, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
		}
		commitPattern = customCommit
		if commitPattern == "" {
			commitPattern = "{type}({scope}): {description}"
		}
	}

	result.GitConv = &types.GitConventions{
		BranchPattern: branchPattern,
		CommitPattern: commitPattern,
		Types: []string{
			"feat", "fix", "docs", "style", "refactor",
			"perf", "test", "build", "ci", "chore", "revert",
		},
		RequireTicket: requireTicket,
		TicketPattern: "[A-Z]+-[0-9]+",
	}

	return result, PhaseContinue, nil
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
