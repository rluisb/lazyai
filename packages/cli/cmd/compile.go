package cmd

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"charm.land/lipgloss/v2"

	"github.com/spf13/cobra"

	"github.com/rluisb/lazyai/packages/cli/internal/adapter"
	"github.com/rluisb/lazyai/packages/cli/internal/aimanifest"
	"github.com/rluisb/lazyai/packages/cli/internal/compiler"
	"github.com/rluisb/lazyai/packages/cli/internal/db"
	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
	"github.com/rluisb/lazyai/packages/cli/internal/lockfile"
	"github.com/rluisb/lazyai/packages/cli/internal/theme"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
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
	compileCmd.GroupID = "lifecycle"
}

var getContractLibraryFS = library.GetLibraryFS

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
	if err := validateToolFlag(toolFilter); err != nil {
		return err
	}

	// Spec 022 / E2.2: validate skill chain before compile. Issues at error
	// severity always block; warnings block only when --strict-contracts is
	// passed.
	if validateContracts {
		libFS := getContractLibraryFS()
		contracts, err := compiler.LoadSkillContracts(libFS)
		if err != nil {
			cmdLog.Warn("contract load failed", "error", err)
		} else {
			issues := compiler.ValidateChain(contracts)
			if len(issues) > 0 {
				cmdLog.Error("contract validation failed", "issues", compiler.FormatContractIssues(issues))
			}
			if compiler.ContractValidationFails(issues, strictContracts) {
				return fmt.Errorf("contract validation failed; pass --validate-contracts=false to override")
			}
		}
	}

	// Open store database via shared logic.
	database, err := openStore(dir)
	if err != nil {
		return fmt.Errorf("opening store: %w", err)
	}
	defer database.Close()

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

	mcpRoot := dir
	if storeData.Config.SetupScope == types.SetupScopeWorkspace && storeData.Config.WorkspaceRoot != "" {
		mcpRoot = storeData.Config.WorkspaceRoot
	}

	// V2: the canonical manifest (.ai/lazyai.json), when present, is the source
	// of truth for which targets to compile. It is validated up front so an
	// invalid manifest fails fast. (Full manifest-required compilation and
	// asset routing through the plan/writer pipeline land with the adapter
	// capabilities model in Phase B.)
	aiDir := filepath.Join(mcpRoot, ".ai")
	var manifestTargets []types.ToolId
	if mf, mErr := aimanifest.Load(aiDir); mErr == nil {
		if err := mf.Validate(); err != nil {
			return fmt.Errorf("invalid .ai/lazyai.json: %w", err)
		}
		enabled, err := mf.EnabledTargets()
		if err != nil {
			return fmt.Errorf("invalid .ai/lazyai.json targets: %w", err)
		}
		manifestTargets = enabled
	} else if !errors.Is(mErr, aimanifest.ErrNotFound) {
		return fmt.Errorf("loading .ai/lazyai.json: %w", mErr)
	}
	mcpConfigPath := filepath.Join(mcpRoot, ".ai", "mcp.json")
	if !fileExists(mcpConfigPath) {
		// Also try .ai/mcp.jsonc
		mcpConfigPath = filepath.Join(mcpRoot, ".ai", "mcp.jsonc")
		if !fileExists(mcpConfigPath) {
			return fmt.Errorf("no MCP config found at .ai/mcp.json. Run 'lazyai-cli init' first")
		}
	}

	// Determine which tools to compile for
	var tools []types.ToolId
	if toolFilter != "" {
		// Single tool requested via flag
		tools = []types.ToolId{types.ToolId(toolFilter)}
	} else if len(manifestTargets) > 0 {
		// V2: manifest is authoritative for target selection when present.
		tools = manifestTargets
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

	// Validate that every library agent resolves to a model on every selected
	// tool, using the same models.Resolve adapters call at write time. Issues
	// warn by default and block only with --strict-contracts. Skipped when
	// --validate-contracts=false. Provider auth defaults to a live probe;
	// callers who pre-populate configured providers via the wizard can pass
	// the configured set in a future enhancement.
	if validateContracts && len(tools) > 0 {
		libFS := getContractLibraryFS()
		// Prefer the stored provider selection from the wizard. If empty
		// (legacy stores or first-run before wizard), pass nil so the
		// validator falls back to a live auth probe.
		var configuredProviders []string
		if storeData != nil && len(storeData.Selections.OpenCodeProviders) > 0 {
			configuredProviders = storeData.Selections.OpenCodeProviders
		}
		agentIssues, err := compiler.ValidateAgentResolutions(libFS, tools, configuredProviders)
		if err != nil {
			cmdLog.Warn("agent resolution check failed", "error", err)
		} else if len(agentIssues) > 0 {
			cmdLog.Warn("agent resolution warnings",
				"issues", "\n"+compiler.FormatAgentValidationIssues(agentIssues))
			if strictContracts {
				return fmt.Errorf("agent resolution failed; pass --validate-contracts=false to override")
			}
		}
	}

	// Track new file records from compilation
	newFileRecords := []types.TrackedFile{}

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(theme.Primary)
	labelStyle := lipgloss.NewStyle().Foreground(theme.Dimmed)
	greenStyle := lipgloss.NewStyle().Foreground(theme.Success)
	cyanStyle := lipgloss.NewStyle().Foreground(theme.Secondary)
	dimStyle := lipgloss.NewStyle().Foreground(theme.Dimmed)

	fmt.Println()
	fmt.Println(headerStyle.Render("⚙️  Compile MCP Config"))
	fmt.Println()
	// Read MCP source once
	mcpData, err := os.ReadFile(mcpConfigPath)
	if err != nil {
		return fmt.Errorf("failed to read MCP config: %w", err)
	}
	var dataMap map[string]any
	if strings.HasSuffix(mcpConfigPath, ".jsonc") {
		dataMap, err = jsonc.ParseJSONC(mcpData)
		if err != nil {
			return fmt.Errorf("failed to parse MCP config: %w", err)
		}
	} else {
		if err := json.Unmarshal(mcpData, &dataMap); err != nil {
			return fmt.Errorf("failed to parse MCP config: %w", err)
		}
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

		// Get adapter name for display, annotating beta/experimental adapters
		// so users know those targets are not yet fully docs-verified (EC-006).
		toolName := adapt.Name()
		if cap := adapt.Capabilities(); cap.IsBeta() {
			toolName = fmt.Sprintf("%s %s", toolName, dimStyle.Render(fmt.Sprintf("(%s)", cap.Support)))
		}

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
			LocalSecrets: localSecrets,
		}
		if storeData.Config.SetupScope == types.SetupScopeWorkspace {
			compileCtx.WorkspaceRoot = storeData.Config.WorkspaceRoot
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
				WorkspaceRoot: mcpRoot,
				Repos:         storeData.Config.Repos,
				Tools:         tools,
			}
			propagated, err := adapter.PropagateMcpToRepos(reg, propagatedCtx)
			if err != nil {
				cmdLog.Warn("workspace propagation failed", "error", err)
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

		// V2 (FR-003): record generated outputs in .ai/lock.json so future
		// compiles can detect drift and skip unchanged files. Best-effort: a
		// lockfile failure warns but does not fail the compile.
		if err := writeCompileLock(aiDir, mcpRoot, mcpData, newFileRecords, storeData.Config.Repos); err != nil {
			cmdLog.Warn("writing .ai/lock.json failed", "error", err)
		}
		fmt.Printf("  %s Compiled %d tool(s).\n", greenStyle.Render("✓"), compiledCount)
	}
	fmt.Println()

	if !dryRun {
		if catalog := adapter.ReadCanonicalMcp(mcpRoot); catalog != nil {
			PrintMcpNextSteps(adapter.GetEnabledServers(catalog))
		}
	}

	return nil
}

