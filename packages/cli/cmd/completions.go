package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionsCmd is a deprecated alias for completion.
// It is hidden from help but still works for backward compatibility.
var completionsCmd = &cobra.Command{
	Use:        "completions",
	Short:      "Generate shell completion scripts",
	Long:       "Generate shell completion scripts for bash, zsh, fish, or powershell.",
	Args:       cobra.ExactArgs(1),
	ValidArgs:  []string{"bash", "zsh", "fish", "powershell"},
	Deprecated: "use 'completion' instead",
	Hidden:     true,
	RunE: func(cmd *cobra.Command, args []string) error {
		switch args[0] {
		case "bash":
			return rootCmd.GenBashCompletion(os.Stdout)
		case "zsh":
			return rootCmd.GenZshCompletion(os.Stdout)
		case "fish":
			return rootCmd.GenFishCompletion(os.Stdout, true)
		case "powershell":
			return rootCmd.GenPowerShellCompletion(os.Stdout)
		default:
			return fmt.Errorf("unsupported shell: %s (supported: bash, zsh, fish, powershell)", args[0])
		}
	},
}

func init() {
	rootCmd.AddCommand(completionsCmd)
}
