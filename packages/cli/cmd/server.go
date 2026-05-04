package cmd

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/spf13/cobra"

	aierror "github.com/rluisb/lazyai/packages/cli/internal/error"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
	"github.com/rluisb/lazyai/packages/cli/internal/library"
)

// ---------------------------------------------------------------------------
// Catalog types — mirrors library/mcp/catalog.json structure
// ---------------------------------------------------------------------------

// CatalogServer describes a single MCP server entry in the catalog.
type CatalogServer struct {
	Description     string            `json:"description"`
	Command         string            `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	URL             string            `json:"url,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Tools           []string          `json:"tools,omitempty"`
	Enabled         bool              `json:"enabled"`
	RequiresInstall bool              `json:"requiresInstall,omitempty"`
	InstallHint     string            `json:"installHint,omitempty"`
}

// Catalog is the top-level structure of library/mcp/catalog.json.
type Catalog struct {
	Servers map[string]CatalogServer `json:"servers"`
}

// ---------------------------------------------------------------------------
// Doctor check types
// ---------------------------------------------------------------------------

// CheckStatus is pass, fail, or skip.
type CheckStatus string

const (
	CheckPass CheckStatus = "pass"
	CheckFail CheckStatus = "fail"
	CheckSkip CheckStatus = "skip"
)

// CheckResult is a single diagnostic check outcome.
type CheckResult struct {
	Name        string      `json:"name"`
	Status      CheckStatus `json:"status"`
	Message     string      `json:"message"`
	Remediation string      `json:"remediation,omitempty"`
}

// ServerHealthReport is the aggregated health for one server.
type ServerHealthReport struct {
	Server  string        `json:"server"`
	Overall string        `json:"overall"` // healthy, unhealthy, partial
	Checks  []CheckResult `json:"checks"`
}

// ---------------------------------------------------------------------------
// Command wiring
// ---------------------------------------------------------------------------

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Manage MCP server configurations",
	Long:  "Add, remove, list, and validate MCP server configurations.",
}

var serverListCmd = &cobra.Command{
	Use:   "list",
	Short: "List MCP server configurations",
	Long:  "List all MCP servers from the catalog and show which are enabled in the current project.",
	RunE:  runServerList,
}

var serverAddCmd = &cobra.Command{
	Use:   "add <name>",
	Short: "Add an MCP server configuration",
	Long:  "Enable an MCP server from the catalog in the current project.",
	Args:  cobra.ExactArgs(1),
	RunE:  runServerAdd,
}

var serverRemoveCmd = &cobra.Command{
	Use:   "remove <name>",
	Short: "Remove an MCP server configuration",
	Long:  "Disable an MCP server in the current project.",
	Args:  cobra.ExactArgs(1),
	RunE:  runServerRemove,
}

var serverDoctorCmd = &cobra.Command{
	Use:   "doctor [name]",
	Short: "Validate MCP server configurations",
	Long:  "Validate that enabled MCP servers have correct configuration in .ai/mcp.json and per-tool config files.",
	Args:  cobra.MaximumNArgs(1),
	RunE:  runServerDoctor,
}

func init() {
	serverListCmd.Flags().Bool("json", false, "Output as JSON")
	serverAddCmd.Flags().Bool("no-interactive", false, "Skip confirmation prompt")
	serverRemoveCmd.Flags().Bool("no-interactive", false, "Skip confirmation prompt")
	serverDoctorCmd.Flags().Bool("json", false, "Output as JSON")

	serverCmd.AddCommand(serverListCmd)
	serverCmd.AddCommand(serverAddCmd)
	serverCmd.AddCommand(serverRemoveCmd)
	serverCmd.AddCommand(serverDoctorCmd)
	rootCmd.AddCommand(serverCmd)
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// readCatalog reads and parses library/mcp/catalog.json.
func readCatalog() (*Catalog, error) {
	libFS := library.GetLibraryFS()

	data, err := files.ReadFS(libFS, "mcp/catalog.json")
	if err != nil {
		return nil, fmt.Errorf("failed to read catalog: %w", err)
	}

	var catalog Catalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil, fmt.Errorf("failed to parse catalog: %w", err)
	}
	return &catalog, nil
}

