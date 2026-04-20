// Package wizard provides the interactive setup wizard for ai-setup,
// built on top of the Charm Bracelet TUI stack (bubbletea, lipgloss, huh).
//
// Phase ordering (as executed in RunWizardWithDefaults):
//   Phase 1 (context) → Phase 2 (features) → Phase 3 (conflicts, conditional)
//   → Phase 5 (optional tooling) → Phase 4 (review & confirm)
//
// Title convention: each interactive sub-screen title uses the format
// "<PhaseTitle> — <n>/<N>: <StepTitle>" where <n> is the current step and
// <N> is the total steps for the user's current branch through that phase.
// Phase titles are neutral labels (e.g., "Setup Context", "Review & Confirm")
// and do not include "Phase X/N" wording.
package wizard

import (
	"fmt"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
	CLIScope             types.SetupScope
	CLITools             []types.ToolId
	CLIName              string
	CLIPreset            types.PresetLevel
	CLIFeatures          []string
	CLIBranch            string
	CLICommit            string
	CLICliTools          []string
	CLIEnableServers     []string
	CLIMemoryPath        string
	CLIEnableObsidian    bool
	CLIObsidianVaultPath string
	CLIEnableQmd         bool
	CLIQmdIndexPath      string
	CLIEnableCodegraph   bool
	CLICodegraphDataPath string

	// CLIDriveCLI, when true, asks Gemini (and future adapters) to delegate
	// scaffolding to the tool's own CLI instead of direct-write.
	CLIDriveCLI bool

	// CLILocalSecrets, when true, routes Claude Code MCP/settings writes to
	// .claude/settings.local.json (gitignored) instead of the committed
	// surfaces (.mcp.json / .claude/settings.json). Opt-in; default false.
	CLILocalSecrets bool

	// CLIOrg and CLITeam populate [YOUR_ORG] / [YOUR_TEAM] in the generated
	// CLAUDE.md. Empty values become <!-- fill-in --> markers.
	CLIOrg  string
	CLITeam string
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
// Phase ordering: 1 → 2 → 3 (conditional) → 5 → 4.
func RunWizardWithDefaults(config *WizardConfig, defaults *WizardResult) (*WizardResult, error) {
	result := &WizardResult{}

	// Seed result from defaults.
	var phase2Defaults *Phase2Result
	if defaults != nil {
		if defaults.Phase1 != nil {
			result.Phase1 = defaults.Phase1
		}
		phase2Defaults = defaults.Phase2
	}

	// Run the Phase 1-2 loop. Returns when both phases are complete.
	result, err := runPhase12Loop(config, result, phase2Defaults)
	if err != nil {
		return nil, err
	}

	// --- Phases 3-5 with outer loop for back navigation ---
	for {
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
			if action == PhaseBack {
				// Go back to Phase 2 — re-run Phase 1-2 loop
				result, err = runPhase12Loop(config, result, result.Phase2)
				if err != nil {
					return nil, err
				}
				continue
			}
		} else {
			// No conflicts — create a default Phase3 result (skip all)
			result.Phase3 = &Phase3Result{
				Strategy: types.ConflictStrategySkip,
			}
		}

		opencodeSelected := toolsContain(result.Phase1.Tools, types.ToolIdOpenCode)
		phase5, action, err := RunPhase5(result.Phase5, !config.Interactive, opencodeSelected)
		if err != nil {
			return nil, err
		}
		if action == PhaseCancel {
			return nil, ErrUserCancelled
		}
		result.Phase5 = phase5
		if action == PhaseBack {
			if len(conflicts) > 0 {
				continue
			}
			result, err = runPhase12Loop(config, result, result.Phase2)
			if err != nil {
				return nil, err
			}
			continue
		}

		// Phase 4: confirm
		phase4, action, err := RunPhase4(plan, !config.Interactive)
		if err != nil {
			return nil, err
		}
		if action == PhaseCancel {
			return nil, ErrUserCancelled
		}
		result.Phase4 = phase4
		if action == PhaseBack {
			// Go back: if there were conflicts, re-run Phase 3; otherwise back to Phase 2
			if len(conflicts) > 0 {
				// Re-run this loop from the top (will re-compute plan)
				continue
			}
			// Go back to Phase 2 — re-run Phase 1-2 loop
			result, err = runPhase12Loop(config, result, result.Phase2)
			if err != nil {
				return nil, err
			}
			continue
		}

		// Confirmed — break out
		break
	}

	return result, nil
}

// runPhase12Loop runs phases 1 and 2 with back-navigation between them.
// If defaults are provided, they are used to pre-fill prompts on re-runs.
func runPhase12Loop(config *WizardConfig, result *WizardResult, phase2Defaults *Phase2Result) (*WizardResult, error) {
	currentPhase := 1

	for currentPhase >= 1 && currentPhase <= 2 {
		switch currentPhase {
		case 1:
			phase1Defaults := result.Phase1
			phase1, action, err := RunPhase1(phase1Defaults, !config.Interactive)
			if err != nil {
				return nil, err
			}
			if action == PhaseCancel {
				return nil, ErrUserCancelled
			}
			result.Phase1 = phase1
			if action == PhaseBack {
				// Can't go back from Phase 1 — just continue
				continue
			}
			currentPhase = 2

		case 2:
			phase2, action, err := RunPhase2(
				result.Phase1.Scope,
				phase2Defaults,
				!config.Interactive,
			)
			if err != nil {
				return nil, err
			}
			if action == PhaseCancel {
				return nil, ErrUserCancelled
			}
			result.Phase2 = phase2
			if action == PhaseBack {
				currentPhase = 1
				continue
			}
			currentPhase = 3
		}
	}

	return result, nil
}

// ErrUserCancelled is returned when the user cancels the wizard.
var ErrUserCancelled = fmt.Errorf("user cancelled")

func toolsContain(tools []types.ToolId, target types.ToolId) bool {
	for _, t := range tools {
		if t == target {
			return true
		}
	}
	return false
}

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
