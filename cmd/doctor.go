package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/ricardoborges-teachable/ai-setup/internal/db"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/manifest"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

var doctorCmd = &cobra.Command{
	Use:   "doctor",
	Short: "Validate setup health and check for issues",
	Long:  "Run diagnostic checks on your AI development environment and report any issues found.",
	RunE:  runDoctor,
}

func init() {
	doctorCmd.Flags().Bool("fix", false, "Print fix instructions for detected issues")
	doctorCmd.Flags().Bool("verbose", false, "Show detailed output for all files")
	doctorCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(doctorCmd)
}

func runDoctor(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	fix, _ := cmd.Flags().GetBool("fix")
	verbose, _ := cmd.Flags().GetBool("verbose")
	outputJSON, _ := cmd.Flags().GetBool("json")

	// Read store data (writable for doctor since we update file statuses)
	storeData, err := readStore(dir)
	if err != nil {
		return err
	}

	checkedAt := time.Now().UTC().Format(time.RFC3339)

	var missingFiles []string
	var modifiedFiles []string
	healthy := 0

	for i := range storeData.Files {
		record := &storeData.Files[i]
		absPath := filepath.Join(dir, record.Path)

		if !files.FileExists(absPath) {
			missingFiles = append(missingFiles, record.Path)
			record.Status = types.FileStatusMissing
			record.LastCheckedAt = checkedAt
			continue
		}

		currentHash, hashErr := files.FileHash(absPath)
		if hashErr != nil || currentHash != record.Hash {
			modifiedFiles = append(modifiedFiles, record.Path)
			record.Status = types.FileStatusModified
			record.LastCheckedAt = checkedAt
			continue
		}

		healthy++
		record.Status = types.FileStatusInstalled
		record.LastCheckedAt = checkedAt
	}

	// Write updated store data back
	if err := writeStoreData(dir, storeData); err != nil {
		// Non-fatal: status update failed, but we can still report
		if verbose {
			fmt.Fprintf(os.Stderr, "Warning: could not update store: %v\n", err)
		}
	}

	total := len(storeData.Files)
	issues := len(missingFiles) + len(modifiedFiles)
	isHealthy := issues == 0

	// JSON output
	if outputJSON {
		output := map[string]any{
			"healthy":       isHealthy,
			"files":         map[string]int{"total": total, "healthy": healthy, "missing": len(missingFiles), "modified": len(modifiedFiles)},
			"missingFiles":  missingFiles,
			"modifiedFiles": modifiedFiles,
			"checkedAt":     checkedAt,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(output)
		if !isHealthy {
			return fmt.Errorf("doctor found %d integrity issue(s)", issues)
		}
		return nil
	}

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00CFC5"))

	fmt.Println()
	fmt.Println(headerStyle.Render("🩺 Integrity Check"))
	fmt.Println()

	statusEmoji := "✅"
	statusText := greenStyle.Render("All files healthy")
	if !isHealthy {
		if issues < 5 {
			statusEmoji = "⚠️"
		} else {
			statusEmoji = "❌"
		}
		statusText = yellowStyle.Render(fmt.Sprintf("%d issue(s) found", issues))
	}

	printKV("  Status", fmt.Sprintf("%s %s", statusEmoji, statusText), labelStyle, lipgloss.NewStyle())
	printKV("  Health", formatHealthBar(healthy, total), labelStyle, lipgloss.NewStyle())
	printKV("  Total files", fmt.Sprintf("%d", total), labelStyle, lipgloss.NewStyle())
	printKV("  Healthy", greenStyle.Render(fmt.Sprintf("%d", healthy)), labelStyle, lipgloss.NewStyle())

	if len(missingFiles) > 0 {
		printKV("  Missing", redStyle.Render(fmt.Sprintf("%d", len(missingFiles))), labelStyle, lipgloss.NewStyle())
	} else {
		printKV("  Missing", dimStyle.Render("0"), labelStyle, lipgloss.NewStyle())
	}
	if len(modifiedFiles) > 0 {
		printKV("  Modified", yellowStyle.Render(fmt.Sprintf("%d", len(modifiedFiles))), labelStyle, lipgloss.NewStyle())
	} else {
		printKV("  Modified", dimStyle.Render("0"), labelStyle, lipgloss.NewStyle())
	}

	// Show missing files
	if len(missingFiles) > 0 {
		fmt.Println()
		fmt.Printf("  %s %s\n", redStyle.Render("✗"), redStyle.Render(fmt.Sprintf("Missing files (%d):", len(missingFiles))))
		display := missingFiles
		if !verbose && len(display) > 5 {
			display = display[:5]
		}
		for _, f := range display {
			fmt.Printf("    %s %s\n", redStyle.Render("✗"), f)
		}
		if !verbose && len(missingFiles) > 5 {
			fmt.Printf("    %s\n", dimStyle.Render(fmt.Sprintf("... and %d more (use --verbose to see all)", len(missingFiles)-5)))
		}
	}

	// Show modified files
	if len(modifiedFiles) > 0 {
		fmt.Println()
		fmt.Printf("  %s %s\n", yellowStyle.Render("~"), yellowStyle.Render(fmt.Sprintf("Modified files (%d):", len(modifiedFiles))))
		display := modifiedFiles
		if !verbose && len(display) > 5 {
			display = display[:5]
		}
		for _, f := range display {
			fmt.Printf("    %s %s\n", yellowStyle.Render("~"), f)
		}
		if !verbose && len(modifiedFiles) > 5 {
			fmt.Printf("    %s\n", dimStyle.Render(fmt.Sprintf("... and %d more (use --verbose to see all)", len(modifiedFiles)-5)))
		}
	}

	// Fix recommendations or instructions
	if !isHealthy {
		fmt.Println()
		if fix {
			fmt.Println(headerStyle.Render("💡 Fix Instructions"))
			fmt.Println()
			if len(missingFiles) > 0 {
				fmt.Printf("  %s Restore missing files:\n", cyanStyle.Render("1."))
				fmt.Printf("    %s\n", cyanStyle.Render("ai-setup update"))
			}
			if len(modifiedFiles) > 0 {
				fmt.Printf("  %s Reset modified files:\n", cyanStyle.Render("2."))
				fmt.Printf("    %s\n", cyanStyle.Render("ai-setup update --force"))
			}
			fmt.Printf("  %s Regenerate tool files:\n", cyanStyle.Render("3."))
			fmt.Printf("    %s\n", cyanStyle.Render("ai-setup compile"))
		} else {
			fmt.Println(headerStyle.Render("💡 Recommendations"))
			fmt.Println()
			printKV("  1", fmt.Sprintf("Run %s to restore missing files", cyanStyle.Render("ai-setup update")), labelStyle, lipgloss.NewStyle())
			printKV("  2", fmt.Sprintf("Run %s to reset modified files", cyanStyle.Render("ai-setup update --force")), labelStyle, lipgloss.NewStyle())
			printKV("  3", fmt.Sprintf("Run %s to regenerate tool files", cyanStyle.Render("ai-setup compile")), labelStyle, lipgloss.NewStyle())
		}
	}

	fmt.Println()
	if isHealthy {
		fmt.Printf("  %s\n", greenStyle.Render("✓ Setup integrity verified"))
	} else {
		fmt.Printf("  %s\n", yellowStyle.Render(fmt.Sprintf("⚠ Setup has %d integrity issue(s)", issues)))
		return fmt.Errorf("doctor found %d integrity issue(s)", issues)
	}

	return nil
}

// writeStoreData writes store data back to the appropriate backend.
func writeStoreData(dir string, data *types.StoreData) error {
	// Try SQLite first if db exists
	dbPath := db.DefaultDBPath(dir)
	if files.FileExists(dbPath) {
		database, err := db.Open(dbPath)
		if err != nil {
			// Fall back to manifest
			return writeManifestStore(dir, data)
		}
		defer database.Close()

		if err := db.RunMigrations(database.DB); err != nil {
			return writeManifestStore(dir, data)
		}

		store := db.NewStore(database)
		return store.WriteStoreData(data)
	}

	return writeManifestStore(dir, data)
}

func writeManifestStore(dir string, data *types.StoreData) error {
	if manifest.ManifestExists(dir) {
		return manifest.WriteManifest(dir, data)
	}
	// If neither store exists, we can't write
	return nil
}

// hasPrefix checks if a string has one of the given prefixes.
func hasPrefix(s string, prefixes []string) bool {
	normalized := filepath.ToSlash(s)
	for _, p := range prefixes {
		if strings.HasPrefix(normalized, p) {
			return true
		}
	}
	return false
}
