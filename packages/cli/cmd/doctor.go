package cmd

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/frontmatter"
	"github.com/rluisb/lazyai/packages/cli/internal/manifest"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
	"github.com/rluisb/lazyai/packages/cli/internal/validation"
)

type metadataGap struct {
	Path     string   `json:"path"`
	Severity string   `json:"severity"`
	Issues   []string `json:"issues"`
}

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
			cmdLog.Warn("could not update store", "error", err)
		}
	}

	total := len(storeData.Files)
	strayAgentsFiles, err := findStrayAgentsFiles(dir)
	if err != nil {
		return err
	}
	metadataGaps, err := findMetadataGaps(dir)
	if err != nil {
		return err
	}
	metadataErrors := countMetadataSeverity(metadataGaps, "error")

	// Diagnose Claude user-scope MCP entries that still reference the legacy
	// ai-setup orchestrator binary/package — they conflict with the
	// project-scope `lazyai-orchestrator connect` LazyAI writes today (#209).
	// Detection is best-effort: a missing $HOME, an absent `~/.claude.json`,
	// or malformed JSON all surface as zero findings rather than errors.
	homeDir, _ := os.UserHomeDir()
	staleMcpEntries, err := findStaleClaudeMcpEntries(homeDir)
	if err != nil {
		// I/O error (e.g., permission denied). Warn but don't block doctor.
		cmdLog.Warn("could not scan Claude user-scope MCP config", "error", err)
		staleMcpEntries = nil
	}

	// Run enhanced health checks
	envChecks := runEnhancedHealthChecks()
	envPass, envWarn, envFail := 0, 0, 0
	for _, c := range envChecks {
		switch c.Status {
		case "pass":
			envPass++
		case "warn":
			envWarn++
		case "fail":
			envFail++
		}
	}

	issues := len(missingFiles) + len(modifiedFiles) + len(strayAgentsFiles) + metadataErrors + len(staleMcpEntries) + envWarn + envFail
	isHealthy := issues == 0

	// JSON output
	if outputJSON {
		output := map[string]any{
			"healthy":          isHealthy,
			"files":            map[string]int{"total": total, "healthy": healthy, "missing": len(missingFiles), "modified": len(modifiedFiles)},
			"missingFiles":     missingFiles,
			"modifiedFiles":    modifiedFiles,
			"strayAgentsFiles": strayAgentsFiles,
			"metadataGaps":     metadataGaps,
			"staleMcpEntries":  staleMcpEntries,
			"healthChecks":     envChecks,
			"checkedAt":        checkedAt,
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
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Dimmed)
	greenStyle := lipgloss.NewStyle().Foreground(theme.Success)
	redStyle := lipgloss.NewStyle().Foreground(theme.Error)
	yellowStyle := lipgloss.NewStyle().Foreground(theme.Warning)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dimmed)
	cyanStyle := lipgloss.NewStyle().Foreground(theme.Secondary)

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
	if len(strayAgentsFiles) > 0 {
		printKV("  Stray AGENTS.md", yellowStyle.Render(fmt.Sprintf("%d", len(strayAgentsFiles))), labelStyle, lipgloss.NewStyle())
	} else {
		printKV("  Stray AGENTS.md", dimStyle.Render("0"), labelStyle, lipgloss.NewStyle())
	}
	if len(metadataGaps) > 0 {
		printKV("  Metadata gaps", yellowStyle.Render(fmt.Sprintf("%d", len(metadataGaps))), labelStyle, lipgloss.NewStyle())
	} else {
		printKV("  Metadata gaps", dimStyle.Render("0"), labelStyle, lipgloss.NewStyle())
	}
	if len(staleMcpEntries) > 0 {
		printKV("  Stale MCP entries", yellowStyle.Render(fmt.Sprintf("%d", len(staleMcpEntries))), labelStyle, lipgloss.NewStyle())
	} else {
		printKV("  Stale MCP entries", dimStyle.Render("0"), labelStyle, lipgloss.NewStyle())
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

	if len(strayAgentsFiles) > 0 {
		fmt.Println()
		fmt.Printf("  %s %s\n", yellowStyle.Render("!"), yellowStyle.Render(fmt.Sprintf("Stray AGENTS.md files (%d):", len(strayAgentsFiles))))
		for _, f := range strayAgentsFiles {
			fmt.Printf("    %s %s\n", yellowStyle.Render("!"), f)
		}
	}

	if len(metadataGaps) > 0 {
		fmt.Println()
		fmt.Printf("  %s %s\n", yellowStyle.Render("!"), yellowStyle.Render(fmt.Sprintf("Metadata gaps (%d):", len(metadataGaps))))
		for _, gap := range metadataGaps {
			issueStyle := yellowStyle
			if gap.Severity == "error" {
				issueStyle = redStyle
			}
			fmt.Printf("    %s %s (%s)\n", issueStyle.Render("!"), gap.Path, gap.Severity)
			if verbose {
				for _, issue := range gap.Issues {
					fmt.Printf("      - %s\n", issue)
				}
			}
		}
	}

	if len(staleMcpEntries) > 0 {
		fmt.Println()
		fmt.Printf("  %s %s\n", yellowStyle.Render("!"), yellowStyle.Render(fmt.Sprintf("Stale Claude MCP entries (%d):", len(staleMcpEntries))))
		for _, entry := range staleMcpEntries {
			fmt.Printf("    %s %s\n", yellowStyle.Render("!"), entry.Reason)
			fmt.Printf("      %s %s\n", dimStyle.Render("fix:"), cyanStyle.Render(entry.Remediation))
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
				fmt.Printf("    %s\n", cyanStyle.Render("lazyai-cli update"))
			}
			if len(modifiedFiles) > 0 {
				fmt.Printf("  %s Reset modified files:\n", cyanStyle.Render("2."))
				fmt.Printf("    %s\n", cyanStyle.Render("lazyai-cli update --force"))
			}
			fmt.Printf("  %s Regenerate tool files:\n", cyanStyle.Render("3."))
			fmt.Printf("    %s\n", cyanStyle.Render("lazyai-cli compile"))
		} else {
			fmt.Println(headerStyle.Render("💡 Recommendations"))
			fmt.Println()
			printKV("  1", fmt.Sprintf("Run %s to restore missing files", cyanStyle.Render("lazyai-cli update")), labelStyle, lipgloss.NewStyle())
			printKV("  2", fmt.Sprintf("Run %s to reset modified files", cyanStyle.Render("lazyai-cli update --force")), labelStyle, lipgloss.NewStyle())
			printKV("  3", fmt.Sprintf("Run %s to regenerate tool files", cyanStyle.Render("lazyai-cli compile")), labelStyle, lipgloss.NewStyle())
		}
	}

	// Show environment health checks
	if len(envChecks) > 0 {
		fmt.Println()
		fmt.Println(headerStyle.Render("🩺 Environment Health Checks"))
		fmt.Println()
		for _, check := range envChecks {
			var emoji string
			var style lipgloss.Style
			switch check.Status {
			case "pass":
				emoji = "✅"
				style = greenStyle
			case "warn":
				emoji = "⚠️"
				style = yellowStyle
			case "fail":
				emoji = "❌"
				style = redStyle
			}
			fmt.Printf("  %s %-40s %s\n", emoji, check.Name, style.Render(check.Detail))
		}
		fmt.Println()
		fmt.Printf("  Summary: %s passed, %s warned, %s failed (of %d total)\n",
			greenStyle.Render(fmt.Sprintf("%d", envPass)),
			yellowStyle.Render(fmt.Sprintf("%d", envWarn)),
			redStyle.Render(fmt.Sprintf("%d", envFail)),
			len(envChecks))
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

func findStrayAgentsFiles(dir string) ([]string, error) {
	results := []string{}
	specsRoot := filepath.Join(dir, "specs")
	if !files.DirExists(specsRoot) {
		return results, nil
	}
	err := filepath.WalkDir(specsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if d.Name() != "AGENTS.md" {
			return nil
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			rel = path
		}
		results = append(results, filepath.ToSlash(rel))
		return nil
	})
	sort.Strings(results)
	return results, err
}

