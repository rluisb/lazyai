package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/migration"
)

var importCmd = &cobra.Command{
	Use:   "import [source]",
	Short: "Import configuration from another AI tool setup",
	Long: `Import configuration from another AI tool setup (e.g., OpenCode, Claude Code, Copilot)
into LazyAI format. The source can be a path to a directory containing an existing setup.`,
	Args: cobra.MaximumNArgs(1),
	RunE: runImport,
}

func init() {
	importCmd.Flags().String("tool", "", "Source tool to import from (auto-detected if omitted)")
	importCmd.Flags().Bool("no-interactive", false, "Run without interactive prompts")
	importCmd.Flags().Bool("preview", false, "Preview changes without executing")
	importCmd.Flags().String("strategy", "smart", "Merge strategy: smart, preserve, replace, append")
	importCmd.Flags().Bool("verbose", false, "Show detailed output")
	importCmd.Flags().Bool("skip-backup", false, "Skip creating backup")
	rootCmd.AddCommand(importCmd)
	importCmd.GroupID = "scaffold"
}

func runImport(cmd *cobra.Command, args []string) error {
	// Resolve source path.
	sourcePath, _ := os.Getwd()
	if len(args) > 0 {
		sourcePath = args[0]
	}

	targetPath, _ := os.Getwd()

	// Parse flags.
	toolFlag, _ := cmd.Flags().GetString("tool")
	nonInteractive, _ := cmd.Flags().GetBool("no-interactive")
	preview, _ := cmd.Flags().GetBool("preview")
	strategyStr, _ := cmd.Flags().GetString("strategy")
	verbose, _ := cmd.Flags().GetBool("verbose")
	skipBackup, _ := cmd.Flags().GetBool("skip-backup")
	if err := validateToolFlag(toolFlag); err != nil {
		return err
	}

	strategy := migration.MergeStrategy(strategyStr)
	if strategy == "" {
		strategy = migration.MergeStrategySmart
	}

	ctx := &migration.MigrationContext{
		SourcePath: sourcePath,
		TargetPath: targetPath,
		Options: migration.MigrationOptions{
			Preview:       preview,
			MergeStrategy: strategy,
			Verbose:       verbose,
			SkipBackup:    skipBackup,
			Interactive:   !nonInteractive,
		},
	}

	// Step 1: Detect existing setups.
	detections, err := migration.DetectSetup(sourcePath)
	if err != nil {
		return fmt.Errorf("detection failed: %w", err)
	}

	if len(detections) == 0 {
		if toolFlag != "" {
			return fmt.Errorf("no %s setup detected in %s", toolFlag, sourcePath)
		}
		return fmt.Errorf("no existing AI setup detected in %s", sourcePath)
	}

	// Filter by --tool flag if specified.
	if toolFlag != "" {
		var filtered []migration.DetectionResult
		for _, d := range detections {
			if d.AdapterID == toolFlag {
				filtered = append(filtered, d)
			}
		}
		if len(filtered) == 0 {
			return fmt.Errorf("no %s setup detected in %s", toolFlag, sourcePath)
		}
		detections = filtered
	}

	// Show detected adapters.
	fmt.Println("Detected AI setups:")
	for _, d := range detections {
		fmt.Printf("  • %s (detector confidence: %.0f%%, %d files)\n", d.AdapterName, d.Confidence*100, len(d.Files))
	}
	fmt.Println()

	// Step 2: Parse canonical candidates.
	parsedDetections := canonicalImportDetections(detections)
	parsedSetups := make([]migration.ParsedSetup, 0, len(parsedDetections))
	if len(parsedDetections) > 0 {
		parsedSetups, err = migration.ParseDetectedSetups(ctx, parsedDetections)
		if err != nil {
			return fmt.Errorf("failed to parse supported setups: %w", err)
		}
	}

	plan, err := migration.BuildCanonicalPlan(ctx, detections, parsedSetups)
	if err != nil {
		return fmt.Errorf("failed to build migration plan: %w", err)
	}

	// Step 3: Show plan.
	fmt.Println(migration.FormatPlan(plan))
	rawPaths, err := planRawPreservation(ctx.SourcePath, ctx.TargetPath, detections, true)
	if err != nil {
		return fmt.Errorf("failed to plan raw preservation: %w", err)
	}
	if len(rawPaths) > 0 {
		fmt.Printf("Preserve raw native config: %d file(s) under .ai/adapters/*/raw/\n", len(rawPaths))
	}
	fmt.Printf("Migration report: %s\n\n", migrationReportRelPath)

	if preview {
		fmt.Println("Preview mode — no changes were made.")
		return nil
	}

	// Step 4: Confirm execution (interactive mode).
	if ctx.Options.Interactive {
		fmt.Printf("Proceed with import? (y/N): ")
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" && answer != "yes" {
			fmt.Println("Import cancelled.")
			return nil
		}
	}

	// Step 5: Execute canonical import + preservation.
	result, err := migration.ExecuteToCanonical(ctx, plan, parsedSetups)
	if err != nil {
		return fmt.Errorf("import failed: %w", err)
	}
	if err := scaffoldImportManifest(targetPath, detections, ctx.Options.Preview); err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("manifest bootstrap failed: %v", err))
	}
	if _, err := planRawPreservation(ctx.SourcePath, ctx.TargetPath, detections, false); err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("raw preservation failed: %v", err))
	}
	reportPath, err := writeMigrationReport(ctx.SourcePath, ctx.TargetPath, detections, result, rawPaths, ctx.Options.Preview)
	if err != nil {
		result.Success = false
		result.Errors = append(result.Errors, fmt.Sprintf("report write failed: %v", err))
	}

	// Step 6: Show results.
	printImportResult(result, reportPath, len(rawPaths))

	return nil
}

func printImportResult(result *migration.MigrationResult, reportPath string, rawPreserved int) {
	fmt.Println()
	fmt.Println("Import complete!")
	fmt.Printf("  Files created:   %d\n", result.Stats.FilesCreated)
	fmt.Printf("  Files modified:  %d\n", result.Stats.FilesModified)
	fmt.Printf("  Files backed up: %d\n", result.Stats.FilesBackedUp)
	fmt.Printf("  Files skipped:   %d\n", result.Stats.FilesSkipped)
	fmt.Printf("  Raw preserved:   %d\n", rawPreserved)

	if result.BackupPath != "" {
		fmt.Printf("  Backup location: %s\n", result.BackupPath)
	}
	if reportPath != "" {
		fmt.Printf("  Report:          %s\n", reportPath)
	}

	if len(result.Errors) > 0 {
		fmt.Println()
		fmt.Println("Errors:")
		for _, e := range result.Errors {
			fmt.Printf("  ! %s\n", e)
		}
	}

	if len(result.Warnings) > 0 {
		fmt.Println()
		fmt.Println("Warnings:")
		for _, w := range result.Warnings {
			fmt.Printf("  ⚠ %s\n", w)
		}
	}

	if result.Success {
		fmt.Println()
		fmt.Println("✓ Import succeeded.")
	} else {
		fmt.Println()
		fmt.Println("✗ Import completed with errors.")
	}
}
