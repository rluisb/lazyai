package wizard

import (
	"context"
	"fmt"
	"os/exec"

	"charm.land/huh/v2"

	"github.com/rluisb/lazyai/packages/cli/internal/auth"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
)

type phase5StepInfo struct {
	Current   int
	Total     int
	StepTitle string
}

func (s phase5StepInfo) Title() string {
	return fmt.Sprintf("Optional Tooling — %d/%d: %s", s.Current, s.Total, s.StepTitle)
}

type Phase5Result struct {
	MemoryPath        string
	EnableObsidian    bool
	ObsidianVaultPath string
	EnableQmd         bool
	QmdIndexPath      string
	EnableCodegraph   bool
	CodegraphDataPath string
	EnableGraphify    bool
	GraphifyDataPath  string
	OpenCodePlugins   []string
	// OpenCodeProviders are the provider IDs (e.g., "openai", "ollama-cloud")
	// the user has authenticated and chosen to expose to OpenCode-side agents
	// at install time. Empty when OpenCode isn't selected; otherwise
	// populated from auth.DetectAll filtered against the OpenCode catalog's
	// deny rules (Anthropic excluded).
	OpenCodeProviders []string
}

// RunPhase5 runs the optional tooling phase.
// opencodeSelected is true when opencode is in the user's tool list — it gates
// the plugin install step (which also requires the binary to be on PATH).
func RunPhase5(defaults *Phase5Result, nonInteractive bool, opencodeSelected ...bool) (*Phase5Result, PhaseAction, error) {
	selected := len(opencodeSelected) > 0 && opencodeSelected[0]
	if nonInteractive {
		return runPhase5NonInteractive(defaults)
	}
	return runPhase5Interactive(defaults, selected)
}

func runPhase5NonInteractive(defaults *Phase5Result) (*Phase5Result, PhaseAction, error) {
	result := defaultPhase5Result()
	if defaults != nil {
		result = *defaults
	}
	return buildPhase5Result(
		result.MemoryPath,
		result.EnableObsidian,
		result.ObsidianVaultPath,
		result.EnableQmd,
		result.QmdIndexPath,
		result.EnableCodegraph,
		result.CodegraphDataPath,
		result.EnableGraphify,
		result.GraphifyDataPath,
		result.OpenCodePlugins,
	), PhaseContinue, nil
}

func runPhase5Interactive(defaults *Phase5Result, opencodeSelected bool) (*Phase5Result, PhaseAction, error) {
	state := defaultPhase5Result()
	if defaults != nil {
		state = *defaults
	}

	// Plugin step is shown only when opencode is selected AND binary is on PATH.
	showPlugins := opencodeSelected && opencodeBinaryPresent()
	// Provider step also gates on opencode selection — same condition as
	// plugins. Providers are derived from a live auth probe so they only
	// make sense when the OpenCode adapter is part of this run.
	showProviders := opencodeSelected

	currentStep := 1
	maxStep := phase5TotalSteps(showPlugins)
	if showProviders {
		maxStep++
	}
	for currentStep >= 1 && currentStep <= maxStep {
		switch currentStep {
		case 1:
			memoryPath, action, err := askMemoryPath(state.MemoryPath, phase5MemoryPathStepInfo(state, showPlugins))
			if err != nil {
				return nil, action, err
			}
			state.MemoryPath = memoryPath
			currentStep++
		case 2:
			if !showPlugins {
				currentStep++
				continue
			}
			plugins, action, err := askOpenCodePlugins(state.OpenCodePlugins, phase5OpenCodePluginsStepInfo(state, showPlugins))
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = 1
				continue
			}
			state.OpenCodePlugins = plugins
			currentStep++
		case 3:
			if !showProviders {
				currentStep++
				continue
			}
			providers, action, err := askOpenCodeProviders(state.OpenCodeProviders)
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				if showPlugins {
					currentStep = 2
				} else {
					currentStep = 1
				}
				continue
			}
			state.OpenCodeProviders = providers
			currentStep++
		}
	}

	result := buildPhase5Result(
		state.MemoryPath,
		state.EnableObsidian,
		state.ObsidianVaultPath,
		state.EnableQmd,
		state.QmdIndexPath,
		state.EnableCodegraph,
		state.CodegraphDataPath,
		state.EnableGraphify,
		state.GraphifyDataPath,
		state.OpenCodePlugins,
	)
	result.OpenCodeProviders = state.OpenCodeProviders
	return result, PhaseContinue, nil
}

