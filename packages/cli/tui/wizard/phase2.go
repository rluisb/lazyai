package wizard

import (
	"fmt"
	"strings"

	"charm.land/huh/v2"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"

	"github.com/rluisb/lazyai/packages/cli/internal/preset"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
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
	Preset            types.PresetLevel
	Features          *types.FeatureFlags
	GitConv           *types.GitConventions
	ChatModes         []types.ChatModeId        // populated only when Preset == Custom
	OpenCodeCommands  []types.OpenCodeCommandId // populated only when Preset == Custom
	OpenCodeModes     []types.OpenCodeModeId    // populated only when Preset == Custom
	ProjectOverview   string
	NamingConventions string
	ErrorHandling     string
	APIConventions    string
	ImportOrder       string
	ProtectedBranch   string
	TestCommand       string
	LintCommand       string
	BuildCommand      string
	CoverageThreshold int
	// UseReversa controls whether deterministic Scout/Reversa analysis may
	// auto-populate mechanical project details. Nil preserves the legacy default
	// of enabled for callers that construct Phase2Result directly.
	UseReversa *bool
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
	huh.NewOption("Adversarial Design Review", "adversarialDesign"),
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

	result := buildPhase2Result(scope, presetValue, features, branchPattern, commitPattern, requireTicket, nil, nil, nil)
	applyPhase2ProfileDefaults(result, defaults)
	return result, PhaseContinue, nil
}

func runPhase2Interactive(scope types.SetupScope, defaults *Phase2Result) (*Phase2Result, PhaseAction, error) {
	state := phase2InteractiveState{
		Preset:        defaultPhase2Preset(scope, defaults),
		Features:      defaultPhase2Features(defaults),
		BranchPattern: defaultPhase2BranchPattern(defaults),
		CommitPattern: defaultPhase2CommitPattern(defaults),
		RequireTicket: defaultPhase2RequireTicket(defaults),
		UseReversa:    defaultUseReversa(defaults),
	}

	currentStep := 1
	for currentStep >= 1 && currentStep <= 8 {
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
				state.ChatModes = nil
				state.OpenCodeCommands = nil
				state.OpenCodeModes = nil
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
			currentStep = nextPhase2Step(currentStep, state.Preset)
		case 6:
			chatmodes, action, err := askChatModes(state.ChatModes, phase2StepInfoFor(6, presetForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase2Step(currentStep, state.Preset)
				continue
			}
			state.ChatModes = chatmodes
			currentStep = nextPhase2Step(currentStep, state.Preset)
		case 7:
			occmds, action, err := askOpenCodeCommands(state.OpenCodeCommands, phase2StepInfoFor(7, presetForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase2Step(currentStep, state.Preset)
				continue
			}
			state.OpenCodeCommands = occmds
			currentStep = nextPhase2Step(currentStep, state.Preset)
		case 8:
			ocmodes, action, err := askOpenCodeModes(state.OpenCodeModes, phase2StepInfoFor(8, presetForStep, defaults))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = previousPhase2Step(currentStep, state.Preset)
				continue
			}
			state.OpenCodeModes = ocmodes
			currentStep++
		}
	}

	result := buildPhase2Result(scope, state.Preset, state.Features, state.BranchPattern, state.CommitPattern, state.RequireTicket, state.ChatModes, state.OpenCodeCommands, state.OpenCodeModes)
	result.UseReversa = boolPtr(state.UseReversa)
	return result, PhaseContinue, nil
}

type phase2InteractiveState struct {
	Preset            types.PresetLevel
	Features          *types.FeatureFlags
	BranchPattern     string
	CommitPattern     string
	RequireTicket     bool
	ChatModes         []types.ChatModeId
	OpenCodeCommands  []types.OpenCodeCommandId
	OpenCodeModes     []types.OpenCodeModeId
	ProjectOverview   string
	NamingConventions string
	ErrorHandling     string
	APIConventions    string
	ImportOrder       string
	ProtectedBranch   string
	TestCommand       string
	LintCommand       string
	BuildCommand      string
	CoverageThreshold int
	UseReversa        bool
}

func buildPhase2Result(scope types.SetupScope, presetValue types.PresetLevel, features *types.FeatureFlags, branch, commit string, requireTicket bool, chatmodes []types.ChatModeId, opencodeCommands []types.OpenCodeCommandId, opencodeModes []types.OpenCodeModeId) *Phase2Result {
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

	// Only carry explicit selections when the user chose custom preset.
	// For non-custom presets, caller (helpers.go) falls back to ALL_* defaults.
	var resolvedChatModes []types.ChatModeId
	var resolvedOpenCodeCommands []types.OpenCodeCommandId
	var resolvedOpenCodeModes []types.OpenCodeModeId
	if resolvedPreset == types.PresetLevelCustom {
		resolvedChatModes = append([]types.ChatModeId(nil), chatmodes...)
		resolvedOpenCodeCommands = append([]types.OpenCodeCommandId(nil), opencodeCommands...)
		resolvedOpenCodeModes = append([]types.OpenCodeModeId(nil), opencodeModes...)
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
		ChatModes:         resolvedChatModes,
		OpenCodeCommands:  resolvedOpenCodeCommands,
		OpenCodeModes:     resolvedOpenCodeModes,
		CoverageThreshold: defaultCoverageThreshold(),
		UseReversa:        boolPtr(defaultUseReversa(nil)),
	}
}

