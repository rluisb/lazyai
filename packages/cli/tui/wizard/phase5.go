package wizard

import (
	"context"
	"fmt"

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
	EnableCodegraph   bool
	CodegraphDataPath string
	// OpenCodeProviders are the provider IDs (e.g., "openai", "ollama-cloud")
	// the user has authenticated and chosen to expose to OpenCode-side agents
	// at install time. Empty when OpenCode isn't selected; otherwise
	// populated from auth.DetectAll filtered against the OpenCode catalog's
	// deny rules (Anthropic excluded).
	OpenCodeProviders []string
}

// RunPhase5 runs the optional tooling phase.
func RunPhase5(defaults *Phase5Result, nonInteractive bool) (*Phase5Result, PhaseAction, error) {
	if nonInteractive {
		return runPhase5NonInteractive(defaults)
	}
	return runPhase5Interactive(defaults)
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
		result.EnableCodegraph,
		result.CodegraphDataPath,
	), PhaseContinue, nil
}

func runPhase5Interactive(defaults *Phase5Result) (*Phase5Result, PhaseAction, error) {
	state := defaultPhase5Result()
	if defaults != nil {
		state = *defaults
	}

	currentStep := 1
	maxStep := 2 // memory path + providers
	for currentStep >= 1 && currentStep <= maxStep {
		switch currentStep {
		case 1:
			memoryPath, action, err := askMemoryPath(state.MemoryPath, phase5MemoryPathStepInfo())
			if err != nil {
				return nil, action, err
			}
			state.MemoryPath = memoryPath
			currentStep++
		case 2:
			providers, action, err := askOpenCodeProviders(state.OpenCodeProviders)
			if err != nil {
				return nil, action, err
			}
			if action == PhaseBack {
				currentStep = 1
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
		state.EnableCodegraph,
		state.CodegraphDataPath,
	)
	result.OpenCodeProviders = state.OpenCodeProviders
	return result, PhaseContinue, nil
}

func buildPhase5Result(memoryPath string, enableObsidian bool, obsidianVaultPath string, enableCodegraph bool, codegraphDataPath string) *Phase5Result {
	if memoryPath == "" {
		memoryPath = ".specify/memory"
	}
	if enableCodegraph && codegraphDataPath == "" {
		codegraphDataPath = ".codegraph/"
	}

	return &Phase5Result{
		MemoryPath:        memoryPath,
		EnableObsidian:    enableObsidian,
		ObsidianVaultPath: obsidianVaultPath,
		EnableCodegraph:   enableCodegraph,
		CodegraphDataPath: codegraphDataPath,
	}
}

func defaultPhase5Result() Phase5Result {
	return Phase5Result{
		MemoryPath:        ".specify/memory",
		EnableObsidian:    true,
		EnableCodegraph:   true,
		CodegraphDataPath: ".codegraph/",
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

func phase5MemoryPathStepInfo() phase5StepInfo {
	return phase5StepInfo{Current: 1, Total: 2, StepTitle: "Memory Path"}
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
		Options(options...).
		Value(&selected)
	if err := theme.NewForm(huh.NewGroup(multiSelectFooterDescription(field, func() string {
		return multiSelectHoverDescription(field, opencodeProviderDescriptions(eligible), defaultHoverHint)
	}))).Run(); err != nil {
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

func opencodeProviderDescriptions(providers []auth.ProviderID) map[string]string {
	descriptions := make(map[string]string, len(providers))
	for _, p := range providers {
		descriptions[string(p)] = fmt.Sprintf("Expose authenticated %s models to OpenCode config resolution.", opencodeProviderLabel(p))
	}
	return descriptions
}
