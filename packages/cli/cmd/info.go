package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	aierror "github.com/rluisb/lazyai/packages/cli/internal/error"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

var infoCmd = &cobra.Command{
	Use:   "info <name>",
	Short: "Show detailed artifact information",
	Long:  "Display detailed information about a specific artifact, including its configuration, dependencies, and source.",
	Args:  cobra.ExactArgs(1),
	RunE:  runInfo,
}

func init() {
	infoCmd.Flags().Bool("json", false, "Output as JSON")
	rootCmd.AddCommand(infoCmd)
	infoCmd.GroupID = "scaffold"
}

func runInfo(cmd *cobra.Command, args []string) error {
	name := args[0]
	outputJSON, _ := cmd.Flags().GetBool("json")

	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}

	// Read store data
	storeData, err := readStore(dir)
	if err != nil {
		return err
	}

	// Find the artifact by name
	var found *types.TrackedFile
	for i := range storeData.Files {
		f := &storeData.Files[i]
		// Match by basename (without extension) or by full path
		base := fileBasename(f.Path)
		if base == name || f.Path == name {
			found = f
			break
		}
	}

	if found == nil {
		return aierror.InvalidInput(
			fmt.Sprintf("artifact %q not found in store", name),
			map[string]any{"hint": "Use 'lazyai-cli list' to see all artifacts"},
		)
	}

	if outputJSON {
		output := map[string]any{
			"path":          found.Path,
			"hash":          found.Hash,
			"source":        found.Source,
			"owner":         string(found.Owner),
			"status":        string(found.Status),
			"installedAt":   found.InstalledAt,
			"lastCheckedAt": found.LastCheckedAt,
		}
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(output)
	}

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Dimmed)
	valueStyle := lipgloss.NewStyle()
	greenStyle := lipgloss.NewStyle().Foreground(theme.Success)
	yellowStyle := lipgloss.NewStyle().Foreground(theme.Warning)
	redStyle := lipgloss.NewStyle().Foreground(theme.Error)

	artifactType := classifyPath(found.Path)

	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf("🔍 %s: %s", capitalize(artifactType), fileBasename(found.Path))))
	fmt.Println()

	printKV("  Path", found.Path, labelStyle, valueStyle)
	printKV("  Hash", found.Hash, labelStyle, valueStyle)
	printKV("  Source", found.Source, labelStyle, valueStyle)
	printKV("  Owner", string(found.Owner), labelStyle, valueStyle)

	// Status with color
	switch found.Status {
	case types.FileStatusInstalled:
		printKV("  Status", greenStyle.Render("✓ installed"), labelStyle, valueStyle)
	case types.FileStatusModified:
		printKV("  Status", yellowStyle.Render("~ modified"), labelStyle, valueStyle)
	case types.FileStatusMissing:
		printKV("  Status", redStyle.Render("✗ missing"), labelStyle, valueStyle)
	default:
		printKV("  Status", string(found.Status), labelStyle, valueStyle)
	}

	// Format dates
	if found.InstalledAt != "" {
		printKV("  Installed", formatDate(found.InstalledAt), labelStyle, valueStyle)
	}
	if found.LastCheckedAt != "" {
		printKV("  Last checked", formatDate(found.LastCheckedAt), labelStyle, valueStyle)
	}

	fmt.Println()
	return nil
}

func fileBasename(path string) string {
	base := path
	// Get last segment
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			base = path[i+1:]
			break
		}
	}
	// Strip extension
	for i := 0; i < len(base); i++ {
		if base[i] == '.' {
			base = base[:i]
			break
		}
	}
	return base
}

func formatDate(isoDate string) string {
	t, err := time.Parse(time.RFC3339, isoDate)
	if err != nil {
		return isoDate
	}
	return t.Format("2006-01-02 15:04:05")
}