func applyPhase2ProfileDefaults(result, defaults *Phase2Result) {
	if result == nil {
		return
	}
	if defaults == nil {
		result.CoverageThreshold = normalizeCoverageThreshold(result.CoverageThreshold)
		if result.UseReversa == nil {
			result.UseReversa = boolPtr(defaultUseReversa(nil))
		}
		return
	}

	result.ProjectOverview = defaults.ProjectOverview
	result.NamingConventions = defaults.NamingConventions
	result.ErrorHandling = defaults.ErrorHandling
	result.APIConventions = defaults.APIConventions
	result.ImportOrder = defaults.ImportOrder
	result.ProtectedBranch = defaults.ProtectedBranch
	result.TestCommand = defaults.TestCommand
	result.LintCommand = defaults.LintCommand
	result.BuildCommand = defaults.BuildCommand
	result.CoverageThreshold = normalizeCoverageThreshold(defaults.CoverageThreshold)
	if defaults.UseReversa != nil {
		result.UseReversa = boolPtr(*defaults.UseReversa)
	} else if result.UseReversa == nil {
		result.UseReversa = boolPtr(defaultUseReversa(nil))
	}
}

func defaultUseReversa(defaults *Phase2Result) bool {
	if defaults != nil && defaults.UseReversa != nil {
		return *defaults.UseReversa
	}
	return true
}

func boolPtr(value bool) *bool {
	return &value
}

func defaultCoverageThreshold() int {
	return 80
}