// enabledServersFromStore reads the store in dir and returns the enableServers list.
// Returns an empty slice (not an error) if the store doesn't exist yet.
func enabledServersFromStore(dir string) []string {
	storeData, err := readStore(dir)
	if err != nil {
		return nil
	}
	if storeData.Config.EnableServers == nil {
		return []string{}
	}
	return storeData.Config.EnableServers
}

// commandDisplay returns a human-readable string for a catalog server's command.
func commandDisplay(s CatalogServer) string {
	if s.Command != "" {
		parts := []string{s.Command}
		parts = append(parts, s.Args...)
		return strings.Join(parts, " ")
	}
	if s.URL != "" {
		return "remote: " + s.URL
	}
	return ""
}

// ---------------------------------------------------------------------------
// server list
// ---------------------------------------------------------------------------

func runServerList(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	outputJSON, _ := cmd.Flags().GetBool("json")

	catalog, err := readCatalog()
	if err != nil {
		return err
	}

	enabledList := enabledServersFromStore(dir)
	enabledSet := make(map[string]bool, len(enabledList))
	for _, s := range enabledList {
		enabledSet[s] = true
	}

	// Build sorted list of server names.
	names := make([]string, 0, len(catalog.Servers))
	for name := range catalog.Servers {
		names = append(names, name)
	}
	sort.Strings(names)

	type serverRow struct {
		Name            string `json:"name"`
		Enabled         bool   `json:"enabled"`
		Description     string `json:"description"`
		Command         string `json:"command,omitempty"`
		RequiresInstall bool   `json:"requiresInstall,omitempty"`
		InstallHint     string `json:"installHint,omitempty"`
	}

	rows := make([]serverRow, 0, len(names))
	for _, name := range names {
		entry := catalog.Servers[name]
		isEnabled := enabledSet[name] || entry.Enabled
		rows = append(rows, serverRow{
			Name:            name,
			Enabled:         isEnabled,
			Description:     entry.Description,
			Command:         commandDisplay(entry),
			RequiresInstall: entry.RequiresInstall,
			InstallHint:     entry.InstallHint,
		})
	}

	// JSON output
	if outputJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		return enc.Encode(rows)
	}

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	boldStyle := lipgloss.NewStyle().Bold(true)
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))

	fmt.Println()
	fmt.Println(headerStyle.Render("🔌 MCP Servers"))
	fmt.Println()

	enabledCount := 0
	for _, row := range rows {
		if row.Enabled {
			enabledCount++
		}

		var marker string
		var nameCol string
		if row.Enabled {
			marker = greenStyle.Render("●")
			nameCol = boldStyle.Render(row.Name)
		} else {
			marker = dimStyle.Render("○")
			nameCol = dimStyle.Render(row.Name)
		}

		extras := []string{}
		if row.RequiresInstall {
			extras = append(extras, yellowStyle.Render("requires install"))
		}
		extraStr := ""
		if len(extras) > 0 {
			extraStr = " [" + strings.Join(extras, ", ") + "]"
		}

		desc := dimStyle.Render(row.Description)
		fmt.Printf("  %s %s — %s%s\n", marker, nameCol, desc, extraStr)
	}

	fmt.Println()
	fmt.Printf("  %s of %s enabled\n",
		greenStyle.Render(fmt.Sprintf("%d", enabledCount)),
		dimStyle.Render(fmt.Sprintf("%d", len(rows))),
	)
	fmt.Println()

	return nil
}

// ---------------------------------------------------------------------------
// server add
// ---------------------------------------------------------------------------

