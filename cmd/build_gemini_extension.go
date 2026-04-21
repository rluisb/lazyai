package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"github.com/ricardoborges-teachable/ai-setup/internal/geminiext"
	"github.com/ricardoborges-teachable/ai-setup/internal/library"
)

var buildGeminiExtensionCmd = &cobra.Command{
	Use:   "build-gemini-extension",
	Short: "Generate a Gemini CLI extension directory from the embedded library",
	Long: `Generate a Gemini CLI extension directory containing a GEMINI.md
template, custom slash commands, and (optionally) MCP servers from the
project's canonical .ai/mcp.json catalog. The output can be installed via
` + "`gemini extensions link <path>`" + ` for development or published as an
extension repository.

By default, writes to ./dist/gemini-extension. Use --out to override. The
output directory must be empty (or absent) unless --force is set.

Servers whose env or headers contain ${VAR} placeholders are skipped from
the bundled MCP list because Gemini extensions load without a prompt
layer for those values; users can hand-edit the extension's ` + "`settings`" + `
array after install to wire them up.`,
	RunE: runBuildGeminiExtension,
}

func init() {
	buildGeminiExtensionCmd.Flags().String("out", "./dist/gemini-extension", "Output directory for the generated extension")
	buildGeminiExtensionCmd.Flags().Bool("force", false, "Overwrite the output directory if it exists and is non-empty")
	rootCmd.AddCommand(buildGeminiExtensionCmd)
}

func runBuildGeminiExtension(cmd *cobra.Command, _ []string) error {
	outDir, _ := cmd.Flags().GetString("out")
	force, _ := cmd.Flags().GetBool("force")

	absOut, err := filepath.Abs(outDir)
	if err != nil {
		return fmt.Errorf("resolve --out: %w", err)
	}

	if err := preflightOutDir(absOut, force); err != nil {
		return err
	}

	// Read the canonical MCP catalog from the current project, if present.
	// Absence is not an error — the extension simply omits mcpServers.
	mcpCatalog := readGeminiExtensionMcpCatalog()

	libFS := library.GetLibraryFS()
	result, err := geminiext.Build(libFS, mcpCatalog, absOut, Version)
	if err != nil {
		return fmt.Errorf("build gemini extension: %w", err)
	}

	fmt.Printf("✓ Wrote %d files to %s\n", result.FileCount, result.OutDir)
	if len(result.SkippedMcpServers) > 0 {
		fmt.Printf("  Skipped %d MCP server(s) with placeholders: %v\n", len(result.SkippedMcpServers), result.SkippedMcpServers)
	}
	fmt.Printf("  Install for development: gemini extensions link %s\n", result.OutDir)
	return nil
}

// readGeminiExtensionMcpCatalog reads .ai/mcp.json from the current
// working directory and returns it as a map the geminiext package can
// consume. Returns nil on any read/parse failure (missing file, invalid
// JSON, empty catalog) — absence is intentional and means "no MCP
// servers shipped in this extension".
func readGeminiExtensionMcpCatalog() map[string]geminiext.McpServer {
	cwd, err := os.Getwd()
	if err != nil {
		return nil
	}
	path := filepath.Join(cwd, ".ai", "mcp.json")
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var parsed struct {
		Servers map[string]geminiext.McpServer `json:"servers"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil
	}
	if len(parsed.Servers) == 0 {
		return nil
	}
	return parsed.Servers
}