func normalizeCoverageThreshold(value int) int {
	if value >= 1 && value <= 100 {
		return value
	}
	return defaultCoverageThreshold()
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
	field.DescriptionFunc(func() string {
		return selectHoverDescription(field, presetDescriptions, defaultHoverHint)
	}, field)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
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
		Options(appendPhase2BackOption(optionsWithDescriptions(featureOptions, featureDescriptions))...).
		Value(&selectedFeatures)
	field.DescriptionFunc(func() string {
		return multiSelectHoverDescription(field, featureDescriptions, defaultHoverHint)
	}, field)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
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
	field.DescriptionFunc(func() string {
		return selectHoverDescription(field, branchPatternDescriptions, defaultHoverHint)
	}, field)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
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
		Description("Template for new branch names; supported placeholders are shown in the title.").
		Placeholder(types.DefaultGitConventions().BranchPattern).
		Value(&customBranch)

	if err := theme.NewForm(huh.NewGroup(input).Title(info.Title())).Run(); err != nil {
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
	field.DescriptionFunc(func() string {
		return selectHoverDescription(field, commitPatternDescriptions, defaultHoverHint)
	}, field)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
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
		Description("Template for commit messages; supported placeholders are shown in the title.").
		Placeholder(types.DefaultGitConventions().CommitPattern).
		Value(&customCommit)

	if err := theme.NewForm(huh.NewGroup(input).Title(info.Title())).Run(); err != nil {
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
		Description("When enabled, branch and commit guidance expects a ticket placeholder.").
		Value(&requireTicket)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return false, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	return requireTicket, PhaseContinue, nil
}

func askChatModes(current []types.ChatModeId, info phase2StepInfo) ([]types.ChatModeId, PhaseAction, error) {
	selected := chatModeIdsToStrings(current)
	if len(selected) == 0 {
		selected = chatModeIdsToStrings(types.ALL_CHATMODES[:])
	}

	options := []huh.Option[string]{
		huh.NewOption("Architect mode (architect)", string(types.ChatModeIdArchitect)),
		huh.NewOption("Reviewer mode (reviewer)", string(types.ChatModeIdReviewer)),
	}

	field := huh.NewMultiSelect[string]().
		Title(info.Title()).
		Options(appendPhase2BackOption(optionsWithDescriptions(options, chatModeDescriptions))...).
		Value(&selected)
	field.DescriptionFunc(func() string {
		return multiSelectHoverDescription(field, chatModeDescriptions, defaultHoverHint)
	}, field)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	if containsString(selected, phase2BackValue) {
		return nil, PhaseBack, nil
	}
	return stringsToChatModeIds(selected), PhaseContinue, nil
}

func askOpenCodeCommands(current []types.OpenCodeCommandId, info phase2StepInfo) ([]types.OpenCodeCommandId, PhaseAction, error) {
	selected := opencodeCommandIdsToStrings(current)
	if len(selected) == 0 {
		selected = opencodeCommandIdsToStrings(types.ALL_OPENCODE_COMMANDS[:])
	}

	options := []huh.Option[string]{
		huh.NewOption("Review branch (review)", string(types.OpenCodeCommandIdReview)),
		huh.NewOption("Run tests (test)", string(types.OpenCodeCommandIdTest)),
		huh.NewOption("Draft commit (commit)", string(types.OpenCodeCommandIdCommit)),
	}

	field := huh.NewMultiSelect[string]().
		Title(info.Title()).
		Options(appendPhase2BackOption(optionsWithDescriptions(options, opencodeCommandDescriptions))...).
		Value(&selected)
	field.DescriptionFunc(func() string {
		return multiSelectHoverDescription(field, opencodeCommandDescriptions, defaultHoverHint)
	}, field)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	if containsString(selected, phase2BackValue) {
		return nil, PhaseBack, nil
	}
	return stringsToOpenCodeCommandIds(selected), PhaseContinue, nil
}

func askOpenCodeModes(current []types.OpenCodeModeId, info phase2StepInfo) ([]types.OpenCodeModeId, PhaseAction, error) {
	selected := opencodeModeIdsToStrings(current)
	if len(selected) == 0 {
		selected = opencodeModeIdsToStrings(types.ALL_OPENCODE_MODES[:])
	}

	options := []huh.Option[string]{
		huh.NewOption("Plan mode (plan)", string(types.OpenCodeModeIdPlan)),
		huh.NewOption("Audit mode (audit)", string(types.OpenCodeModeIdAudit)),
	}

	field := huh.NewMultiSelect[string]().
		Title(info.Title()).
		Options(appendPhase2BackOption(optionsWithDescriptions(options, opencodeModeDescriptions))...).
		Value(&selected)
	field.DescriptionFunc(func() string {
		return multiSelectHoverDescription(field, opencodeModeDescriptions, defaultHoverHint)
	}, field)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 2 cancelled: %w", err)
	}

	if containsString(selected, phase2BackValue) {
		return nil, PhaseBack, nil
	}
	return stringsToOpenCodeModeIds(selected), PhaseContinue, nil
}

func opencodeCommandIdsToStrings(ids []types.OpenCodeCommandId) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

func stringsToOpenCodeCommandIds(values []string) []types.OpenCodeCommandId {
	out := make([]types.OpenCodeCommandId, 0, len(values))
	for _, v := range values {
		if v == phase2BackValue {
			continue
		}
		out = append(out, types.OpenCodeCommandId(v))
	}
	return out
}

func opencodeModeIdsToStrings(ids []types.OpenCodeModeId) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

func stringsToOpenCodeModeIds(values []string) []types.OpenCodeModeId {
	out := make([]types.OpenCodeModeId, 0, len(values))
	for _, v := range values {
		if v == phase2BackValue {
			continue
		}
		out = append(out, types.OpenCodeModeId(v))
	}
	return out
}

func chatModeIdsToStrings(ids []types.ChatModeId) []string {
	out := make([]string, 0, len(ids))
	for _, id := range ids {
		out = append(out, string(id))
	}
	return out
}

func stringsToChatModeIds(values []string) []types.ChatModeId {
	out := make([]types.ChatModeId, 0, len(values))
	for _, v := range values {
		if v == phase2BackValue {
			continue
		}
		out = append(out, types.ChatModeId(v))
	}
	return out
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
	if flags.AdversarialDesign {
		selected = append(selected, "adversarialDesign")
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
	case 6:
		info.StepTitle = "Copilot Chat Modes"
		if defaults != nil && len(defaults.ChatModes) > 0 {
			info.Previous = strings.Join(chatModeIdsToStrings(defaults.ChatModes), ", ")
		}
	case 7:
		info.StepTitle = "OpenCode Commands"
		if defaults != nil && len(defaults.OpenCodeCommands) > 0 {
			info.Previous = strings.Join(opencodeCommandIdsToStrings(defaults.OpenCodeCommands), ", ")
		}
	case 8:
		info.StepTitle = "OpenCode Modes"
		if defaults != nil && len(defaults.OpenCodeModes) > 0 {
			info.Previous = strings.Join(opencodeModeIdsToStrings(defaults.OpenCodeModes), ", ")
		}
	}
	return info
}

func phase2Total(presetValue types.PresetLevel) int {
	if presetValue == types.PresetLevelCustom {
		return 8
	}
	return 4
}

func previousPhase2Step(current int, presetValue types.PresetLevel) int {
	previous := current - 1
	// Non-custom presets skip the Features step (2) and custom-only
	// artifact selection steps.
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
	// After RequireTicket (step 5), jump past the custom-only selection
	// steps (6 = Copilot Chat Modes, 7 = OpenCode Commands, 8 = OpenCode
	// Modes) for non-custom presets.
	if presetValue != types.PresetLevelCustom && next == 6 {
		next = 9 // exit loop (loop condition is currentStep <= 8)
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
		case "adversarialDesign":
			f.AdversarialDesign = true
		}
	}
	return f
}