func runServerAdd(cmd *cobra.Command, args []string) error {
	name := args[0]
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}

	catalog, err := readCatalog()
	if err != nil {
		return err
	}

	entry, exists := catalog.Servers[name]
	if !exists {
		available := make([]string, 0, len(catalog.Servers))
		for k := range catalog.Servers {
			available = append(available, k)
		}
		sort.Strings(available)
		return aierror.InvalidInput(
			fmt.Sprintf("unknown MCP server: %s", name),
			map[string]any{"available": available},
		)
	}

	// Read store data
	storeData, err := readStore(dir)
	if err != nil {
		return err
	}

	// Check if already enabled
	currentSet := make(map[string]bool)
	for _, s := range storeData.Config.EnableServers {
		currentSet[s] = true
	}

	if currentSet[name] {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		fmt.Println()
		fmt.Printf("  %s %s is already enabled.\n", headerStyle.Render("ℹ"), name)
		fmt.Println()
		return nil
	}

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	warnStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))

	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf("🔌 Enabling MCP server: %s", name)))
	fmt.Println()

	if entry.RequiresInstall && entry.InstallHint != "" {
		fmt.Printf("  %s %s requires installation: %s\n", warnStyle.Render("⚠"), name, entry.InstallHint)
		fmt.Println()
	}

	// Add to enableServers
	currentSet[name] = true
	newServers := make([]string, 0, len(currentSet))
	for s := range currentSet {
		newServers = append(newServers, s)
	}
	sort.Strings(newServers)
	storeData.Config.EnableServers = newServers

	// Write store data back
	if err := writeStoreData(dir, storeData); err != nil {
		return fmt.Errorf("failed to update store: %w", err)
	}

	fmt.Printf("  %s Enabled %s (%d servers now active)\n",
		greenStyle.Render("✓"), name, len(newServers))
	fmt.Println()
	fmt.Printf("  Run %s to regenerate per-tool configs.\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00CFC5")).Render("lazyai-cli compile"))
	fmt.Println()

	return nil
}

// ---------------------------------------------------------------------------
// server remove
// ---------------------------------------------------------------------------

func runServerRemove(cmd *cobra.Command, args []string) error {
	name := args[0]
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}

	catalog, err := readCatalog()
	if err != nil {
		return err
	}

	if _, exists := catalog.Servers[name]; !exists {
		available := make([]string, 0, len(catalog.Servers))
		for k := range catalog.Servers {
			available = append(available, k)
		}
		sort.Strings(available)
		return aierror.InvalidInput(
			fmt.Sprintf("unknown MCP server: %s", name),
			map[string]any{"available": available},
		)
	}

	// Read store data
	storeData, err := readStore(dir)
	if err != nil {
		return err
	}

	// Check if currently enabled
	currentSet := make(map[string]bool)
	for _, s := range storeData.Config.EnableServers {
		currentSet[s] = true
	}

	if !currentSet[name] {
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		fmt.Println()
		fmt.Printf("  %s %s is not currently enabled.\n", headerStyle.Render("ℹ"), name)
		fmt.Println()
		return nil
	}

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))

	fmt.Println()
	fmt.Println(headerStyle.Render(fmt.Sprintf("🔌 Disabling MCP server: %s", name)))
	fmt.Println()

	// Remove from enableServers
	delete(currentSet, name)
	newServers := make([]string, 0, len(currentSet))
	for s := range currentSet {
		newServers = append(newServers, s)
	}
	sort.Strings(newServers)
	storeData.Config.EnableServers = newServers

	// Write store data back
	if err := writeStoreData(dir, storeData); err != nil {
		return fmt.Errorf("failed to update store: %w", err)
	}

	fmt.Printf("  %s Disabled %s (%d servers remaining)\n",
		greenStyle.Render("✓"), name, len(newServers))
	fmt.Println()
	fmt.Printf("  Run %s to regenerate per-tool configs.\n",
		lipgloss.NewStyle().Foreground(lipgloss.Color("#00CFC5")).Render("lazyai-cli compile"))
	fmt.Println()

	return nil
}

// ---------------------------------------------------------------------------
// server doctor
// ---------------------------------------------------------------------------

