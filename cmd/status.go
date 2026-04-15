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
	aierror "github.com/ricardoborges-teachable/ai-setup/internal/error"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/manifest"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show the current setup state",
	Long:  "Display the current state of your AI development environment, including installed tools, agents, and skills.",
	RunE:  runStatus,
}

func init() {
	statusCmd.Flags().String("dir", "", "Target directory (default: current directory)")
	statusCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(statusCmd)
}

func runStatus(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	outputJSON, _ := cmd.Flags().GetBool("json")

	// Try SQLite store first, fall back to JSON manifest
	storeData, err := readStore(dir)
	if err != nil {
		return err
	}

	// Check file health
	healthy, missing, modified := 0, 0, 0
	for _, record := range storeData.Files {
		absPath := filepath.Join(dir, record.Path)
		if !files.FileExists(absPath) {
			missing++
			continue
		}
		currentHash, hashErr := files.FileHash(absPath)
		if hashErr != nil || currentHash != record.Hash {
			modified++
			continue
		}
		healthy++
	}
	total := len(storeData.Files)

	projectName := storeData.Config.ProjectName
	if projectName == "" {
		projectName = storeData.Config.WorkspaceName
	}
	if projectName == "" {
		projectName = "(unnamed)"
	}

	tools := storeData.Config.Tools
	if len(tools) == 0 {
		tools = []types.ToolId{"(none)"}
	}

	planningDir := storeData.Config.PlanningDir
	if planningDir == "" {
		planningDir = ".planning"
	}

	// JSON output
	if outputJSON {
		output := map[string]any{
			"scope":       storeData.Config.SetupScope,
			"projectName": projectName,
			"tools":       storeData.Config.Tools,
			"planningDir": planningDir,
			"files":       map[string]int{"total": total, "healthy": healthy, "missing": missing, "modified": modified},
			"lastInit":    storeData.Meta.InstalledAt,
			"cliVersion":  storeData.Meta.CLIVersion,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Styled output
	boldStyle := lipgloss.NewStyle().Bold(true)
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))
	valueStyle := lipgloss.NewStyle()
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))

	fmt.Println()
	fmt.Println(boldStyle.Render("ai-setup status"))
	fmt.Println()

	// Project info
	fmt.Println(headerStyle.Render("📦 Project"))
	printKV("  Name", projectName, labelStyle, valueStyle)
	printKV("  Scope", string(storeData.Config.SetupScope), labelStyle, valueStyle)
	printKV("  Planning dir", planningDir, labelStyle, valueStyle)
	fmt.Println()

	// Tools
	fmt.Println(headerStyle.Render("🔧 Tools"))
	printKV("  Installed", toolList(tools), labelStyle, valueStyle)
	fmt.Println()

	// Features
	features := formatFeatures(storeData.Selections.Features)
	if len(features) > 0 {
		fmt.Println(headerStyle.Render("⚡ Features"))
		printKV("  Enabled", strings.Join(features, ", "), labelStyle, valueStyle)
		fmt.Println()
	}

	// Git conventions
	if storeData.Selections.GitConventions != nil {
		gc := storeData.Selections.GitConventions
		fmt.Println(headerStyle.Render("📝 Git Conventions"))
		branchPattern := gc.BranchPattern
		if branchPattern == "" {
			branchPattern = "{type}/{ticket}-{description}"
		}
		commitPattern := gc.CommitPattern
		if commitPattern == "" {
			commitPattern = "{type}({scope}): {description}"
		}
		typesStr := "(defaults)"
		if len(gc.Types) > 0 {
			typesStr = strings.Join(gc.Types, ", ")
		}
		printKV("  Branch", branchPattern, labelStyle, valueStyle)
		printKV("  Commit", commitPattern, labelStyle, valueStyle)
		printKV("  Require ticket", boolStr(gc.RequireTicket), labelStyle, valueStyle)
		printKV("  Types", typesStr, labelStyle, valueStyle)
		fmt.Println()
	}

	// File health
	fmt.Println(headerStyle.Render("📁 File Health"))
	healthStatus := greenStyle.Render("✓ All files healthy")
	if missing > 0 || modified > 0 {
		healthStatus = yellowStyle.Render(fmt.Sprintf("⚠ %d issues", missing+modified))
	}
	printKV("  Status", healthStatus, labelStyle, valueStyle)
	printKV("  Health", formatHealthBar(healthy, total), labelStyle, valueStyle)
	printKV("  Total", fmt.Sprintf("%d managed files", total), labelStyle, valueStyle)

	if missing > 0 {
		printKV("  Missing", redStyle.Render(fmt.Sprintf("%d", missing)), labelStyle, valueStyle)
	} else {
		printKV("  Missing", dimStyle.Render("0"), labelStyle, valueStyle)
	}
	if modified > 0 {
		printKV("  Modified", yellowStyle.Render(fmt.Sprintf("%d", modified)), labelStyle, valueStyle)
	} else {
		printKV("  Modified", dimStyle.Render("0"), labelStyle, valueStyle)
	}
	fmt.Println()

	// Version info
	fmt.Println(headerStyle.Render("ℹ️  Info"))
	printKV("  CLI version", storeData.Meta.CLIVersion, labelStyle, valueStyle)
	installedAt := storeData.Meta.InstalledAt
	if t, err := time.Parse(time.RFC3339, installedAt); err == nil {
		installedAt = t.Format("2006-01-02")
	}
	printKV("  Last init", installedAt, labelStyle, valueStyle)
	fmt.Println()

	if missing > 0 || modified > 0 {
		cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00CFC5"))
		fmt.Printf("  Run %s for details\n", cyanStyle.Render("ai-setup doctor"))
	} else {
		fmt.Printf("  %s\n", greenStyle.Render("✓ Setup is healthy"))
	}

	return nil
}

