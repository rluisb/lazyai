// Package wizard provides the interactive setup wizard for ai-setup,
// built on top of the Charm Bracelet TUI stack (bubbletea, lipgloss, huh).
//
// Phase ordering (as executed in RunWizardWithDefaults):
//
//	Phase 1 (context) → Phase 2 (features) → Phase 3 (conflicts, conditional)
//	→ Phase 5 (optional tooling) → Phase 4 (review & confirm)
//
// Title convention: each interactive sub-screen title uses the format
// "<PhaseTitle> — <n>/<N>: <StepTitle>" where <n> is the current step and
// <N> is the total steps for the user's current branch through that phase.
// Phase titles are neutral labels (e.g., "Setup Context", "Review & Confirm")
// and do not include "Phase X/N" wording.
package wizard

import (
	"fmt"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// PhaseAction indicates what the user wants to do after completing a phase.
type PhaseAction int

const (
	// PhaseContinue means proceed to the next phase.
	PhaseContinue PhaseAction = iota
	// PhaseBack means go back to the previous phase.
	PhaseBack
	// PhaseCancel means the user cancelled the wizard.
	PhaseCancel
)

// WizardMode controls how many interactive prompts are shown.
type WizardMode string

const (
	// WizardModeAuto asks the user which interactive flow to use.
	WizardModeAuto WizardMode = ""
	// WizardModeExpress uses a compact flow for faster setup.
	WizardModeExpress WizardMode = "express"
	// WizardModePersonalized uses the existing full interactive flow.
	WizardModePersonalized WizardMode = "personalized"
)

// WizardConfig holds all inputs needed to run the wizard.
type WizardConfig struct {
	// Interactive is true when the wizard should collect input via TUI prompts.
	// When false, CLI flags and defaults are used.
	Interactive bool

	// Force overwrite existing files without prompting.
	Force bool

	// DryRun shows what would be done without making changes.
	DryRun bool

	// HomeDir is the user's home directory (for global scope resolution).
	HomeDir string

	// TargetDir is the directory where setup is being run.
	TargetDir string

	// CLI overrides (populated from flags).
	CLIScope               types.SetupScope
	CLITools               []types.ToolId
	CLIName                string
	CLIPreset              types.PresetLevel
	CLIFeatures            []string
	CLIBranch              string
	CLICommit              string
	CLIWorkspaceRoot       string
	CLICliTools            []string
	CLIEnableServers       []string
	CLIMemoryPath          string
	CLIEnableObsidian      bool
	CLIObsidianVaultPath   string
	CLIEnableCodegraph     bool
	CLICodegraphDataPath   string
	CLIExistingSetupPolicy types.SetupPolicy
	CLIUseReversa          *bool
	CLIWizardMode          WizardMode

	// CLIDriveCLI, when true, asks Gemini (and future adapters) to delegate
	// scaffolding to the tool's own CLI instead of direct-write.
	CLIDriveCLI bool

	// CLILocalSecrets, when true, routes Claude Code MCP/settings writes to
	// .claude/settings.local.json (gitignored) instead of the committed
	// surfaces (.mcp.json / .claude/settings.json). Opt-in; default false.
	CLILocalSecrets bool

	// CLIOrg and CLITeam populate [YOUR_ORG] / [YOUR_TEAM] in the generated
	// AGENTS.md. Empty values become <!-- fill-in --> markers.
	CLIOrg  string
	CLITeam string

	// Project profile CLI overrides. Empty strings are safe and preserve
	// compiler fallback markers for skipped optional fields.
	CLIProjectOverview   string
	CLINamingConventions string
	CLIErrorHandling     string
	CLIApiConventions    string
	CLIImportOrder       string
	CLIProtectedBranch   string
	CLITestCommand       string
	CLILintCommand       string
	CLIBuildCommand      string
	CLICoverageThreshold int
}

// WizardResult aggregates the results from all five phases.
type WizardResult struct {
	Phase1 *Phase1Result
	Phase2 *Phase2Result
	Phase3 *Phase3Result
	Phase4 *Phase4Result
	Phase5 *Phase5Result
}

// RunWizard executes the full setup wizard.
//
// It runs phases in sequence with back-navigation support:
//   - Phase 1 → if back, re-run Phase 1
//   - Phase 2 → if back, go to Phase 1
//   - Phase 3 (only when conflicts exist) → if back, go to Phase 2
//   - Phase 5 → if back, go to Phase 3 (or Phase 2 if no conflicts)
//   - Phase 4 → if back, go to Phase 5 (or Phase 3/2 depending on conflicts)
//
// Returns the final WizardResult or an error if the wizard is cancelled
// or fails.
func RunWizard(config *WizardConfig) (*WizardResult, error) {
	return RunWizardWithDefaults(config, nil)
}

// RunWizardWithDefaults executes the setup wizard with optional pre-filled defaults.
//
// The defaults are used when re-running phases (e.g., when the user goes back).
// If defaults are nil, no pre-filling occurs.
//
// Phase ordering: 1, 2, 5 (interactive form) → 3 (conditional) → 4.
func RunWizardWithDefaults(config *WizardConfig, defaults *WizardResult) (*WizardResult, error) {
	result := &WizardResult{}
	var mode WizardMode

	if !config.Interactive {
		var err error
		var p1 *Phase1Result
		var p2 *Phase2Result
		var p5 *Phase5Result

		if defaults != nil {
			p1, _, err = RunPhase1(defaults.Phase1, true)
			if err != nil {
				return nil, err
			}
			result.Phase1 = p1

			p2, _, err = RunPhase2(result.Phase1.Scope, defaults.Phase2, true)
			if err != nil {
				return nil, err
			}
			result.Phase2 = p2

			p5, _, err = RunPhase5(defaults.Phase5, true)
			if err != nil {
				return nil, err
			}
			result.Phase5 = p5
		}
	} else {
		mode = config.CLIWizardMode
		if mode == WizardModeAuto {
			selection, err := askWizardMode()
			if err != nil {
				return nil, ErrUserCancelled
			}
			mode = selection
		}
		if mode == "" {
			mode = WizardModePersonalized
		}

		state := initWizardState(defaults)
		switch mode {
		case WizardModeExpress:
			state = initExpressWizardState(defaults)
		case WizardModePersonalized:
			// Keep the existing full interactive defaults.
		default:
			return nil, fmt.Errorf("invalid wizard mode %q", mode)
		}
		if config.CLIUseReversa != nil {
			state.AnalyzeExistingCode = *config.CLIUseReversa
		}

		form := buildInteractiveForm(state)
		if mode == WizardModeExpress {
			form = buildExpressInteractiveForm(state)
		}

		if err := form.Run(); err != nil {
			return nil, ErrUserCancelled
		}

		p1, p2, p5 := extractResults(state)
		result.Phase1 = p1
		result.Phase2 = p2
		result.Phase5 = p5
	}

	// Compute required install-consent hints (Express-only).
	var installConsents []string
	if mode == WizardModeExpress {
		installConsents = formatInstallConsents(missingInstallConsents(result.Phase1.EnableServers))
	}

	// Compute the install plan from Phase 1+2 results.
	plan, err := ComputePlan(config)
	if err != nil {
		return nil, fmt.Errorf("computing install plan: %w", err)
	}
	// Convert internal ConflictInfo to conflict.Conflict for Phase 3.
	conflicts := BuildConflictList(plan)

	// Phase 3: conflict resolution (only if conflicts exist)
	if len(conflicts) > 0 {
		phase3, action, err := RunPhase3(conflicts, !config.Interactive)
		if err != nil {
			return nil, err
		}
		if action == PhaseCancel {
			return nil, ErrUserCancelled
		}
		result.Phase3 = phase3
	} else {
		// No conflicts — create a default Phase3 result (skip all)
		result.Phase3 = &Phase3Result{
			Strategy: types.ConflictStrategySkip,
		}
	}

	// Phase 4: confirm
	phase4, action, err := RunPhase4(plan, result, !config.Interactive, installConsents)
	if err != nil {
		return nil, err
	}
	if action == PhaseBack {
		return RunWizardWithDefaults(config, result)
	}
	if action == PhaseCancel {
		return nil, ErrUserCancelled
	}
	result.Phase4 = phase4

	return result, nil
}

// ErrUserCancelled is returned when the user cancels the wizard.
var ErrUserCancelled = fmt.Errorf("user cancelled")

// BuildConflictList builds a list of conflict.Conflict structs from the plan's
// ConflictInfo list. It converts the internal representation to the format
// expected by the conflict resolution phase.
func BuildConflictList(plan *InstallPlan) []conflict.Conflict {
	var conflicts []conflict.Conflict
	for _, c := range plan.Conflicts {
		conflicts = append(conflicts, conflict.Conflict{
			Path:           c.Target,
			CurrentContent: c.ExistingContent,
			NewContent:     c.Content,
			CurrentHash:    "",
			NewHash:        "",
			TargetDir:      "",
		})
	}
	return conflicts
}
