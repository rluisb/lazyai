// Package adapter provides the MCP compiler that generates per-tool MCP
// configuration files from the canonical .ai/mcp.json.
// Ported from the TypeScript mcp-compiler.ts.
package adapter

import (
	"context"
	"encoding/json"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/ricardoborges-teachable/ai-setup/internal/configmerge"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/jsonc"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// McpServer represents a single MCP server entry in the canonical config.
type McpServer struct {
	Description string            `json:"description,omitempty"`
	Command     string            `json:"command,omitempty"`
	Args        []string          `json:"args,omitempty"`
	Env         map[string]string `json:"env,omitempty"`
	URL         string            `json:"url,omitempty"`
	Headers     map[string]string `json:"headers,omitempty"`
	Tools       []string          `json:"tools,omitempty"`
	Enabled     *bool             `json:"enabled,omitempty"`
}

// McpCatalog represents the canonical .ai/mcp.json structure.
type McpCatalog struct {
	Servers map[string]McpServer `json:"servers"`
}

// isEnabled returns true if the server is explicitly enabled or the field is
// nil (default: enabled).
func (s McpServer) isEnabled() bool {
	if s.Enabled == nil {
		return true
	}
	return *s.Enabled
}

// CompileMCPForTool reads the canonical .ai/mcp.json and generates the
// tool-specific MCP configuration file. The ctx carries scope information so
// each adapter writes to the correct path per scope. Tools with no global
// layout (Copilot × global) are skipped cleanly.
func CompileMCPForTool(toolId types.ToolId, ctx CompileContext) ([]types.TrackedFile, error) {
	catalog := readCanonicalMcp(ctx.TargetDir)
	if catalog == nil {
		return ctx.FileRecords, nil
	}

	enabledServers := getEnabledServers(catalog)
	if len(enabledServers) == 0 {
		return ctx.FileRecords, nil
	}

	// Copilot × global is unsupported upstream — skip with no error.
	if !IsScopeSupported(toolId, ctx.SetupScope) {
		log.Printf("[compile] %s × %s scope is unsupported; skipping", toolId, ctx.SetupScope)
		return ctx.FileRecords, nil
	}

	switch toolId {
	case types.ToolIdOpenCode:
		return compileOpenCodeMCP(ctx, catalog)
	case types.ToolIdClaudeCode:
		return compileClaudeCodeMCP(ctx, enabledServers)
	case types.ToolIdCopilot:
		return compileCopilotMCP(ctx, enabledServers)
	case types.ToolIdGemini:
		return compileGeminiMCP(ctx, enabledServers)
	case types.ToolIdCodex:
		return compileCodexMCP(ctx, enabledServers)
	default:
		// Tool has no MCP config format — return unchanged.
		return ctx.FileRecords, nil
	}
}

// readCanonicalMcp reads and parses the canonical .ai/mcp.json.
func readCanonicalMcp(targetDir string) *McpCatalog {
	mcpPath := filepath.Join(targetDir, ".ai", "mcp.json")
	if !files.FileExists(mcpPath) {
		return nil
	}
	data, err := files.ReadFile(mcpPath)
	if err != nil {
		return nil
	}
	var catalog McpCatalog
	if err := json.Unmarshal(data, &catalog); err != nil {
		return nil
	}
	return &catalog
}

// getEnabledServers returns only the servers that are enabled.
func getEnabledServers(catalog *McpCatalog) map[string]McpServer {
	result := make(map[string]McpServer)
	for name, server := range catalog.Servers {
		if server.isEnabled() {
			result[name] = server
		}
	}
	return result
}

// ---------------------------------------------------------------------------
// OpenCode MCP compilation
// ---------------------------------------------------------------------------

func compileOpenCodeMCP(ctx CompileContext, catalog *McpCatalog) ([]types.TrackedFile, error) {
	root, err := ResolveToolRoot(types.ToolIdOpenCode, ctx.SetupScope, ctx.toAdapterContext())
	if err != nil {
		return ctx.FileRecords, err
	}
	// Global root is ~/.config/opencode — opencode.jsonc lives directly in it.
	// Project/workspace root is <target>/.opencode — same placement. The
	// filename constant is shared with OpenCodeAdapter.Install so install and
	// compile always target the same file (no .json/.jsonc split).
	configPath := filepath.Join(root, OpenCodeConfigFilename)
	ocMcp := toOpenCodeMcp(catalog.Servers)

	// Read existing config and merge.
	var existingConfig map[string]any
	if files.FileExists(configPath) {
		parsed, err := jsonc.ReadJSONCFile(configPath)
		if err == nil {
			existingConfig = parsed
		}
	}
	if existingConfig == nil {
		existingConfig = make(map[string]any)
	}

	existingConfig["$schema"] = "https://opencode.ai/config.json"
	existingConfig["mcp"] = mergeOpenCodeMcpServers(existingConfig["mcp"], ocMcp)

	if err := WriteJSONFile(configPath, existingConfig); err != nil {
		return ctx.FileRecords, err
	}

	hash, _ := files.FileHash(configPath)
	recordPath, _ := filepath.Rel(ctx.TargetDir, configPath)
	if recordPath == "" || recordPath == "." {
		recordPath = configPath
	}
	return append(ctx.FileRecords, types.TrackedFile{
		Path: recordPath, Hash: hash, Source: "compiled:mcp:opencode", Owner: types.FileOwnerLibrary,
	}), nil
}

// mergeOpenCodeMcpServers merges the ai-setup-managed mcp map into whatever
// the user currently has under `mcp` in their opencode.jsonc. Managed
// servers (those present in `managed`) are upserted; any existing entry
// keyed by a name NOT in `managed` is preserved untouched — so a user who
// hand-adds an MCP server directly in opencode.jsonc does not lose it on
// the next `ai-setup compile`.
//
// Documented limit: if a user hand-authors a server with the same name as a
// managed server, the managed definition wins (this matches the catalog's
// intent that `ai-setup` owns the named server).
func mergeOpenCodeMcpServers(existingRaw any, managed map[string]any) map[string]any {
	merged := make(map[string]any, len(managed))
	if existing, ok := existingRaw.(map[string]any); ok {
		for name, entry := range existing {
			if _, isManaged := managed[name]; isManaged {
				continue // skip; managed entry wins
			}
			merged[name] = entry
		}
	}
	for name, entry := range managed {
		merged[name] = entry
	}
	return merged
}

func toOpenCodeMcp(servers map[string]McpServer) map[string]any {
	mcp := make(map[string]any)
	for name, server := range servers {
		isEnabled := server.isEnabled()
		if server.URL != "" {
			entry := map[string]any{
				"type":    "remote",
				"enabled": isEnabled,
				"url":     server.URL,
			}
			if server.Headers != nil {
				entry["headers"] = transformEnvSyntax(server.Headers, "{env:$1}")
			}
			mcp[name] = entry
		} else {
			entry := map[string]any{
				"type":    "local",
				"enabled": isEnabled,
				"command": append([]string{server.Command}, server.Args...),
			}
			if server.Env != nil {
				entry["environment"] = transformEnvSyntax(server.Env, "{env:$1}")
			}
			mcp[name] = entry
		}
	}
	return mcp
}

// ---------------------------------------------------------------------------
// Claude Code MCP compilation
// ---------------------------------------------------------------------------

func compileClaudeCodeMCP(ctx CompileContext, servers map[string]McpServer) ([]types.TrackedFile, error) {
	// Claude Code has two distinct config surfaces:
	//   - project: <target>/.mcp.json (user-committed project-scope MCP file)
	//   - global:  ~/.claude/settings.json (mcpServers merged by init's Install)
	// At global scope we skip compile entirely — init has already merged the
	// canonical MCP list into settings.json via configmerge.MergeJSONFile.
	if ctx.SetupScope == types.SetupScopeGlobal {
		log.Println("[compile] claude-code × global: mcpServers live in settings.json; skipping")
		return ctx.FileRecords, nil
	}

	// Task 010: Try CLI-driven reconciliation if claude is on PATH.
	// Always write .mcp.json as the canonical project-scope config (backup/audit trail).
	useCliForMCP(ctx, servers)

	// Write .mcp.json for both CLI and fallback paths.
	mcpPath := filepath.Join(ctx.TargetDir, ".mcp.json")
	content := toClaudeCodeMcp(servers)

	if err := WriteJSONFile(mcpPath, content); err != nil {
		return ctx.FileRecords, err
	}

	hash, _ := files.FileHash(mcpPath)
	return append(ctx.FileRecords, types.TrackedFile{
		Path: ".mcp.json", Hash: hash, Source: "compiled:mcp", Owner: types.FileOwnerLibrary,
	}), nil
}

// useCliForMCP attempts to reconcile MCP servers via the claude CLI.
// Returns true if reconciliation succeeded, false if CLI unavailable or error (fallback will run).
func useCliForMCP(ctx CompileContext, servers map[string]McpServer) bool {
	_, found := LookupClaudeBinary()
	if !found {
		return false
	}

	runner := &DefaultClaudeCLIRunner{}
	runCtx := context.Background()

	// For project/workspace scopes, use -s project with the target dir as working dir.
	scopeFlag := "project"
	workingDir := ctx.TargetDir

	// Register all enabled servers via CLI.
	for name, srv := range servers {
		if !srv.isEnabled() {
			continue
		}

		// Pre-check: is server already registered?
		checkCtx, cancel := context.WithTimeout(runCtx, 10*time.Second)
		_, _, err := runner.Run(checkCtx, workingDir, "mcp", "get", name)
		cancel()
		if err == nil {
			// Already registered, skip it.
			log.Printf("[compile-mcp] server %q already registered, skipping", name)
			continue
		}

		// Not found, add it via CLI.
		payload := mcpServerToJSON(srv)
		addCtx, cancel := context.WithTimeout(runCtx, 30*time.Second)
		_, stderr, err := runner.Run(addCtx, workingDir, "mcp", "add-json", name, payload, "-s", scopeFlag)
		cancel()
		if err != nil {
			log.Printf("[compile-mcp] mcp add-json %q failed: %v\nstderr: %s", name, err, string(stderr))
			// CLI error — fall back to direct-write.
			return false
		}

		log.Printf("[compile-mcp] registered server %q via CLI", name)
	}

	return true
}

func toClaudeCodeMcp(servers map[string]McpServer) map[string]any {
	mcpServers := make(map[string]any)
	for name, server := range servers {
		if server.URL != "" {
			entry := map[string]any{"url": server.URL}
			if server.Headers != nil {
				entry["headers"] = server.Headers
			}
			mcpServers[name] = entry
			continue
		}
		entry := map[string]any{
			"command": server.Command,
			"args":    server.Args,
		}
		if server.Env != nil {
			entry["env"] = server.Env
		}
		mcpServers[name] = entry
	}
	return map[string]any{"mcpServers": mcpServers}
}

// ---------------------------------------------------------------------------
// Copilot MCP compilation
// ---------------------------------------------------------------------------

func compileCopilotMCP(ctx CompileContext, servers map[string]McpServer) ([]types.TrackedFile, error) {
	// Copilot × global is filtered upstream in CompileMCPForTool; this code
	// path always runs at project or workspace scope.
	vscodeDir := filepath.Join(ctx.TargetDir, ".vscode")
	_ = files.EnsureDir(vscodeDir)
	vscodeMcpPath := filepath.Join(vscodeDir, "mcp.json")
	content := toCopilotMcp(servers)

	if err := WriteJSONFile(vscodeMcpPath, content); err != nil {
		return ctx.FileRecords, err
	}

	hash, _ := files.FileHash(vscodeMcpPath)
	return append(ctx.FileRecords, types.TrackedFile{
		Path: ".vscode/mcp.json", Hash: hash, Source: "compiled:mcp:copilot", Owner: types.FileOwnerLibrary,
	}), nil
}

func toCopilotMcp(servers map[string]McpServer) map[string]any {
	result := make(map[string]any)
	for name, server := range servers {
		if server.URL != "" {
			entry := map[string]any{
				"type": "sse",
				"url":  server.URL,
			}
			if server.Headers != nil {
				entry["headers"] = server.Headers
			}
			result[name] = entry
			continue
		}
		entry := map[string]any{
			"type":    "stdio",
			"command": server.Command,
			"args":    server.Args,
		}
		if server.Env != nil {
			entry["env"] = server.Env
		}
		result[name] = entry
	}
	return map[string]any{"servers": result}
}

// ---------------------------------------------------------------------------
// Gemini MCP compilation
// ---------------------------------------------------------------------------

func compileGeminiMCP(ctx CompileContext, servers map[string]McpServer) ([]types.TrackedFile, error) {
	root, err := ResolveToolRoot(types.ToolIdGemini, ctx.SetupScope, ctx.toAdapterContext())
	if err != nil {
		return ctx.FileRecords, err
	}
	_ = files.EnsureDir(root)
	settingsPath := filepath.Join(root, "settings.json")
	content := toGeminiSettings(servers)

	if err := WriteJSONFile(settingsPath, content); err != nil {
		return ctx.FileRecords, err
	}

	hash, _ := files.FileHash(settingsPath)
	recordPath, _ := filepath.Rel(ctx.TargetDir, settingsPath)
	if recordPath == "" || recordPath == "." {
		recordPath = settingsPath
	}
	return append(ctx.FileRecords, types.TrackedFile{
		Path: recordPath, Hash: hash, Source: "compiled:mcp:gemini", Owner: types.FileOwnerLibrary,
	}), nil
}

func toGeminiSettings(servers map[string]McpServer) map[string]any {
	mcpServers := make(map[string]any)
	for name, server := range servers {
		if server.URL != "" {
			log.Printf("Skipping remote server %q for gemini (not supported)", name)
			continue
		}
		entry := map[string]any{
			"command": server.Command,
			"args":    server.Args,
		}
		if server.Env != nil {
			entry["env"] = transformEnvSyntax(server.Env, "$$$1")
		}
		mcpServers[name] = entry
	}
	return map[string]any{"mcpServers": mcpServers}
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

var envPattern = regexp.MustCompile(`\$\{(\w+)\}`)

// transformEnvSyntax replaces ${VAR} patterns with the target pattern.
func transformEnvSyntax(envObj map[string]string, targetPattern string) map[string]string {
	result := make(map[string]string, len(envObj))
	for key, value := range envObj {
		result[key] = envPattern.ReplaceAllString(value, targetPattern)
	}
	return result
}

// ---------------------------------------------------------------------------
// Codex MCP compilation
// ---------------------------------------------------------------------------

// compileCodexMCP writes enabled MCP servers into the scope-correct config.toml
// under the [mcp_servers.*] tables using deep-merge so user-authored keys survive.
func compileCodexMCP(ctx CompileContext, enabledServers map[string]McpServer) ([]types.TrackedFile, error) {
	configRoot, _, err := ResolveCodexRoots(ctx.SetupScope, ctx.toAdapterContext())
	if err != nil {
		return ctx.FileRecords, err
	}
	configPath := filepath.Join(configRoot, "config.toml")

	mcpServers := make(map[string]any, len(enabledServers))
	for name, srv := range enabledServers {
		entry := map[string]any{}
		if srv.Command != "" {
			entry["command"] = srv.Command
		}
		if len(srv.Args) > 0 {
			entry["args"] = srv.Args
		}
		if len(srv.Env) > 0 {
			entry["env"] = srv.Env
		}
		mcpServers[name] = entry
	}

	patch := map[string]any{"mcp_servers": mcpServers}
	if _, err := configmerge.MergeTOMLFile(configPath, patch); err != nil {
		return ctx.FileRecords, err
	}

	hash, _ := files.FileHash(configPath)
	recordPath, _ := filepath.Rel(ctx.TargetDir, configPath)
	if recordPath == "" || recordPath == "." {
		recordPath = configPath
	}
	log.Printf("[codex] compiled MCP config -> %s (%d servers)", configPath, len(enabledServers))
	return append(ctx.FileRecords, types.TrackedFile{
		Path: recordPath, Hash: hash, Source: "compiled:mcp:codex", Owner: types.FileOwnerLibrary,
	}), nil
}

// Ensure all format strings with %s in log statements are correct.
// This blank import ensures strings is available.
var _ = strings.TrimSpace