// writeCompileLock records compiled outputs in <aiDir>/lock.json. Output paths
// are resolved relative to mcpRoot when not absolute; unreadable outputs are
// skipped. The source hash for every MCP-derived output is the canonical
// .ai/mcp.json content hash.
//
// When repos is non-nil, records whose Source starts with "workspace:" are
// treated as per-repo propagated outputs: the path is resolved against
// mcpRoot/<repoPath> for reading, and stored in the lockfile as
// <repoPath>/<rec.Path> so entries are unique and resolvable.
func writeCompileLock(aiDir, mcpRoot string, mcpSource []byte, records []types.TrackedFile, repos []types.RepoInfo) error {
	lock, err := lockfile.Load(aiDir)
	if err != nil {
		return err
	}
	srcHash := lockfile.HashBytes(mcpSource)

	// Build a lookup from repo name to repo path for workspace propagation.
	repoPathByName := make(map[string]string, len(repos))
	for _, r := range repos {
		repoPathByName[r.Name] = r.Path
	}

	for _, rec := range records {
		p := rec.Path
		lockPath := p
		if !filepath.IsAbs(p) {
			// Check if this is a propagated record (workspace:repoName/... source).
			// Propagated records have repo-relative paths; resolve against the
			// repo directory under mcpRoot.
			if repoName, ok := workspaceRepoName(rec.Source); ok {
				if rp, found := repoPathByName[repoName]; found {
					p = filepath.Join(mcpRoot, rp, p)
					lockPath = filepath.Join(rp, rec.Path)
				} else {
					// Unknown repo — fall back to mcpRoot resolution.
					p = filepath.Join(mcpRoot, p)
				}
			} else {
				p = filepath.Join(mcpRoot, p)
			}
		}
		lockPath = filepath.ToSlash(lockPath)
		data, readErr := os.ReadFile(p)
		if readErr != nil {
			continue
		}
		lock.Upsert(lockfile.Generated{
			Path:       lockPath,
			Target:     "mcp",
			SourceHash: srcHash,
			OutputHash: lockfile.HashBytes(data),
			Managed:    false,
		})
	}
	lock.Version = lockfile.SchemaVersion
	lock.LazyaiVersion = Version
	lock.CompiledAt = time.Now().UTC().Format(time.RFC3339)
	return lock.Save(aiDir)
}

// workspaceRepoName extracts the repo name from a workspace-tagged source
// string of the form "workspace:repoName/...". Returns ("", false) when the
// source is not a workspace-tagged record.
func workspaceRepoName(source string) (string, bool) {
	const prefix = "workspace:"
	if !strings.HasPrefix(source, prefix) {
		return "", false
	}
	rest := source[len(prefix):]
	if idx := strings.IndexByte(rest, '/'); idx > 0 {
		return rest[:idx], true
	}
	return "", false
}

// printKV is a helper for printing key-value pairs with styling
func printKV(label string, value string, labelStyle lipgloss.Style, valueStyle lipgloss.Style) {
	if value == "" {
		value = "-"
	}
	fmt.Printf("    %s %s\n", labelStyle.Render(label+":"), valueStyle.Render(value))
}
