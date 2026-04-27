package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var completionsCmd = &cobra.Command{
	Use:       "completions [shell]",
	Short:     "Generate shell completion scripts",
	Long:      "Generate shell completion scripts for bash, zsh, fish, or powershell.",
	Args:      cobra.ExactArgs(1),
	ValidArgs: []string{"bash", "zsh", "fish", "powershell"},
	RunE: func(cmd *cobra.Command, args []string) error {
		shell := args[0]
		switch shell {
		case "bash":
			return rootCmd.GenBashCompletion(cmd.OutOrStdout())
		case "zsh":
			return rootCmd.GenZshCompletion(cmd.OutOrStdout())
		case "fish":
			return rootCmd.GenFishCompletion(cmd.OutOrStdout(), true)
		case "powershell":
			return rootCmd.GenPowerShellCompletion(cmd.OutOrStdout())
		default:
			return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", shell)
		}
	},
}

func init() {
	rootCmd.AddCommand(completionsCmd)
}