func runServerDoctor(cmd *cobra.Command, args []string) error {
	dir, _ := cmd.Flags().GetString("dir")
	if dir == "" {
		dir, _ = os.Getwd()
	}
	outputJSON, _ := cmd.Flags().GetBool("json")

	// Read store data
	storeData, err := readStore(dir)
	if err != nil {
		return err
	}

	catalog, err := readCatalog()
	if err != nil {
		return err
	}

	enabled := storeData.Config.EnableServers
	tools := storeData.Config.Tools

	// Determine which servers to check
	var targets []string
	if len(args) > 0 {
		targets = []string{args[0]}
	} else {
		targets = enabled
	}

	if len(targets) == 0 {
		if outputJSON {
			enc := json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
			return enc.Encode(map[string]any{"reports": []ServerHealthReport{}})
		}
		headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
		fmt.Println()
		fmt.Printf("  %s No enabled MCP servers to check.\n", headerStyle.Render("ℹ"))
		fmt.Println()
		return nil
	}

	// Convert tools to string slice for health checks
	toolStrs := make([]string, len(tools))
	for i, t := range tools {
		toolStrs[i] = string(t)
	}

	// Run health checks for each target
	reports := make([]ServerHealthReport, 0, len(targets))
	for _, target := range targets {
		report := runServerHealthChecks(dir, target, catalog, toolStrs)
		reports = append(reports, report)
	}

	// JSON output
	if outputJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		_ = enc.Encode(map[string]any{"reports": reports})
		hasUnhealthy := false
		for _, r := range reports {
			if r.Overall == "unhealthy" {
				hasUnhealthy = true
			}
		}
		if hasUnhealthy {
			unhealthyCount := 0
			for _, r := range reports {
				if r.Overall == "unhealthy" {
					unhealthyCount++
				}
			}
			return fmt.Errorf("doctor found issues in %d server(s)", unhealthyCount)
		}
		return nil
	}

	// Styled output
	headerStyle := lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("#7D56F4"))
	labelStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))
	greenStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#04B575"))
	redStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FF4672"))
	yellowStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#FFA500"))
	dimStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("#6C6C6C"))

	fmt.Println()
	fmt.Println(headerStyle.Render("🩺 Server Doctor"))
	fmt.Println()

	for _, report := range reports {
		emoji := "✅"
		if report.Overall == "unhealthy" {
			emoji = "❌"
		} else if report.Overall == "partial" {
			emoji = "⚠️"
		}

		passCount := 0
		failCount := 0
		skipCount := 0
		for _, c := range report.Checks {
			switch c.Status {
			case CheckPass:
				passCount++
			case CheckFail:
				failCount++
			case CheckSkip:
				skipCount++
			}
		}

		fmt.Printf("  %s %s\n", emoji, boldStyle().Render(report.Server))
		printKV("    Overall", report.Overall, labelStyle, lipgloss.NewStyle())
		printKV("    Pass", greenStyle.Render(fmt.Sprintf("%d", passCount)), labelStyle, lipgloss.NewStyle())
		if failCount > 0 {
			printKV("    Fail", redStyle.Render(fmt.Sprintf("%d", failCount)), labelStyle, lipgloss.NewStyle())
		} else {
			printKV("    Fail", dimStyle.Render("0"), labelStyle, lipgloss.NewStyle())
		}
		printKV("    Skip", dimStyle.Render(fmt.Sprintf("%d", skipCount)), labelStyle, lipgloss.NewStyle())
		fmt.Println()

		// Show individual checks
		for _, check := range report.Checks {
			var glyph string
			switch check.Status {
			case CheckPass:
				glyph = greenStyle.Render("✓")
			case CheckFail:
				glyph = redStyle.Render("✗")
			case CheckSkip:
				glyph = dimStyle.Render("—")
			}

			fmt.Printf("    %s %s — %s\n", glyph, check.Name, check.Message)
			if check.Status == CheckFail && check.Remediation != "" {
				fmt.Printf("      %s %s\n", dimStyle.Render("→"), dimStyle.Render(check.Remediation))
			}
		}
		fmt.Println()
	}

	// Summary
	hasUnhealthy := false
	for _, r := range reports {
		if r.Overall == "unhealthy" {
			hasUnhealthy = true
		}
	}

	if hasUnhealthy {
		unhealthyCount := 0
		for _, r := range reports {
			if r.Overall == "unhealthy" {
				unhealthyCount++
			}
		}
		fmt.Printf("  %s\n", yellowStyle.Render(fmt.Sprintf("⚠ %d server(s) have issues", unhealthyCount)))
		fmt.Println()
		return fmt.Errorf("doctor found issues in %d server(s)", unhealthyCount)
	}

	fmt.Printf("  %s\n", greenStyle.Render("✓ All enabled servers healthy"))
	fmt.Println()
	return nil
}

