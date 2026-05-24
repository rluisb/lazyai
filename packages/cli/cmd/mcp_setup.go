package cmd

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/scaffold"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ---------------------------------------------------------------------------
// Command definition
// ---------------------------------------------------------------------------

var mcpSetupCmd = &cobra.Command{
	Use:   "mcp-setup",
	Short: "Configure AI agents to use the LazyAI orchestrator MCP server",
	Long: `Set up or update one or more AI agent configurations (OpenCode, Claude Code,
Copilot, Cursor) to connect to the LazyAI orchestrator daemon via MCP.

This command:
  1. Ensures .ai/mcp.json has the orchestrator server registered
  2. Compiles per-tool MCP configuration files for each requested tool
  3. Prints a summary of what was configured

Examples:
  lazyai-cli mcp-setup --tools opencode,claude-code
  lazyai-cli mcp-setup --all --dry-run
  lazyai-cli mcp-setup --tools cursor --project ./my-project`,
	RunE: runMcpSetup,
}

func init() {
	mcpSetupCmd.Flags().StringSliceVar(&mcpSetupToolsFlag, "tools", nil,
		"Tools to configure: opencode, claude-code, copilot, cursor (comma-separated)")
	mcpSetupCmd.Flags().BoolVar(&mcpSetupAllFlag, "all", false,
		"Configure all supported tools")
	mcpSetupCmd.Flags().BoolVar(&mcpSetupDryRunFlag, "dry-run", false,
		"Show what would be done without modifying files")
	mcpSetupCmd.Flags().StringVar(&mcpSetupScopeFlag, "scope", "project",
		"Scope: project, global, or workspace")
	mcpSetupCmd.Flags().StringVar(&mcpSetupProjectFlag, "project", "",
		"Target project directory (defaults to current working directory)")
	rootCmd.AddCommand(mcpSetupCmd)
	mcpSetupCmd.GroupID = "auth"
}

// ---------------------------------------------------------------------------
// Flag variables
// ---------------------------------------------------------------------------

var (
	mcpSetupToolsFlag   []string
	mcpSetupAllFlag     bool
	mcpSetupDryRunFlag  bool
	mcpSetupScopeFlag   string
	mcpSetupProjectFlag string
)

// ---------------------------------------------------------------------------
// Tool name -> ToolId mapping
// ---------------------------------------------------------------------------

var toolNameToID = map[string]types.ToolId{
	"opencode":    types.ToolIdOpenCode,
	"claude-code": types.ToolIdClaudeCode,
	"copilot":     types.ToolIdCopilot,
	"cursor":      types.ToolId("cursor"),
}

var supportedToolNames = []string{"opencode", "claude-code", "copilot", "cursor"}

// ---------------------------------------------------------------------------
// Run function
// ---------------------------------------------------------------------------

