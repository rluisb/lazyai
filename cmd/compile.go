package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/jsonc"
)

var compileCmd = &cobra.Command{
	Use:   "compile",
	Short: "Compile .ai/mcp.json to per-tool MCP configs",
	Long:  "Compile the unified MCP server configuration into per-tool configuration files.",
	RunE:  runCompile,
}

func init() {
	compileCmd.Flags().String("tool", "", "Compile only for a specific tool")
	compileCmd.Flags().Bool("dry-run", false, "Preview changes without writing files")
	rootCmd.AddCommand(compileCmd)
}

// Per-tool MCP config file mapping (from TypeScript PER_TOOL_MCP_CONFIG)
var perToolMCPConfig = map[string]string{
	"opencode":    "opencode.jsonc",
	"claude-code": ".mcp.json",
	"copilot":     ".vscode/mcp.json",
	"gemini":      ".gemini/settings.json",
	"codex":       "",
}

func runCompile(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	toolFilter, _ := cmd.Flags().GetString("tool")
	dryRun, _ := cmd.Flags().GetBool("dry-run")

	mcpConfigPath := filepath.Join(dir, ".ai", "mcp.json")
	if !files.FileExists(mcpConfigPath) {
		// Also try .ai/mcp.jsonc
		mcpConfigPath = filepath.Join(dir, ".ai", "mcp.jsonc")
		if !files.FileExists(mcpConfigPath) {
			return fmt.Errorf("no MCP config found at .ai/mcp.json. Run 'ai-setup init' first")
		}
	}

	// Read the MCP config (supports JSONC)
	mcpData, err := jsonc.ReadJSONCFile(mcpConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read MCP config: %w", err)
	}

	// Extract MCP servers from the config
	mcpServers, ok := mcpData["mcpServers"].(map[string]any)
	if !ok {
		// Try "servers" key as alternative
		if servers, ok := mcpData["servers"].(map[string]any); ok {
			mcpServers = servers
		} else {
			mcpServers = map[string]any{}
		}
	}

	// Determine which tools to compile for
	var tools []string
	if toolFilter != "" {
		tools = []string{toolFilter}
	} else {
		// Compile for all tools with known config paths
		for tool := range perToolMCPConfig {
			tools = append(tools, tool)
		}
	}

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00CFC5"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))

	fmt.Println()
	fmt.Println(headerStyle.Render("⚙️  Compile MCP Config"))
	fmt.Println()

	printKV("  Source", mcpConfigPath, labelStyle, lipgloss.NewStyle())
	printKV("  Servers", fmt.Sprintf("%d configured", len(mcpServers)), labelStyle, lipgloss.NewStyle())
	fmt.Println()

	// Compile for each tool
	compiledCount := 0
	for _, tool := range tools {
		configFile := perToolMCPConfig[tool]
		if configFile == "" {
			fmt.Printf("  %s %s — no config path defined (stub for Wave 3)\n", dimStyle.Render("○"), tool)
			continue
		}

		targetPath := filepath.Join(dir, configFile)
		printKV("  Tool", tool, labelStyle, lipgloss.NewStyle())
		printKV("  Target", targetPath, labelStyle, lipgloss.NewStyle())

		if dryRun {
			fmt.Printf("    %s Would compile %d servers -> %s\n", cyanStyle.Render("[dry-run]"), len(mcpServers), targetPath)
		} else {
			// Build per-tool MCP config
			toolConfig := map[string]any{
				"mcpServers": mcpServers,
			}

			configData, err := json.MarshalIndent(toolConfig, "", "  ")
			if err != nil {
				fmt.Printf("    %s Failed to serialize config: %v\n", dimStyle.Render("✗"), err)
				continue
			}

			if err := files.WriteFile(targetPath, configData, 0o644); err != nil {
				fmt.Printf("    %s Failed to write config: %v\n", dimStyle.Render("✗"), err)
				continue
			}

			fmt.Printf("    %s Compiled %d servers -> %s\n", greenStyle.Render("✓"), len(mcpServers), targetPath)
			compiledCount++
		}
		fmt.Println()
	}

	if dryRun {
		fmt.Printf("  %s Dry run complete. Would compile %d tool(s).\n", cyanStyle.Render("[dry-run]"), len(tools))
	} else {
		fmt.Printf("  %s Compiled %d tool(s).\n", greenStyle.Render("✓"), compiledCount)
	}
	fmt.Println()

	return nil
}
