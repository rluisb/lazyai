package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/db"
	aierror "github.com/rluisb/lazyai/packages/cli/internal/error"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
)

var ejectCmd = &cobra.Command{
	Use:   "eject",
	Short: "Remove library management and keep files as-is",
	Long:  "Eject from LazyAI management, converting all managed files to standalone copies that are no longer tracked or updated.",
	RunE:  runEject,
}

func init() {
	ejectCmd.Flags().Bool("no-interactive", false, "Skip confirmation prompt")
	rootCmd.AddCommand(ejectCmd)
}

func runEject(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	nonInteractive, _ := cmd.Flags().GetBool("no-interactive")

	// Check that a store exists
	dbPath := db.DefaultDBPath(dir)
	manifestPath := filepath.Join(dir, ".ai-setup.json")

	dbExists := files.FileExists(dbPath)
	manifestExists := files.FileExists(manifestPath)

	if !dbExists && !manifestExists {
		return aierror.ManifestNotFound(dir)
	}

	// Read store to count files
	storeData, err := readStore(dir)
	if err != nil {
		return err
	}

	numFiles := len(storeData.Files)

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	warnStyle := lipgloss.NewStyle().Foreground(theme.Warning)
	greenStyle := lipgloss.NewStyle().Foreground(theme.Success)

	fmt.Println()
	fmt.Println(headerStyle.Render("🚀 Ejecting from LazyAI"))
	fmt.Println()
	fmt.Printf("  %s This will remove the .ai-setup.json manifest and .ai-setup.db database.\n", warnStyle.Render("⚠"))
	fmt.Printf("  %s Your %d managed files will be kept, but LazyAI will no longer update them.\n", warnStyle.Render("⚠"), numFiles)
	fmt.Println()

	// Confirmation
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

	// Delete .ai-setup.db
	if dbExists {
		if err := os.Remove(dbPath); err != nil {
			cmdLog.Warn("could not remove database", "path", dbPath, "error", err)
		} else {
			fmt.Printf("  Removed %s\n", dbPath)
		}
	}

	// Delete .ai-setup.json
	if manifestExists {
		if err := os.Remove(manifestPath); err != nil {
			cmdLog.Warn("could not remove manifest", "path", manifestPath, "error", err)
		} else {
			fmt.Printf("  Removed %s\n", manifestPath)
		}
	}

	fmt.Println()
	fmt.Printf("  %s\n", greenStyle.Render("✓ Ejected successfully. Your files remain in place."))
	fmt.Println()

	return nil
}