// readStore is a shared helper that tries the SQLite DB first, then falls back to JSON manifest.
func readStore(dir string) (*types.StoreData, error) {
	// Try SQLite database first
	dbPath := db.DefaultDBPath(dir)
	if files.FileExists(dbPath) {
		database, err := db.Open(dbPath)
		if err == nil {
			defer database.Close()
			if err := db.RunMigrations(database.DB); err == nil {
				store := db.NewStore(database)
				data, err := store.ReadStoreData()
				if err == nil && data != nil {
					return data, nil
				}
			}
		}
	}

	// Fall back to JSON manifest
	if manifest.ManifestExists(dir) {
		data, err := manifest.ReadManifest(dir)
		if err != nil {
			return nil, err
		}
		return data, nil
	}

	return nil, aierror.ManifestNotFound(dir)
}

func printKV(label, value string, labelStyle, valueStyle lipgloss.Style) {
	fmt.Printf("  %s  %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
}

func toolList(tools []types.ToolId) string {
	names := make([]string, len(tools))
	for i, t := range tools {
		names[i] = string(t)
	}
	return strings.Join(names, ", ")
}

func formatFeatures(f *types.FeatureFlags) []string {
	if f == nil {
		return nil
	}
	var enabled []string
	if f.ContextEngineering {
		enabled = append(enabled, "contextEngineering")
	}
	if f.RPIWorkflow {
		enabled = append(enabled, "rpiWorkflow")
	}
	if f.ChainOfThought {
		enabled = append(enabled, "chainOfThought")
	}
	if f.TreeOfThoughts {
		enabled = append(enabled, "treeOfThoughts")
	}
	if f.ADREnforcement {
		enabled = append(enabled, "adrEnforcement")
	}
	if f.QualityGates {
		enabled = append(enabled, "qualityGates")
	}
	if f.AgentHarness {
		enabled = append(enabled, "agentHarness")
	}
	if f.BugResolution {
		enabled = append(enabled, "bugResolution")
	}
	if f.PivotHandling {
		enabled = append(enabled, "pivotHandling")
	}
	if len(enabled) == 0 {
		return []string{"(none)"}
	}
	return enabled
}

func formatHealthBar(healthy, total int) string {
	percent := 100
	if total > 0 {
		percent = (healthy * 100) / total
	}
	barWidth := 20
	filled := 0
	if total > 0 {
		filled = (healthy * barWidth) / total
	}
	empty := barWidth - filled

	barColor := lipgloss.Color("#04B575") // green
	if percent < 90 {
		barColor = lipgloss.Color("#FFA500") // yellow
	}
	if percent < 70 {
		barColor = lipgloss.Color("#FF4672") // red
	}

	barStyle := lipgloss.NewStyle().Foreground(barColor)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))

	filledStr := barStyle.Render(strings.Repeat("━", filled))
	emptyStr := dimStyle.Render(strings.Repeat("━", empty))
	return fmt.Sprintf("%s%s %d%%", filledStr, emptyStr, percent)
}

func boolStr(b bool) string {
	if b {
		return "Yes"
	}
	return "No"
}
