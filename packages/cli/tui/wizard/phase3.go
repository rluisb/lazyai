package wizard

import (
	"fmt"

	"charm.land/huh/v2"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ConflictResolution records the user's decision for a single conflict.
type ConflictResolution struct {
	Path   string
	Action ReviewAction
}

// Phase3Result holds the outcome of conflict resolution.
type Phase3Result struct {
	Strategy         types.ConflictStrategy
	Resolutions      []ConflictResolution
	PerFileOverrides map[string]types.ConflictStrategy
}

// RunPhase3 runs the conflict resolution phase.
//
// In non-interactive mode, it uses the default strategy:
//   - With --force: backup-and-replace (accept library versions)
//   - Without --force: skip (keep existing files)
//
// In interactive mode, it presents the conflict strategy choice and
// (for "align") the side-by-side diff viewer for per-file resolution.
func RunPhase3(conflicts []conflict.Conflict, nonInteractive bool) (*Phase3Result, PhaseAction, error) {
	if nonInteractive {
		return runPhase3NonInteractive(conflicts)
	}
	return runPhase3Interactive(conflicts, NewDiffReviewClient())
}

func runPhase3NonInteractive(conflicts []conflict.Conflict) (*Phase3Result, PhaseAction, error) {
	// Non-interactive: skip conflicting files by default.
	// The --force flag is handled at the caller level.
	return &Phase3Result{
		Strategy:         types.ConflictStrategySkip,
		Resolutions:      nil,
		PerFileOverrides: nil,
	}, PhaseContinue, nil
}

func runPhase3Interactive(conflicts []conflict.Conflict, reviewer DiffReviewClient) (*Phase3Result, PhaseAction, error) {
	if len(conflicts) == 0 {
		return &Phase3Result{
			Strategy:         types.ConflictStrategySkip,
			Resolutions:      nil,
			PerFileOverrides: nil,
		}, PhaseContinue, nil
	}

	// Ask for global strategy first.
	var strategyValue string
	strategySelect := huh.NewSelect[string]().
		Title(fmt.Sprintf("Found %d conflicting file(s). How should conflicts be handled?", len(conflicts))).
		Options(
			huh.NewOption("Align — Review each conflict and choose per file", "align"),
			huh.NewOption("Backup and Replace — Create backups and overwrite all conflicts", "backup-and-replace"),
			huh.NewOption("Skip — Keep existing files and skip conflicts", "skip"),
		).
		Value(&strategyValue)
	strategySelect.DescriptionFunc(func() string {
		return selectHoverDescription(strategySelect, conflictStrategyDescriptions, defaultHoverHint)
	}, strategySelect)

	form := theme.NewForm(huh.NewGroup(strategySelect).Title("Conflict Resolution"))
	if err := form.Run(); err != nil {
		return nil, PhaseCancel, fmt.Errorf("phase 3 cancelled: %w", err)
	}

	result := &Phase3Result{
		Strategy:         types.ConflictStrategy(strategyValue),
		PerFileOverrides: make(map[string]types.ConflictStrategy),
	}

	// If "align", show the interactive diff viewer for each conflict.
	if result.Strategy == types.ConflictStrategyAlign {
		resolutions := make([]ConflictResolution, 0, len(conflicts))
		if reviewer == nil {
			reviewer = InlineDiffReviewer{}
		}

		diffResolutions, err := reviewer.RunReview(conflicts)
		if err != nil {
			return nil, PhaseCancel, err
		}
		for _, res := range diffResolutions {
			resolutions = append(resolutions, ConflictResolution{
				Path:   res.Path,
				Action: res.Action,
			})
			result.PerFileOverrides[res.Path] = ConflictStrategyForReviewAction(res.Action)
		}

		result.Resolutions = resolutions
	}

	return result, PhaseContinue, nil
}
