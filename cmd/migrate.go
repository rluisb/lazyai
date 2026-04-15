package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Migrate from a previous ai-setup version",
	Long:  "Migrate your existing ai-setup configuration from a previous version to the current format.",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println("TODO: implement migrate")
		return nil
	},
}

func init() {
	migrateCmd.Flags().String("from", "", "Source version to migrate from")
	migrateCmd.Flags().Bool("non-interactive", false, "Run without interactive prompts")
	rootCmd.AddCommand(migrateCmd)
}