func buildPhase5Result(memoryPath string, enableObsidian bool, obsidianVaultPath string, enableQmd bool, qmdIndexPath string, enableCodegraph bool, codegraphDataPath string, enableGraphify bool, graphifyDataPath string, opencodePlugins []string) *Phase5Result {
	if memoryPath == "" {
		memoryPath = ".specify/memory"
	}
	if enableCodegraph && codegraphDataPath == "" {
		codegraphDataPath = ".codegraph/"
	}
	if enableGraphify && graphifyDataPath == "" {
		graphifyDataPath = "graphify-out"
	}

	return &Phase5Result{
		MemoryPath:        memoryPath,
		EnableObsidian:    enableObsidian,
		ObsidianVaultPath: obsidianVaultPath,
		EnableQmd:         enableQmd,
		QmdIndexPath:      qmdIndexPath,
		EnableCodegraph:   enableCodegraph,
		CodegraphDataPath: codegraphDataPath,
		EnableGraphify:    enableGraphify,
		GraphifyDataPath:  graphifyDataPath,
		OpenCodePlugins:   opencodePlugins,
	}
}

func defaultPhase5Result() Phase5Result {
	return Phase5Result{
		MemoryPath:        ".specify/memory",
		EnableObsidian:    true,
		EnableQmd:         true,
		EnableCodegraph:   true,
		CodegraphDataPath: ".codegraph/",
		EnableGraphify:    true,
		GraphifyDataPath:  "graphify-out",
	}
}

