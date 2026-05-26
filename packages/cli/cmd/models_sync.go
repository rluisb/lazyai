package cmd

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/models"
)

var modelsCmd = &cobra.Command{
	Use:   "models",
	Short: "Inspect and refresh the per-tool model catalog",
}

var modelsSyncCmd = &cobra.Command{
	Use:   "sync",
	Short: "Refresh internal/models/catalog_gen.go from models.dev/api.json",
	Long: `Fetch the live model catalog from models.dev, filter to the providers
each target CLI cares about (anthropic, openai, github-copilot, google,
ollama-cloud, opencode, opencode-go), diff against the vendored snapshot,
and (with confirmation) regenerate internal/models/catalog_gen.go.

Aborts if any curated tier reference in catalog.go points at an upstream-
removed model — the curation must be repaired before the catalog is
regenerated.`,
	RunE: runModelsSync,
}

func init() {
	modelsSyncCmd.Flags().Bool("yes", false, "Skip the confirmation prompt and write the new catalog if there are no missing references")
	modelsSyncCmd.Flags().String("output", "", "Path to write the regenerated catalog (default: packages/cli/internal/models/catalog_gen.go relative to cwd)")
	modelsSyncCmd.Flags().Bool("dry-run", false, "Print the diff and exit without writing")
	modelsCmd.AddCommand(modelsSyncCmd)
	rootCmd.AddCommand(modelsCmd)
	modelsCmd.GroupID = "scaffold"
}

func runModelsSync(cmd *cobra.Command, _ []string) error {
	yes, _ := cmd.Flags().GetBool("yes")
	output, _ := cmd.Flags().GetString("output")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	if output == "" {
		output = filepath.Join("packages", "cli", "internal", "models", "catalog_gen.go")
	}

	report, err := models.Sync(context.Background(), nil)
	if err != nil {
		return err
	}
	report.Render(os.Stdout)

	if len(report.MissingCurated) > 0 {
		return fmt.Errorf("\ncatalog.go references %d upstream-removed model(s); repair curation before re-running sync",
			len(report.MissingCurated))
	}
	if len(report.Diffs) == 0 && !dryRun {
		fmt.Println("\ncatalog already up to date — no write")
		return nil
	}
	if dryRun {
		fmt.Println("\ndry-run: not writing")
		return nil
	}
	if !yes && !confirmWrite(output) {
		fmt.Println("aborted")
		return nil
	}

	if err := os.WriteFile(output, report.GeneratedSource, 0o644); err != nil {
		return fmt.Errorf("write %s: %w", output, err)
	}
	fmt.Printf("wrote %s (%d bytes)\n", output, len(report.GeneratedSource))
	fmt.Println("run `go fmt ./internal/models/...` and `go test ./internal/models/...` to verify")
	return nil
}

func confirmWrite(path string) bool {
	fmt.Printf("\nwrite %s? [y/N] ", path)
	scanner := bufio.NewScanner(os.Stdin)
	if !scanner.Scan() {
		return false
	}
	answer := strings.ToLower(strings.TrimSpace(scanner.Text()))
	return answer == "y" || answer == "yes"
}