func findMetadataGaps(dir string) ([]metadataGap, error) {
	results := []metadataGap{}
	specsRoot := filepath.Join(dir, "specs")
	if !files.DirExists(specsRoot) {
		return results, nil
	}
	err := filepath.WalkDir(specsRoot, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		content, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}
		fm, _, parseErr := frontmatter.ExtractFrontmatter(content)
		if parseErr != nil {
			return parseErr
		}
		issues := validation.ValidateSpec006Metadata(fm)
		if len(issues) == 0 {
			return nil
		}
		severity := "warning"
		if isNewSpecArtifact(fm) {
			severity = "error"
		}
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			rel = path
		}
		results = append(results, metadataGap{Path: filepath.ToSlash(rel), Severity: severity, Issues: issues})
		return nil
	})
	sort.Slice(results, func(i, j int) bool {
		return results[i].Path < results[j].Path
	})
	return results, err
}

func isNewSpecArtifact(fm map[string]any) bool {
	createdAt, ok := fm["created_at"].(string)
	if !ok || createdAt == "" {
		return false
	}
	created, err := time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return false
	}
	_, hasSchemaVersion := fm["schema_version"]
	return created.Year() >= 2026 || hasSchemaVersion
}

func countMetadataSeverity(gaps []metadataGap, severity string) int {
	count := 0
	for _, gap := range gaps {
		if gap.Severity == severity {
			count++
		}
	}
	return count
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

		if err := db.RunMigrations(database); err != nil {
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
