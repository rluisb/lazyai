package wizard

import (
	"fmt"

	"charm.land/huh/v2"
)

type Phase5Result struct {
	MemoryPath        string
	EnableObsidian    bool
	ObsidianVaultPath string
	EnableQmd         bool
	QmdIndexPath      string
	EnableCodegraph   bool
	CodegraphDataPath string
}

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
	if result.MemoryPath == "" {
		result.MemoryPath = "specs/memory"
	}
	if result.EnableQmd && result.QmdIndexPath == "" {
		result.QmdIndexPath = ".qmd-index"
	}
	if result.EnableCodegraph && result.CodegraphDataPath == "" {
		result.CodegraphDataPath = ".codegraph"
	}
	return &result, PhaseContinue, nil
}

func runPhase5Interactive(defaults *Phase5Result) (*Phase5Result, PhaseAction, error) {
	result := defaultPhase5Result()
	if defaults != nil {
		result = *defaults
	}

	memoryPath := result.MemoryPath
	enableObsidian := result.EnableObsidian
	obsidianVaultPath := result.ObsidianVaultPath
	enableQmd := result.EnableQmd
	qmdIndexPath := result.QmdIndexPath
	enableCodegraph := result.EnableCodegraph
	codegraphDataPath := result.CodegraphDataPath

	group := huh.NewGroup(
		huh.NewInput().Title("Memory path").Description("Project-local default for bootstrap and housekeeping.").Placeholder("specs/memory").Value(&memoryPath),
		huh.NewConfirm().Title("Enable Obsidian integration?").Description("Read-only discovery only by default; future config writes remain explicit.").Value(&enableObsidian),
		huh.NewInput().Title("Obsidian vault path (optional)").Value(&obsidianVaultPath),
		huh.NewConfirm().Title("Enable qmd markdown retrieval?").Description("Read-only retrieval allowed; sync/index writes remain approval-gated.").Value(&enableQmd),
		huh.NewInput().Title("qmd index path").Placeholder(".qmd-index").Value(&qmdIndexPath),
		huh.NewConfirm().Title("Enable codegraph analysis?").Description("Read-only drift checks allowed; sync/index writes remain approval-gated.").Value(&enableCodegraph),
		huh.NewInput().Title("Codegraph data path").Placeholder(".codegraph").Value(&codegraphDataPath),
	).Title("Phase 5/5: Optional Tooling")

	if err := huh.NewForm(group).Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 5 cancelled: %w", err)
	}

	if memoryPath == "" {
		memoryPath = "specs/memory"
	}
	if enableQmd && qmdIndexPath == "" {
		qmdIndexPath = ".qmd-index"
	}
	if enableCodegraph && codegraphDataPath == "" {
		codegraphDataPath = ".codegraph"
	}

	return &Phase5Result{
		MemoryPath:        memoryPath,
		EnableObsidian:    enableObsidian,
		ObsidianVaultPath: obsidianVaultPath,
		EnableQmd:         enableQmd,
		QmdIndexPath:      qmdIndexPath,
		EnableCodegraph:   enableCodegraph,
		CodegraphDataPath: codegraphDataPath,
	}, PhaseContinue, nil
}

func defaultPhase5Result() Phase5Result {
	return Phase5Result{
		MemoryPath:        "specs/memory",
		QmdIndexPath:      ".qmd-index",
		CodegraphDataPath: ".codegraph",
	}
}