func runMcpSetup(cmd *cobra.Command, args []string) error {
	// Resolve target directory.
	targetDir := mcpSetupProjectFlag
	if targetDir == "" {
		wd, err := os.Getwd()
		if err != nil {
			return fmt.Errorf("determine working directory: %w", err)
		}
		targetDir = wd
	}

	// Parse scope.
	scope := types.SetupScope(mcpSetupScopeFlag)
	if !types.IsValidSetupScope(scope) {
		return fmt.Errorf("invalid scope %q (valid: project, global, workspace)", mcpSetupScopeFlag)
	}

	// Determine which tools to configure.
	toolIDs, err := resolveMcpSetupTools()
	if err != nil {
		return err
	}

	// Resolve home directory.
	homeDir, err := os.UserHomeDir()
	if err != nil {
		homeDir = ""
	}

	// Build compile options.
	compileOpts := adapter.CompileContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: scope,
	}

	// Styled output helpers.
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	greenStyle := lipgloss.NewStyle().Foreground(theme.Success)
	cyanStyle := lipgloss.NewStyle().Foreground(theme.Secondary)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dimmed)
	yellowStyle := lipgloss.NewStyle().Foreground(theme.Warning)

	fmt.Println()
	fmt.Println(headerStyle.Render("🔧 MCP Setup"))
	fmt.Println()

	// Step 1: Ensure .ai/mcp.json exists with orchestrator registered.
	fmt.Printf("  %s Ensuring .ai/mcp.json has orchestrator server...\n", dimStyle.Render("○"))

	aiDir := filepath.Join(targetDir, ".ai")
	mcpJSONPath := filepath.Join(aiDir, "mcp.json")

	if !files.FileExists(mcpJSONPath) {
		// Try .ai/mcp.jsonc as alternative.
		mcpJSONPath = filepath.Join(aiDir, "mcp.jsonc")
	}

	orchestratorEnabled := false
	if files.FileExists(mcpJSONPath) {
		// Check if orchestrator is already registered.
		catalog := adapter.ReadCanonicalMcp(targetDir)
		if catalog != nil {
			if server, ok := catalog.Servers["orchestrator"]; ok {
				orchestratorEnabled = server.Enabled == nil || *server.Enabled
			}
		}
	}

	if !orchestratorEnabled {
		// Scaffold orchestrator into .ai/mcp.json via library catalog.
		// libraryDir is the repo root so ScaffoldMcp can build the orchestrator
		// binary from packages/orchestrator if needed. If not available,
		// ScaffoldMcp will skip orchestrator binary preparation.
		libraryDir, _ := library.FindLibraryDir()
		libFS := library.GetLibraryFS()
		cliTools := []string{}
		enableServers := []string{"orchestrator"}
		var fileRecords []types.TrackedFile
		err := scaffold.ScaffoldMcp(targetDir, libraryDir, libFS, cliTools, enableServers, &fileRecords, types.ConflictStrategySkip, nil)
		if err != nil {
			cmdLog.Warn("scaffold MCP failed", "error", err)
		} else {
			fmt.Printf("    %s Scaffolded .ai/mcp.json with orchestrator\n",
				greenStyle.Render("✓"))
		}
	} else {
		fmt.Printf("    %s orchestrator already registered in .ai/mcp.json\n",
			greenStyle.Render("✓"))
	}

	fmt.Println()

	// Step 2: Compile MCP config for each requested tool.
	if len(toolIDs) == 0 {
		fmt.Printf("  %s No tools selected — nothing to compile.\n",
			dimStyle.Render("○"))
		fmt.Println()
		return nil
	}

	fmt.Printf("  %s Compiling per-tool MCP configs for %d tool(s)...\n",
		dimStyle.Render("○"), len(toolIDs))
	fmt.Println()

	registry := adapter.NewRegistry()
	compiled := 0
	skipped := 0

	// Sort tool IDs for deterministic output.
	sort.Slice(toolIDs, func(i, j int) bool {
		return string(toolIDs[i]) < string(toolIDs[j])
	})

	for _, toolID := range toolIDs {
		adapt, err := registry.Get(toolID)
		if err != nil {
			// Adapter not found — skip gracefully.
			fmt.Printf("    %s %s: no adapter registered, skipping\n",
				dimStyle.Render("○"), string(toolID))
			skipped++
			continue
		}

		toolName := adapt.Name()

		// Check scope support.
		if !adapter.IsScopeSupported(toolID, scope) {
			fmt.Printf("    %s %s: scope %q not supported for this tool, skipping\n",
				dimStyle.Render("○"), toolName, scope)
			skipped++
			continue
		}

		if mcpSetupDryRunFlag {
			fmt.Printf("    %s Would compile MCP config for %s\n",
				cyanStyle.Render("[dry-run]"), toolName)
			compiled++
			continue
		}

		// Compile the MCP config.
		_, err = adapter.CompileMCPForTool(toolID, compileOpts)
		if err != nil {
			fmt.Printf("    %s %s: failed to compile: %v\n",
				yellowStyle.Render("⚠"), toolName, err)
			continue
		}

		// Get the output path for display.
		outputPath := mcpConfigOutputPath(toolID, scope, targetDir)
		fmt.Printf("    %s Compiled MCP config for %s -> %s\n",
			greenStyle.Render("✓"), toolName, outputPath)
		compiled++
	}

	fmt.Println()

	// Summary.
	if mcpSetupDryRunFlag {
		fmt.Printf("  %s Dry run complete. Would configure %d tool(s), skipped %d.\n",
			cyanStyle.Render("[dry-run]"), compiled, skipped)
	} else {
		fmt.Printf("  %s Configured %d tool(s), skipped %d.\n",
			greenStyle.Render("✓"), compiled, skipped)
	}

	fmt.Println()

	// Next steps.
	printMcpSetupNextSteps(toolIDs)

	return nil
}

