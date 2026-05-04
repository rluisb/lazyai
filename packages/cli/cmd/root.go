package cmd

import (
	"context"

	"github.com/charmbracelet/fang"
	buildversion "github.com/rluisb/lazyai/packages/cli/internal/version"
	"github.com/spf13/cobra"
)

// Version is set at build time via ldflags.
// It is kept in cmd for existing release/build scripts and synced into
// internal/version so non-cmd packages can read the same value without cycles.
var Version = "0.0.0-dev"

var rootCmd = &cobra.Command{
	Use:     "lazyai-cli",
	Short:   "LazyAI development environment scaffold",
	Long:    "LazyAI development environment scaffold — one command to set up your AI tools",
	Version: Version,
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if verbose, _ := cmd.Root().PersistentFlags().GetBool("verbose"); verbose {
			// Will set debug mode later
		}
	},
}

func init() {
	syncBuildVersion()
	rootCmd.Version = Version
	rootCmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose debug output")
}

func syncBuildVersion() {
	if Version != buildversion.DevVersion {
		buildversion.Version = Version
		return
	}
	Version = buildversion.Version
}

func Execute(ctx context.Context) error {
	return fang.Execute(ctx, rootCmd,
		fang.WithVersion(Version),
	)
}
