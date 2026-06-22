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
	Short: "Generate a target bundle from the embedded library",
	Long: `Generate a distributable bundle from the embedded library.

Supported targets:
  - claude: Claude Code plugin directory
  - copilot-cli: GitHub Copilot CLI plugin bundle
  - omp: OMP plugin bundle
  - pi: Pi package bundle

By default, writes to ./dist/plugin. Use --out to override. The output
directory must be empty (or absent) unless --force is set.`,
	RunE: runBuildPlugin,
}

func init() {
	buildPluginCmd.Flags().String("out", "./dist/plugin", "Output directory for the generated bundle")
	buildPluginCmd.Flags().String("target", "claude", "Bundle target: claude, copilot-cli, omp, pi")
	buildPluginCmd.Flags().Bool("force", false, "Overwrite the output directory if it exists and is non-empty")
	rootCmd.AddCommand(buildPluginCmd)
	buildPluginCmd.GroupID = "auth"
}

func runBuildPlugin(cmd *cobra.Command, _ []string) error {
	outDir, _ := cmd.Flags().GetString("out")
	targetRaw, _ := cmd.Flags().GetString("target")
	force, _ := cmd.Flags().GetBool("force")

	target, err := plugin.NormalizeTarget(targetRaw)
	if err != nil {
		return err
	}

	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("resolve --out: %w", err)
	}

	if err := preflightOutDir(absOut, force); err != nil {
		return err
	}

	libFS := library.GetLibraryFS()
	result, err := plugin.BuildTarget(libFS, absOut, Version, target)
	if err != nil {
		return fmt.Errorf("build plugin: %w", err)
	}

	fmt.Printf("✓ Wrote %d files to %s\n", result.FileCount, result.OutDir)
	switch target {
	case plugin.BundleTargetClaude:
		fmt.Printf("  Install ephemerally: claude --plugin-dir %s\n", result.OutDir)
	case plugin.BundleTargetCopilotCLI:
		fmt.Printf("  Bundle target: GitHub Copilot CLI plugin\n")
	case plugin.BundleTargetOmp:
		fmt.Printf("  Bundle target: OMP plugin\n")
	case plugin.BundleTargetPi:
		fmt.Printf("  Bundle target: Pi package\n")
	}
	return nil
}
