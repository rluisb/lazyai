package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/eject"
	aierror "github.com/rluisb/lazyai/packages/cli/internal/error"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
)

var ejectCmd = &cobra.Command{
	Use:   "eject",
	Short: "Remove library management and keep files as-is",
	Long:  "Eject from LazyAI management, removing LazyAI metadata while leaving native host-tool files in place.",
	RunE:  runEject,
}

func init() {
	ejectCmd.Flags().Bool("no-interactive", false, "Skip confirmation prompt")
	rootCmd.AddCommand(ejectCmd)
	ejectCmd.GroupID = "lifecycle"
}

func runEject(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	nonInteractive, _ := cmd.Flags().GetBool("no-interactive")

	plan := eject.Inspect(dir)
	if len(plan.Existing) == 0 {
		return aierror.ManifestNotFound(dir)
	}

	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	warnStyle := lipgloss.NewStyle().Foreground(theme.Warning)
	greenStyle := lipgloss.NewStyle().Foreground(theme.Success)

	fmt.Println()
	fmt.Println(headerStyle.Render("🚀 Ejecting from LazyAI"))
	fmt.Println()
	fmt.Printf("  %s This will remove LazyAI management metadata but keep native files in place.\n", warnStyle.Render("⚠"))
	fmt.Printf("  %s Metadata files to remove: %d\n", warnStyle.Render("⚠"), len(plan.Existing))
	for _, path := range plan.Existing {
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			rel = path
		}
		fmt.Printf("    - %s\n", rel)
	}
	fmt.Println()

	if !nonInteractive {
		fmt.Print("  Are you sure you want to eject? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		response, err := reader.ReadString('\n')
		if err != nil {
			return aierror.Unknown("failed to read confirmation", err)
		}
		response = strings.TrimSpace(strings.ToLower(response))
		if response != "y" && response != "yes" {
			fmt.Println()
			fmt.Printf("  %s\n", greenStyle.Render("Eject cancelled"))
			return nil
		}
	}

	result, err := eject.Run(dir)
	if err != nil {
		return err
	}
	for _, path := range result.Removed {
		rel, relErr := filepath.Rel(dir, path)
		if relErr != nil {
			rel = path
		}
		fmt.Printf("  Removed %s\n", rel)
	}

	fmt.Println()
	fmt.Printf("  %s\n", greenStyle.Render("✓ Ejected successfully. Native files remain in place."))
	fmt.Println()

	return nil
}
