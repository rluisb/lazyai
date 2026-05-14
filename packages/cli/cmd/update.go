package cmd

import (
	"fmt"
	"os"
	"path/filepath"

	"charm.land/huh/v2"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/scaffold"
	setupsvc "github.com/rluisb/lazyai/packages/cli/internal/setup"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update files to match current library versions",
	Long:  "Update managed files to match the current library versions, resolving conflicts as needed.",
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().Bool("force", false, "Overwrite local changes without prompting")
	updateCmd.Flags().Bool("no-interactive", false, "Run without interactive prompts")
	updateCmd.Flags().Bool("dry-run", false, "Show what would be changed without making changes")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	nonInteractive, _ := cmd.Flags().GetBool("no-interactive")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	targetDir, _ := os.Getwd()

	// Open the store and read existing configuration.
	database, err := openStore(targetDir)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer database.Close()

	store := db.NewStore(database)
	storeData, err := store.ReadStoreData()
	if err != nil {
		return fmt.Errorf("reading store data: %w", err)
	}

	ctx, presetLevel, err := setupsvc.BuildUpdateScaffoldContext(targetDir, setupsvc.Library{
		Dir: getLibraryDir(),
		FS:  library.GetLibraryFS(),
	}, storeData, setupsvc.UpdateOptions{
		Force:  force,
		DryRun: dryRun,
	})
	if err != nil {
		return err
	}

	if nonInteractive {
		return runUpdateNonInteractive(ctx, database, presetLevel, storeData)
	}
	return runUpdateInteractive(ctx, database, presetLevel, force, dryRun, storeData)
}

func runUpdateInteractive(
	ctx *scaffold.ScaffoldContext,
	database *db.DB,
	presetLevel types.PresetLevel,
	force, dryRun bool,
	storeData *types.StoreData,
) error {
	// Ask about force behavior if not already set.
	if !force {
		var forceConfirm bool
		forcePrompt := huh.NewConfirm().
			Title("Overwrite local changes without prompting?").
			Value(&forceConfirm)

		if err := theme.NewForm(huh.NewGroup(forcePrompt)).Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		force = forceConfirm
		ctx.Force = force
		if force {
			ctx.Strategy = types.ConflictStrategyBackupAndReplace
		}
	}

	// Confirm the update.
	var proceed bool
	confirmPrompt := huh.NewConfirm().
		Title("Proceed with update?").
		Value(&proceed)

	if err := theme.NewForm(huh.NewGroup(confirmPrompt)).Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	if !proceed {
		fmt.Println("Update cancelled.")
		return nil
	}

	// Run the scaffold pipeline.
	result, err := scaffold.ScaffoldAll(ctx)
	if err != nil {
		return fmt.Errorf("scaffold failed: %w", err)
	}

	if err := removeMigratedStrayAgentsArtifacts(ctx.TargetDir, storeData.Files); err != nil {
		return fmt.Errorf("removing migrated stray AGENTS.md artifacts: %w", err)
	}

	// Update the store with the new file records.
	if err := writeStoreFromScaffoldResult(database, ctx, presetLevel, result); err != nil {
		return fmt.Errorf("writing store data: %w", err)
	}

	// Print summary.
	fmt.Println()
	printScaffoldSummary(result, ctx)
	fmt.Println()
	return nil
}

func runUpdateNonInteractive(
	ctx *scaffold.ScaffoldContext,
	database *db.DB,
	presetLevel types.PresetLevel,
	storeData *types.StoreData,
) error {
	// Run the scaffold pipeline.
	result, err := scaffold.ScaffoldAll(ctx)
	if err != nil {
		return fmt.Errorf("scaffold failed: %w", err)
	}

	if err := removeMigratedStrayAgentsArtifacts(ctx.TargetDir, storeData.Files); err != nil {
		return fmt.Errorf("removing migrated stray AGENTS.md artifacts: %w", err)
	}

	// Update the store with the new file records.
	if err := writeStoreFromScaffoldResult(database, ctx, presetLevel, result); err != nil {
		return fmt.Errorf("writing store data: %w", err)
	}

	// Print summary.
	fmt.Println()
	printScaffoldSummary(result, ctx)
	fmt.Println()
	return nil
}

var migratedStrayAgentsPaths = []string{
	"specs/adrs/AGENTS.md",
	"specs/features/AGENTS.md",
}

func removeMigratedStrayAgentsArtifacts(targetDir string, trackedFiles []types.TrackedFile) error {
	trackedByPath := make(map[string]types.TrackedFile, len(trackedFiles))
	for _, tracked := range trackedFiles {
		trackedByPath[filepath.ToSlash(tracked.Path)] = tracked
	}

	for _, relPath := range migratedStrayAgentsPaths {
		tracked, ok := trackedByPath[relPath]
		if !ok {
			continue
		}

		absPath := filepath.Join(targetDir, filepath.FromSlash(relPath))
		if !files.FileExists(absPath) {
			continue
		}

		currentHash, err := files.FileHash(absPath)
		if err != nil {
			return err
		}
		if tracked.Hash == "" || currentHash != tracked.Hash {
			continue
		}

		if err := os.Remove(absPath); err != nil && !os.IsNotExist(err) {
			return err
		}
	}

	return nil
}
