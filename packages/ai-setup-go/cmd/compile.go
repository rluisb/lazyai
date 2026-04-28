package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	"github.com/ricardoborges-teachable/ai-setup/internal/adapter"
	"github.com/ricardoborges-teachable/ai-setup/internal/db"
	"github.com/ricardoborges-teachable/ai-setup/internal/library"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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
	compileCmd.Flags().Bool("local-secrets", false, "Route Claude Code MCP writes to gitignored .claude/settings.local.json")
	// Spec 022 / E2.2: contract validation runs before MCP compile to catch
	// broken producer/consumer chains in skill frontmatter. Default on with
	// warn-only behavior; --strict-contracts upgrades warnings to failures.
	compileCmd.Flags().Bool("validate-contracts", true, "Validate skill output/produces_for/consumes chain before compile")
	compileCmd.Flags().Bool("strict-contracts", false, "Fail compile on contract warnings (default: warn-only)")
	rootCmd.AddCommand(compileCmd)
}

// fileExists returns true if the given path exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func runCompile(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	toolFilter, _ := cmd.Flags().GetString("tool")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	localSecrets, _ := cmd.Flags().GetBool("local-secrets")
	validateContracts, _ := cmd.Flags().GetBool("validate-contracts")
	strictContracts, _ := cmd.Flags().GetBool("strict-contracts")

	// Spec 022 / E2.2: validate skill chain before compile. Issues at error
	// severity always block; warnings block only when --strict-contracts is
	// passed.
	if validateContracts {
		libFS := library.GetLibraryFS()
		contracts, err := adapter.LoadSkillContracts(libFS)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  Warning: contract load failed: %v\n", err)
		} else {
			issues := adapter.ValidateContracts(contracts)
			if len(issues) > 0 {
				fmt.Fprintln(os.Stderr, adapter.FormatContractIssues(issues))
			}
			if adapter.HasContractErrors(issues) || (strictContracts && len(issues) > 0) {
				return fmt.Errorf("contract validation failed; pass --validate-contracts=false to override")
			}
		}
	}

	mcpConfigPath := filepath.Join(dir, ".ai", "mcp.json")
	if !fileExists(mcpConfigPath) {
		// Also try .ai/mcp.jsonc
		mcpConfigPath = filepath.Join(dir, ".ai", "mcp.jsonc")
		if !fileExists(mcpConfigPath) {
			return fmt.Errorf("no MCP config found at .ai/mcp.json. Run 'ai-setup init' first")
		}
	}

	// Open the store database (similar to helpers.go openStore)
	dbPath := filepath.Join(dir, ".ai-setup.db")
	database, err := db.Open(dbPath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	defer database.Close()

	// Run migrations.
	if err := db.RunMigrations(database); err != nil {
		database.Close()
		return fmt.Errorf("running migrations: %w", err)
	}

	// Auto-import from JSON if DB is new.
	imported, err := db.AutoImportJSON(dir, database)
	if err != nil {
		fmt.Fprintf(os.Stderr, "  Warning: JSON import failed: %v\n", err)
	}
	if imported {
		fmt.Println("  Imported existing .ai-setup.json → SQLite")
	}

	store := db.NewStore(database)
	storeData, err := store.ReadStoreData()
	if err != nil {
		// If the DB exists but has no initialized store rows yet, use defaults
		if errors.Is(err, sql.ErrNoRows) {
			defaults := types.DefaultStoreData()
			storeData = &defaults
		} else {
			return fmt.Errorf("reading store data: %w", err)
		}
	}

	// Determine which tools to compile for
	var tools []types.ToolId
	if toolFilter != "" {
		// Single tool requested via flag
		tools = []types.ToolId{types.ToolId(toolFilter)}
	} else {
		// Use tools from store configuration
		tools = storeData.Config.Tools
		// If store is empty, fall back to all known tools
		if len(tools) == 0 {
			// Get all registered tools from adapter registry
			reg := adapter.NewRegistry()
			for _, t := range reg.List() {
				tools = append(tools, t)
			}
		}
	}

	// Get adapter registry
	reg := adapter.NewRegistry()

	// Track new file records from compilation
	newFileRecords := []types.TrackedFile{}

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	cyanStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#00CFC5"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))

	fmt.Println()
	fmt.Println(headerStyle.Render("⚙️  Compile MCP Config"))
	fmt.Println()

	// Read MCP source once
	mcpData, err := os.ReadFile(mcpConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read MCP config: %w", err)
	}
	var dataMap map[string]any
	if err := json.Unmarshal(mcpData, &dataMap); err != nil {
		return fmt.Errorf("failed to parse MCP config: %w", err)
	}

	// Extract MCP servers from the config
	mcpServers, ok := dataMap["mcpServers"].(map[string]any)
	if !ok {
		// Try "servers" key as alternative
		if servers, ok := dataMap["servers"].(map[string]any); ok {
			mcpServers = servers
		} else {
			mcpServers = map[string]any{}
		}
	}

	printKV("  Source", mcpConfigPath, labelStyle, lipgloss.NewStyle())
	printKV("  Servers", fmt.Sprintf("%d configured", len(mcpServers)), labelStyle, lipgloss.NewStyle())
	fmt.Println()

	// Compile for each tool
	compiledCount := 0
	for _, toolId := range tools {
		// Get adapter for this tool
		adapt, err := reg.Get(toolId)
		if err != nil {
			fmt.Printf("    %s Skipping %s: %v\n", dimStyle.Render("○"), toolId, err)
			continue
		}

		// Get adapter name for display
		toolName := adapt.Name()

		// Compile MCP config for this tool
		var toolRecords []types.TrackedFile
		if dryRun {
			// For dry-run, we'd need to simulate compilation, but for now just show what we would do
			fmt.Printf("    %s Would compile MCP config for %s\n", cyanStyle.Render("[dry-run]"), toolName)
			compiledCount++
			continue
		}

		// Build CompileContext with scope info from the store. At workspace
		// scope, populate WorkspaceRoot+Repos so PropagateMcpToRepos has
		// what it needs after the root compile completes (Spec 022 / E2.3).
		homeDir, _ := os.UserHomeDir()
		compileCtx := adapter.CompileContext{
			TargetDir:    dir,
			HomeDir:      homeDir,
			SetupScope:   storeData.Config.SetupScope,
			FileRecords:  newFileRecords,
			LocalSecrets: localSecrets,
		}
		if storeData.Config.SetupScope == types.SetupScopeWorkspace {
			compileCtx.WorkspaceRoot = dir
			compileCtx.Repos = storeData.Config.Repos
		}

		// Actually compile
		toolRecords, err = adapt.CompileMCP(compileCtx)
		if err != nil {
			fmt.Printf("    %s Failed to compile %s: %v\n", dimStyle.Render("✗"), toolName, err)
			continue
		}

		// Check if any new files were generated
		if len(toolRecords) > 0 {
			// Add new records to our collection
			newFileRecords = append(newFileRecords, toolRecords...)

			// Get the primary config file path for this tool (first record)
			if len(toolRecords) > 0 {
				targetPath := toolRecords[0].Path
				fmt.Printf("    %s Compiled MCP config for %s -> %s\n", greenStyle.Render("✓"), toolName, targetPath)
			} else {
				fmt.Printf("    %s Compiled MCP config for %s (no files)\n", greenStyle.Render("✓"), toolName)
			}
			compiledCount++
		} else {
			fmt.Printf("    %s No MCP config generated for %s\n", dimStyle.Render("○"), toolName)
		}
	}
	fmt.Println()

	if dryRun {
		fmt.Printf("  %s Dry run complete. Would compile %d tool(s).\n", cyanStyle.Render("[dry-run]"), len(tools))
	} else {
		// Spec 022 / E2.3: at workspace scope, after the root compile,
		// propagate per-repo configs into each registered repo. This is
		// best-effort — a single repo failure logs but does not abort.
		if storeData.Config.SetupScope == types.SetupScopeWorkspace && len(storeData.Config.Repos) > 0 {
			homeDir, _ := os.UserHomeDir()
			propagatedCtx := adapter.CompileContext{
				HomeDir:       homeDir,
				SetupScope:    types.SetupScopeWorkspace,
				LocalSecrets:  localSecrets,
				WorkspaceRoot: dir,
				Repos:         storeData.Config.Repos,
			}
			propagated, err := adapter.PropagateMcpToRepos(reg, propagatedCtx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "  Warning: workspace propagation: %v\n", err)
			}
			if len(propagated) > 0 {
				newFileRecords = append(newFileRecords, propagated...)
				summary := adapter.SummarizeWorkspaceCompile(newFileRecords, propagated, storeData.Config.Repos)
				fmt.Printf("    %s Propagated MCP config to %d repo(s): %v\n",
					greenStyle.Render("✓"), len(summary.Repos), summary.Repos)
			}
		}

		// If we compiled any new records, update the store
		if len(newFileRecords) > 0 {
			// Merge new file records with existing ones
			allRecords := append(storeData.Files, newFileRecords...)
			storeData.Files = allRecords

			// Write back to store
			if err := store.WriteStoreData(storeData); err != nil {
				return fmt.Errorf("writing updated store: %w", err)
			}
		}
		fmt.Printf("  %s Compiled %d tool(s).\n", greenStyle.Render("✓"), compiledCount)
	}
	fmt.Println()

	return nil
}

// printKV is a helper for printing key-value pairs with styling
func printKV(label string, value string, labelStyle lipgloss.Style, valueStyle lipgloss.Style) {
	if value == "" {
		value = "-"
	}
	fmt.Printf("    %s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
}
