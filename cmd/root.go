package cmd

import (
	"context"

	"github.com/charmbracelet/fang"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags
var Version = "0.0.0-dev"

var rootCmd = &cobra.Command{
	Use:     "ai-setup",
	Short:   "AI development environment scaffold",
	Long:    "AI development environment scaffold — one command to set up your AI tools",
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose"); verbose {
			// Will set debug mode later
		}
	},
}

func init() {
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose debug output")
}

func Execute(ctx context.Context) error {
	return fang.Execute(ctx, rootCmd,
		fang.WithVersion(Version),
	)
}