// runServerHealthChecks performs L1 config checks for a single server.
// L3 stdio handshake is not implemented in Go (requires Node.js MCP SDK).
func runServerHealthChecks(dir string, name string, catalog *Catalog, tools []string) ServerHealthReport {
	var checks []CheckResult

	// Check if server exists in catalog
	entry, exists := catalog.Servers[name]
	if !exists {
		checks = append(checks, CheckResult{
			Name:        "catalog",
			Status:      CheckFail,
			Message:     fmt.Sprintf("server '%s' not found in library/mcp/catalog.json", name),
			Remediation: "Run 'lazyai-cli server list' to see available servers",
		})
		return finalizeServerReport(name, checks)
	}

	// L1.1: Check .ai/mcp.json exists
	canonicalPath := filepath.Join(dir, ".ai", "mcp.json")
	if !files.FileExists(canonicalPath) {
		checks = append(checks, CheckResult{
			Name:        "canonical mcp.json",
			Status:      CheckFail,
			Message:     ".ai/mcp.json is missing",
			Remediation: "Run 'lazyai-cli compile' to regenerate it",
		})
		return finalizeServerReport(name, checks)
	}

	// Read .ai/mcp.json
	canonical, err := jsonc.ReadJSONCFile(canonicalPath)
	if err != nil {
		checks = append(checks, CheckResult{
			Name:    "canonical mcp.json",
			Status:  CheckFail,
			Message: fmt.Sprintf(".ai/mcp.json is not valid JSON: %v", err),
		})
		return finalizeServerReport(name, checks)
	}

	// Check the server entry in .ai/mcp.json
	// The canonical mcp.json may have "servers" or "mcpServers" as the key
	serverMap, _ := canonical["servers"].(map[string]any)
	if serverMap == nil {
		serverMap, _ = canonical["mcpServers"].(map[string]any)
	}
	if serverMap == nil {
		serverMap = map[string]any{}
	}

	canonicalEntry, hasEntry := serverMap[name]
	if !hasEntry {
		checks = append(checks, CheckResult{
			Name:        "canonical mcp.json entry",
			Status:      CheckFail,
			Message:     fmt.Sprintf("'%s' missing from .ai/mcp.json", name),
			Remediation: fmt.Sprintf("Run 'lazyai-cli server add %s'", name),
		})
		return finalizeServerReport(name, checks)
	}

	// Check if the entry is enabled
	if entryMap, ok := canonicalEntry.(map[string]any); ok {
		if enabled, _ := entryMap["enabled"].(bool); !enabled {
			checks = append(checks, CheckResult{
				Name:        "canonical mcp.json entry",
				Status:      CheckFail,
				Message:     fmt.Sprintf("'%s' is present but not enabled in .ai/mcp.json", name),
				Remediation: fmt.Sprintf("Run 'lazyai-cli server add %s'", name),
			})
			return finalizeServerReport(name, checks)
		}
	}

	checks = append(checks, CheckResult{
		Name:    "canonical mcp.json entry",
		Status:  CheckPass,
		Message: fmt.Sprintf("'%s' enabled in .ai/mcp.json", name),
	})

	// L1.2: Check per-tool compiled configs
	for _, tool := range tools {
		relPath := perToolMCPConfig[string(tool)]
		if relPath == "" {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("%s mcp config", tool),
				Status:  CheckSkip,
				Message: fmt.Sprintf("%s has no project-local MCP config file (managed globally)", tool),
			})
			continue
		}

		absPath := filepath.Join(dir, relPath)
		if !files.FileExists(absPath) {
			checks = append(checks, CheckResult{
				Name:        fmt.Sprintf("%s mcp config", tool),
				Status:      CheckFail,
				Message:     fmt.Sprintf("%s is missing", relPath),
				Remediation: "Run 'lazyai-cli compile'",
			})
			continue
		}

		// Parse the per-tool config and check for the server entry
		raw, err := files.ReadFile(absPath)
		if err != nil {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("%s mcp config", tool),
				Status:  CheckFail,
				Message: fmt.Sprintf("%s is not readable: %v", relPath, err),
			})
			continue
		}

		parsed, err := jsonc.ParseJSONC(raw)
		if err != nil {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("%s mcp config", tool),
				Status:  CheckFail,
				Message: fmt.Sprintf("%s is not valid JSON: %v", relPath, err),
			})
			continue
		}

		// Check various key names (mcp, mcpServers, servers)
		toolServerMap := extractServerMap(parsed)
		if toolServerMap[name] != nil {
			checks = append(checks, CheckResult{
				Name:    fmt.Sprintf("%s mcp config", tool),
				Status:  CheckPass,
				Message: fmt.Sprintf("%s contains '%s'", relPath, name),
			})
		} else {
			checks = append(checks, CheckResult{
				Name:        fmt.Sprintf("%s mcp config", tool),
				Status:      CheckFail,
				Message:     fmt.Sprintf("%s exists but does not contain '%s'", relPath, name),
				Remediation: "Run 'lazyai-cli compile'",
			})
		}
	}

	// L1.3: Orchestrator-specific checks
	if name == "orchestrator" {
		chainsDir := filepath.Join(dir, ".ai", "orchestration", "chains")
		if files.FileExists(chainsDir) {
			checks = append(checks, CheckResult{
				Name:    "orchestration chains",
				Status:  CheckPass,
				Message: ".ai/orchestration/chains/ present",
			})
		} else {
			checks = append(checks, CheckResult{
				Name:        "orchestration chains",
				Status:      CheckFail,
				Message:     ".ai/orchestration/chains/ is missing",
				Remediation: "Run 'lazyai-cli server add orchestrator'",
			})
		}
	}

	// Note: L3 stdio handshake is skipped in the Go implementation.
	// The TypeScript version spawns the MCP server process and performs a
	// tools/list handshake, but this requires the Node.js MCP SDK.
	if entry.Command != "" {
		checks = append(checks, CheckResult{
			Name:    "stdio handshake",
			Status:  CheckSkip,
			Message: "stdio handshake not available in Go binary (requires Node.js MCP SDK)",
		})
	} else {
		checks = append(checks, CheckResult{
			Name:    "stdio handshake",
			Status:  CheckSkip,
			Message: "server has no stdio command (remote or url-based)",
		})
	}

	return finalizeServerReport(name, checks)
}