// ---------------------------------------------------------------------------
// Tool resolution
// ---------------------------------------------------------------------------

func resolveMcpSetupTools() ([]types.ToolId, error) {
	hasTools := len(mcpSetupToolsFlag) > 0
	hasAll := mcpSetupAllFlag

	if hasTools && hasAll {
		return nil, fmt.Errorf("--tools and --all cannot be used together")
	}

	if !hasTools && !hasAll {
		return nil, fmt.Errorf("specify --tools or --all")
	}

	if hasAll {
		return allMcpSetupTools(), nil
	}

	// Parse comma-separated tool names.
	var ids []types.ToolId
	seen := make(map[string]bool)
	for _, name := range mcpSetupToolsFlag {
		trimmed := strings.TrimSpace(name)
		if trimmed == "" {
			continue
		}
		if seen[trimmed] {
			continue
		}
		seen[trimmed] = true

		id, ok := toolNameToID[trimmed]
		if !ok {
			return nil, fmt.Errorf("unsupported tool %q (supported: %s)",
				trimmed, strings.Join(supportedToolNames, ", "))
		}
		ids = append(ids, id)
	}

	return ids, nil
}

// allMcpSetupTools returns all tool IDs that can be specified in --tools,
// regardless of whether an adapter is currently registered.
func allMcpSetupTools() []types.ToolId {
	ids := make([]types.ToolId, 0, len(toolNameToID))
	for _, id := range toolNameToID {
		ids = append(ids, id)
	}
	return ids
}

// ---------------------------------------------------------------------------
// Output path helpers
// ---------------------------------------------------------------------------

// mcpConfigOutputPath returns a human-readable path for the compiled MCP config.
func mcpConfigOutputPath(toolID types.ToolId, scope types.SetupScope, targetDir string) string {
	switch toolID {
	case types.ToolIdOpenCode:
		if scope == types.SetupScopeGlobal {
			return "~/.config/opencode/opencode.jsonc"
		}
		return filepath.Join(targetDir, ".opencode", "opencode.jsonc")
	case types.ToolIdClaudeCode:
		if scope == types.SetupScopeGlobal {
			return "~/.claude/settings.json"
		}
		return filepath.Join(targetDir, ".mcp.json")
	case types.ToolIdCopilot:
		if scope == types.SetupScopeGlobal {
			return "~/.copilot/mcp-config.json"
		}
		return filepath.Join(targetDir, ".vscode", "mcp.json")
	default:
		return fmt.Sprintf("<%s>", toolID)
	}
}

// printMcpSetupNextSteps prints context-specific next steps.
func printMcpSetupNextSteps(toolIDs []types.ToolId) {
	boldStyle := lipgloss.NewStyle().Bold(true)
	cyanStyle := lipgloss.NewStyle().Foreground(theme.Secondary)

	// Determine which tools were configured.
	hasOpenCode := false
	hasClaudeCode := false
	hasCopilot := false
	hasCursor := false

	for _, id := range toolIDs {
		switch id {
		case types.ToolIdOpenCode:
			hasOpenCode = true
		case types.ToolIdClaudeCode:
			hasClaudeCode = true
		case types.ToolIdCopilot:
			hasCopilot = true
		default:
			// cursor or unknown
			hasCursor = true
		}
	}

	fmt.Println(boldStyle.Render("  Next steps:"))
	fmt.Println()

	if hasOpenCode {
		fmt.Printf("    • %s Start OpenCode with MCP support:\n", cyanStyle.Render("OpenCode"))
		fmt.Printf("      %s\n", "opencode --mcp")
		fmt.Println()
	}

	if hasClaudeCode {
		fmt.Printf("    • %s Restart Claude Code to pick up MCP config:\n", cyanStyle.Render("Claude Code"))
		fmt.Printf("      %s\n", "claude")
		fmt.Println()
	}

	if hasCopilot {
		fmt.Printf("    • %s Restart VS Code to activate the MCP extension:\n", cyanStyle.Render("Copilot"))
		fmt.Printf("      %s\n", "code --reload")
		fmt.Println()
	}

	if hasCursor {
		fmt.Printf("    • %s Restart Cursor IDE to activate MCP connection\n", cyanStyle.Render("Cursor"))
		fmt.Println()
	}

	fmt.Printf("    • %s Verify setup: %s\n", cyanStyle.Render("Check"), "lazyai-cli server doctor")
	fmt.Println()
}
