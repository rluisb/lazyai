package cmd

import (
	"fmt"

	"charm.land/huh/v2"

	"github.com/spf13/cobra"
)

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update files to match current library versions",
	Long:  "Update managed files to match the current library versions, resolving conflicts as needed.",
	RunE:  runUpdate,
}

func init() {
	updateCmd.Flags().Bool("force", false, "Overwrite local changes without prompting")
	updateCmd.Flags().Bool("non-interactive", false, "Run without interactive prompts")
	updateCmd.Flags().Bool("dry-run", false, "Show what would be changed without making changes")
	rootCmd.AddCommand(updateCmd)
}

func runUpdate(cmd *cobra.Command, args []string) error {
	force, _ := cmd.Flags().GetBool("force")
	nonInteractive, _ := cmd.Flags().GetBool("non-interactive")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if nonInteractive {
		return runUpdateNonInteractive(force, dryRun)
	}
	return runUpdateInteractive(force, dryRun)
}

func runUpdateInteractive(force, dryRun bool) error {
	// Ask about force behavior if not already set.
	if !force {
		var forceConfirm bool
		forcePrompt := huh.NewConfirm().
			Title("Overwrite local changes without prompting?").
			Value(&forceConfirm)

		if err := huh.NewForm(huh.NewGroup(forcePrompt)).Run(); err != nil {
			return fmt.Errorf("cancelled: %w", err)
		}
		force = forceConfirm
	}

	// Confirm the update.
	var proceed bool
	confirmPrompt := huh.NewConfirm().
		Title("Proceed with update?").
		Value(&proceed)

	if err := huh.NewForm(huh.NewGroup(confirmPrompt)).Run(); err != nil {
		return fmt.Errorf("cancelled: %w", err)
	}

	if !proceed {
		fmt.Println("Update cancelled.")
		return nil
	}

	// TODO: Implement actual update logic once scaffold packages are ported.
	fmt.Printf("Would update (force=%v, dryRun=%v)\n", force, dryRun)
	return nil
}

func runUpdateNonInteractive(force, dryRun bool) error {
	// TODO: Implement actual update logic once scaffold packages are ported.
	fmt.Printf("Would update (force=%v, dryRun=%v)\n", force, dryRun)
	return nil
}
