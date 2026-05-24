package cmd

import (
	"fmt"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/plugin"
)

var buildPluginCmd = &cobra.Command{
	Use:   "build-plugin",
	Short: "Generate a Claude Code plugin directory from the embedded library",
	Long: `Generate a Claude Code plugin directory containing LazyAI's agents,
skills, commands, and output styles. The output can be installed via
` + "`claude --plugin-dir <path>`" + ` or published to a marketplace.

By default, writes to ./dist/plugin. Use --out to override. The output
directory must be empty (or absent) unless --force is set.`,
	RunE: runBuildPlugin,
}

func init() {
	buildPluginCmd.Flags().String("out", "./dist/plugin", "Output directory for the generated plugin")
	buildPluginCmd.Flags().Bool("force", false, "Overwrite the output directory if it exists and is non-empty")
	rootCmd.AddCommand(buildPluginCmd)
	buildPluginCmd.GroupID = "auth"
}

func runBuildPlugin(cmd *cobra.Command, _ []string) error {
	outDir, _ := cmd.Flags().GetString("out")
	force, _ := cmd.Flags().GetBool("force")

	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("resolve --out: %w", err)
	}

	if err := preflightOutDir(absOut, force); err != nil {
		return err
	}

	libFS := library.GetLibraryFS()
	result, err := plugin.Build(libFS, absOut, Version)
	if err != nil {
		return fmt.Errorf("build plugin: %w", err)
	}

	fmt.Printf("✓ Wrote %d files to %s\n", result.FileCount, result.OutDir)
	fmt.Printf("  Install ephemerally: claude --plugin-dir %s\n", result.OutDir)
	return nil
}