func askMemoryPath(defaultValue string, info phase5StepInfo) (string, PhaseAction, error) {
	memoryPath := defaultValue
	group := huh.NewGroup(
		huh.NewInput().Title(info.Title()).Description("Project-local default for bootstrap and housekeeping.").Placeholder(".specify/memory").Value(&memoryPath),
	)

	if err := theme.NewForm(group).Run(); err != nil {
		return "", PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	return memoryPath, PhaseContinue, nil
}

func phase5MemoryPathStepInfo(_ Phase5Result, showPlugins ...bool) phase5StepInfo {
	return phase5StepInfo{Current: 1, Total: phase5TotalSteps(showPlugins...), StepTitle: "Memory Path"}
}

func phase5OpenCodePluginsStepInfo(_ Phase5Result, showPlugins bool) phase5StepInfo {
	return phase5StepInfo{Current: phase5TotalSteps(showPlugins), Total: phase5TotalSteps(showPlugins), StepTitle: "OpenCode Plugins"}
}

func phase5TotalSteps(showPlugins ...bool) int {
	if len(showPlugins) > 0 && showPlugins[0] {
		return 2
	}
	return 1
}

func opencodeBinaryPresent() bool {
	_, err := exec.LookPath("opencode")
	return err == nil
}

func askOpenCodePlugins(current []string, info phase5StepInfo) ([]string, PhaseAction, error) {
	selected := append([]string(nil), current...)

	field := huh.NewMultiSelect[string]().
		Title(info.Title()).
		Description("Select OpenCode plugins to install via `opencode plugin`. Deselect to skip.").
		Options(append(opencodePluginOptions(), huh.NewOption("↩ Back", "__phase5_back__"))...).
		Value(&selected)

	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	filtered := selected[:0]
	for _, s := range selected {
		if s != "__phase5_back__" {
			filtered = append(filtered, s)
		}
	}
	for _, s := range selected {
		if s == "__phase5_back__" {
			return nil, PhaseBack, nil
		}
	}
	return filtered, PhaseContinue, nil
}

func opencodePluginOptions() []huh.Option[string] {
	options := make([]huh.Option[string], 0, len(opencodePluginURLs))
	for _, url := range opencodePluginURLs {
		options = append(options, huh.NewOption(url, url))
	}
	return options
}

var opencodePluginURLs = []string{
	"https://github.com/Opencode-DCP/opencode-dynamic-context-pruning",
	"https://github.com/spoons-and-mirrors/subtask2",
	"https://github.com/JRedeker/opencode-shell-strategy",
	"https://github.com/boxpositron/envsitter-guard",
	"https://github.com/kdcokenny/opencode-background-agents",
}

// askOpenCodeProviders runs a live auth probe, filters out providers OpenCode
// rejects (anthropic), and presents the remaining set as a multiselect. The
// answer flows into Phase5Result.OpenCodeProviders and from there into
// AdapterContext.ConfiguredProviders so models.Resolve picks only from
// providers the user has actually authenticated.
//
// Falls through quietly when no eligible provider is detected — the
// adapter still has the live-probe fallback at install time, so this isn't
// a hard requirement.
func askOpenCodeProviders(current []string) ([]string, PhaseAction, error) {
	probeCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	detected := auth.DetectAll(probeCtx)
	eligible := filterOpenCodeEligible(detected)
	if len(eligible) == 0 {
		return nil, PhaseContinue, nil
	}

	options := make([]huh.Option[string], 0, len(eligible)+1)
	for _, p := range eligible {
		options = append(options, huh.NewOption(opencodeProviderLabel(p), string(p)))
	}
	options = append(options, huh.NewOption("↩ Back", "__phase5_back__"))

	selected := append([]string(nil), current...)
	if len(selected) == 0 {
		// Default: pre-select every detected eligible provider.
		for _, p := range eligible {
			selected = append(selected, string(p))
		}
	}

	field := huh.NewMultiSelect[string]().
		Title("OpenCode Providers").
		Description("Authenticated providers OpenCode-side agents may pull models from. Anthropic is excluded by policy.").
		Options(options...).
		Value(&selected)
	if err := theme.NewForm(huh.NewGroup(field)).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	filtered := selected[:0]
	wantBack := false
	for _, s := range selected {
		if s == "__phase5_back__" {
			wantBack = true
			continue
		}
		filtered = append(filtered, s)
	}
	if wantBack {
		return nil, PhaseBack, nil
	}
	return filtered, PhaseContinue, nil
}

// filterOpenCodeEligible drops providers blocked by OpenCodeCatalog
// (currently just "anthropic"). Importing models here would create a cycle
// with auth, so the deny list is hard-coded; if the catalog gains another
// hard-deny provider, update both this function and OpenCodeCatalog.
func filterOpenCodeEligible(detected []auth.ProviderID) []auth.ProviderID {
	out := make([]auth.ProviderID, 0, len(detected))
	for _, p := range detected {
		if p == auth.ProviderAnthropic {
			continue
		}
		out = append(out, p)
	}
	return out
}

func opencodeProviderLabel(p auth.ProviderID) string {
	for _, def := range auth.Probes {
		if def.ID == p {
			return def.Label + "  (" + string(p) + ")"
		}
	}
	return string(p)
}

// opencodeProviderHuhOptions runs auth.DetectAll, drops disallowed providers
// (anthropic), and returns huh options for the wizard's combined-form path.
// The dynamic OptionsFunc reruns this when the tool selection changes — but
// since it ignores its own args and returns based on current detection,
// the result is stable per process run.
func opencodeProviderHuhOptions() []huh.Option[string] {
	probeCtx, cancel := context.WithCancel(context.Background())
	defer cancel()
	detected := auth.DetectAll(probeCtx)
	eligible := filterOpenCodeEligible(detected)
	out := make([]huh.Option[string], 0, len(eligible))
	for _, p := range eligible {
		out = append(out, huh.NewOption(opencodeProviderLabel(p), string(p)))
	}
	return out
}