// extractServerMap tries to extract a server map from a parsed JSON object,
// checking common key names.
func extractServerMap(parsed map[string]any) map[string]any {
	for _, key := range []string{"mcp", "mcpServers", "servers"} {
		if m, ok := parsed[key].(map[string]any); ok {
			return m
		}
	}
	return map[string]any{}
}

// finalizeServerReport determines overall health from checks.
func finalizeServerReport(server string, checks []CheckResult) ServerHealthReport {
	hasFail := false
	hasPass := false
	for _, c := range checks {
		if c.Status == CheckFail {
			hasFail = true
		}
		if c.Status == CheckPass {
			hasPass = true
		}
	}

	overall := "partial"
	if hasFail {
		overall = "unhealthy"
	} else if hasPass {
		overall = "healthy"
	}

	return ServerHealthReport{
		Server:  server,
		Overall: overall,
		Checks:  checks,
	}
}

// boldStyle returns a bold lipgloss style (helper to avoid repetition).
func boldStyle() lipgloss.Style {
	return lipgloss.NewStyle().Bold(true)
}

// perToolMCPConfig maps each tool ID to its per-tool MCP config file path.
// An empty string means the tool uses a global config (no project-local file).
var perToolMCPConfig = map[string]string{
	"opencode":    ".opencode/opencode.jsonc",
	"claude-code": ".mcp.json",
	"copilot":     ".vscode/mcp.json",
}
