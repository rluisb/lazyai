package cmd

import (
	"fmt"
	"os"

	"github.com/rluisb/lazyai/packages/cli/internal/migration"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate from a previous LazyAI/ai-setup version",
	Long:  "Migrate your existing LazyAI or ai-setup configuration from a previous version to the current format.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runMigrate,
}

func init() {
	migrateCmd.Flags().String("from", "", "Source version to migrate from")
	migrateCmd.Flags().Bool("no-interactive", false, "Run without interactive prompts")
	migrateCmd.Flags().Bool("preview", false, "Preview changes without executing")
	migrateCmd.Flags().String("strategy", "smart", "Merge strategy: smart, preserve, replace, append")
	migrateCmd.Flags().Bool("verbose", false, "Show detailed output")
	migrateCmd.Flags().Bool("skip-backup", false, "Skip creating backup")
	rootCmd.AddCommand(migrateCmd)
}

func runMigrate(cmd *cobra.Command, args []string) error {
	sourcePath, _ := os.Getwd()
	if len(args) > 0 {
		sourcePath = args[0]
	}
	targetPath, _ := os.Getwd()

	nonInteractive, _ := cmd.Flags().GetBool("no-interactive")
	preview, _ := cmd.Flags().GetBool("preview")
	strategyStr, _ := cmd.Flags().GetString("strategy")
	verbose, _ := cmd.Flags().GetBool("verbose")
	skipBackup, _ := cmd.Flags().GetBool("skip-backup")

	ctx := &migration.MigrationContext{
		SourcePath: sourcePath,
		TargetPath: targetPath,
		Options: migration.MigrationOptions{
			Preview:       preview,
			MergeStrategy: migration.MergeStrategy(strategyStr),
			Verbose:       verbose,
			SkipBackup:    skipBackup,
			Interactive:   !nonInteractive,
		},
	}

	detections, err := migration.DetectSetup(sourcePath)
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}
	if len(detections) == 0 {
		return fmt.Errorf("no existing AI setup detected in %s", sourcePath)
	}

	parsedSetups, err := migration.ParseDetectedSetups(ctx, detections)
	if err != nil {
		return fmt.Errorf("parse failed: %w", err)
	}

	plan, err := migration.BuildCanonicalPlan(ctx, detections, parsedSetups)
	if err != nil {
		return fmt.Errorf("failed to build migration plan: %w", err)
	}

	fmt.Println(migration.FormatPlan(plan))
	fmt.Println()

	if preview {
		fmt.Println("Preview mode — no changes were made.")
		return nil
	}

	if ctx.Options.Interactive {
		fmt.Printf("Proceed with migration? (y/N): ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" && answer != "yes" {
			fmt.Println("Migration cancelled.")
			return nil
		}
	}

	result, err := migration.ExecuteToCanonical(ctx, plan, parsedSetups)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}

	printImportResult(result)
	if !result.Success {
		return fmt.Errorf("migration completed with errors")
	}

	return nil
}
