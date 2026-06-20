// Package adapter provides the MCP compiler that generates per-tool MCP
// configuration files from the canonical .ai/mcp.json.
// Ported from the TypeScript mcp-compiler.ts.
package adapter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/configmerge"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/jsonc"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// McpServer represents a single MCP server entry in the canonical config.
type McpServer struct {
	Description        string            `json:"description,omitempty"`
	Command            string            `json:"command,omitempty"`
	Args               []string          `json:"args,omitempty"`
	Env                map[string]string `json:"env,omitempty"`
	URL                string            `json:"url,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	Tools              []string          `json:"tools,omitempty"`
	Enabled            *bool             `json:"enabled,omitempty"`
	SetupHint          string            `json:"setupHint,omitempty"`
	PreferredInterface string            `json:"preferred_interface,omitempty"`
	CliEquivalent      string            `json:"cli_equivalent,omitempty"`
	TokenPolicy        string            `json:"token_policy,omitempty"`
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
	catalogRoot := mcpWorkspaceRoot(ctx)
	catalog := ReadCanonicalMcp(catalogRoot)
	if catalog == nil && catalogRoot != ctx.TargetDir {
		catalog = ReadCanonicalMcp(ctx.TargetDir)
	}
	if catalog == nil {
		return ctx.FileRecords, nil
	}

	enabledServers := GetEnabledServers(catalog)
	if len(enabledServers) == 0 {
		return ctx.FileRecords, nil
	}

	// Skip unsupported scopes.
	if !IsScopeSupported(toolId, ctx.SetupScope) {
		adapterLog.Info("scope is unsupported; skipping", "operation", "compile", "tool", toolId, "scope", ctx.SetupScope)
		return ctx.FileRecords, nil
	}

	// Copilot × global requires CLI or ~/.copilot/ to be present; at global
	// scope with probes failing we bail out of the whole compile. At
	// project/workspace the VS Code surface still writes regardless (the
	// CLI surface is gated internally in compileCopilotMCP).
	if toolId == types.ToolIdCopilot && ctx.SetupScope == types.SetupScopeGlobal {
		if !copilotProbePasses(ctx) {
			adapterLog.Info("copilot CLI or ~/.copilot/ not found; skipping global MCP compilation", "operation", "compile", "tool", types.ToolIdCopilot)
			return ctx.FileRecords, nil
		}
	}

	switch toolId {
	case types.ToolIdOpenCode:
		return compileOpenCodeMCP(ctx, catalog)
	case types.ToolIdClaudeCode:
		return compileClaudeCodeMCP(ctx, enabledServers)
	case types.ToolIdCopilot:
		return compileCopilotMCP(ctx, enabledServers)
	case types.ToolIdKiro:
		return compileKiroMCP(ctx, enabledServers)
	case types.ToolIdPi, types.ToolIdOmp, types.ToolIdAntigravity:
		return ctx.FileRecords, nil
	default:
		return ctx.FileRecords, fmt.Errorf("unsupported tool %q (supported tools: opencode, claude-code, copilot, pi, omp, kiro, antigravity)", toolId)
	}
}

// ReadCanonicalMcp reads and parses the canonical .ai/mcp.json.
func ReadCanonicalMcp(targetDir string) *McpCatalog {
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

// GetEnabledServers returns only the servers that are enabled.
func GetEnabledServers(catalog *McpCatalog) map[string]McpServer {
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
	// Global root is ~/.config/opencode — lazyai.mcp.jsonc lives directly in it.
	// Project/workspace root is <target>/.opencode — same placement.
	configPath := filepath.Join(root, OpenCodeRuntimeMCPFilename)
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
	recordPath, _ := filepath.Rel(mcpWorkspaceRoot(ctx), configPath)
	if recordPath == "" || recordPath == "." {
		recordPath = configPath
	}
	return append(ctx.FileRecords, types.TrackedFile{
		Path: recordPath, Hash: hash, Source: "compiled:mcp:opencode", Owner: types.FileOwnerLibrary,
	}), nil
}

// mergeOpenCodeMcpServers merges the ai-setup-managed mcp map into whatever
// the user currently has under `mcp` in their lazyai.mcp.jsonc. Managed
// servers (those present in `managed`) are upserted; any existing entry
// keyed by a name NOT in `managed` is preserved untouched — so a user who
// hand-adds an MCP server directly keeps it on the next `ai-setup compile`.
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
	// When LocalSecrets is set, route the project catalog to the gitignored
	// .claude/settings.local.json instead of the committed .mcp.json, and at
	// global scope write to ~/.claude/settings.local.json (ai-setup convention).
	if ctx.LocalSecrets {
		return writeClaudeSettingsLocal(ctx, servers)
	}

	// Default path: at global, skip (settings.json merge at init time covers it).
	if ctx.SetupScope == types.SetupScopeGlobal {
		adapterLog.Info("mcpServers live in settings.json; skipping", "operation", "compile", "tool", types.ToolIdClaudeCode, "scope", ctx.SetupScope)
		return ctx.FileRecords, nil
	}

	// Task 010: Try CLI-driven reconciliation if claude is on PATH.
	// Always write .mcp.json as the canonical project-scope config (backup/audit trail).
	useCliForMCP(ctx, servers)

	// Write .mcp.json for both CLI and fallback paths.
	root := mcpWorkspaceRoot(ctx)
	mcpPath := filepath.Join(root, ".mcp.json")
	content := toClaudeCodeMcp(servers)

	if err := WriteJSONFile(mcpPath, content); err != nil {
		return ctx.FileRecords, err
	}

	hash, _ := files.FileHash(mcpPath)
	return append(ctx.FileRecords, types.TrackedFile{
		Path: ".mcp.json", Hash: hash, Source: "compiled:mcp", Owner: types.FileOwnerLibrary,
	}), nil
}

// writeClaudeSettingsLocal routes the MCP catalog into the gitignored
// .claude/settings.local.json (project/workspace) or ~/.claude/settings.local.json
// (global, ai-setup convention). Uses configmerge.MergeJSONFile so user-authored
// keys are preserved and a .bak is created on first touch.
func writeClaudeSettingsLocal(ctx CompileContext, servers map[string]McpServer) ([]types.TrackedFile, error) {
	var settingsPath string
	switch ctx.SetupScope {
	case types.SetupScopeGlobal:
		home := ctx.HomeDir
		if home == "" {
			var err error
			home, err = os.UserHomeDir()
			if err != nil {
				return ctx.FileRecords, fmt.Errorf("resolve home dir: %w", err)
			}
		}
		settingsPath = filepath.Join(home, ".claude", "settings.local.json")
		adapterLog.Info("writing local secrets settings with ai-setup convention", "operation", "compile", "tool", types.ToolIdClaudeCode, "scope", ctx.SetupScope, "path", settingsPath)
	default:
		settingsPath = filepath.Join(mcpWorkspaceRoot(ctx), ".claude", "settings.local.json")
	}

	_ = files.EnsureDir(filepath.Dir(settingsPath))
	patch := map[string]any{"mcpServers": toClaudeCodeMcpInner(servers)}
	if _, err := configmerge.MergeJSONFile(settingsPath, patch); err != nil {
		return ctx.FileRecords, fmt.Errorf("merge %s: %w", settingsPath, err)
	}

	relPath := settingsPath
	if rel, err := filepath.Rel(mcpWorkspaceRoot(ctx), settingsPath); err == nil {
		relPath = rel
	}
	hash, _ := files.FileHash(settingsPath)
	return append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: "compiled:mcp:claude-local", Owner: types.FileOwnerLibrary,
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
	workingDir := mcpWorkspaceRoot(ctx)

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
			adapterLog.Info("server already registered, skipping", "operation", "compile-mcp", "server", name)
			continue
		}

		// Not found, add it via CLI.
		payload := mcpServerToJSON(srv)
		addCtx, cancel := context.WithTimeout(runCtx, 30*time.Second)
		_, stderr, err := runner.Run(addCtx, workingDir, "mcp", "add-json", name, payload, "-s", scopeFlag)
		cancel()
		if err != nil {
			adapterLog.Error("mcp add-json failed", "operation", "compile-mcp", "server", name, "error", err, "stderr", string(stderr))
			// CLI error — fall back to direct-write.
			return false
		}

		adapterLog.Info("registered server via CLI", "operation", "compile-mcp", "server", name)
	}

	return true
}

// toClaudeCodeMcpInner builds the raw mcpServers map without the top-level
// "mcpServers" wrapper. Shared between the committed .mcp.json output and the
// settings.local.json merge payload.
func toClaudeCodeMcpInner(servers map[string]McpServer) map[string]any {
	mcpServers := make(map[string]any)
	for name, server := range servers {
		if server.URL != "" {
			entry := map[string]any{
				"type": "http",
				"url":  server.URL,
			}
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
	return mcpServers
}

func toClaudeCodeMcp(servers map[string]McpServer) map[string]any {
	return map[string]any{"mcpServers": toClaudeCodeMcpInner(servers)}
}

// ---------------------------------------------------------------------------
// Kiro MCP compilation
// ---------------------------------------------------------------------------

func compileKiroMCP(ctx CompileContext, servers map[string]McpServer) ([]types.TrackedFile, error) {
	ctxAdapter := &AdapterContext{
		TargetDir:     ctx.TargetDir,
		HomeDir:       ctx.HomeDir,
		WorkspaceRoot: ctx.WorkspaceRoot,
		SetupScope:    ctx.SetupScope,
	}
	kiroRoot, err := ResolveToolRoot(types.ToolIdKiro, ctx.SetupScope, ctxAdapter)
	if err != nil {
		return nil, err
	}
	settingsDir := filepath.Join(kiroRoot, "settings")
	_ = files.EnsureDir(settingsDir)
	mcpPath := filepath.Join(settingsDir, "mcp.json")
	content := toClaudeCodeMcp(servers) // reuse identical format
	if err := WriteJSONFile(mcpPath, content); err != nil {
		return ctx.FileRecords, err
	}
	hash, _ := files.FileHash(mcpPath)
	return append(ctx.FileRecords, types.TrackedFile{
		Path:   mcpPath,
		Hash:   hash,
		Source: "compiled:mcp",
		Owner:  types.FileOwnerLibrary,
	}), nil
}
// ---------------------------------------------------------------------------
// Copilot MCP compilation
// ---------------------------------------------------------------------------

// copilotProbePasses returns true if the Copilot CLI is on PATH or the
// ~/.copilot/ directory exists (i.e., the standalone CLI is installed or has
// been used at least once).
func copilotProbePasses(ctx CompileContext) bool {
	if _, found := LookupCopilotBinary(); found {
		return true
	}
	home := ctx.HomeDir
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			return false
		}
	}
	return files.DirExists(filepath.Join(home, ".copilot"))
}

// resolveCopilotHomeDir returns the home directory that should hold
// ~/.copilot/mcp-config.json, preferring ctx.HomeDir for test isolation.
func resolveCopilotHomeDir(ctx CompileContext) (string, error) {
	if ctx.HomeDir != "" {
		return ctx.HomeDir, nil
	}
	return os.UserHomeDir()
}

func compileCopilotMCP(ctx CompileContext, servers map[string]McpServer) ([]types.TrackedFile, error) {
	// At project/workspace scope, emit the VS Code mcp.json unconditionally;
	// the CLI mcp-config.json is only written if the Copilot CLI probe passes.
	// At global scope, skip the VS Code emit entirely — no project directory
	// to anchor it to — and rely on the CLI emit (already probe-gated upstream).
	if ctx.SetupScope != types.SetupScopeGlobal {
		root := mcpWorkspaceRoot(ctx)
		vscodeDir := filepath.Join(root, ".vscode")
		_ = files.EnsureDir(vscodeDir)
		vscodeMcpPath := filepath.Join(vscodeDir, "mcp.json")
		content := toCopilotVSCodeMcp(servers)

		if err := WriteJSONFile(vscodeMcpPath, content); err != nil {
			return ctx.FileRecords, err
		}

		hash, _ := files.FileHash(vscodeMcpPath)
		ctx.FileRecords = append(ctx.FileRecords, types.TrackedFile{
			Path: ".vscode/mcp.json", Hash: hash, Source: "compiled:mcp:copilot", Owner: types.FileOwnerLibrary,
		})
	}

	// Emit ~/.copilot/mcp-config.json for the standalone CLI whenever the
	// probe passes (all scopes). At global scope the caller has already
	// verified the probe; at project/workspace we check here and silently
	// skip if Copilot CLI is not installed.
	if !copilotProbePasses(ctx) {
		return ctx.FileRecords, nil
	}
	return compileCopilotCLIMcp(ctx, servers)
}

// compileCopilotCLIMcp writes ~/.copilot/mcp-config.json via deep-merge so
// user-authored server entries and other keys are preserved across re-runs.
func compileCopilotCLIMcp(ctx CompileContext, servers map[string]McpServer) ([]types.TrackedFile, error) {
	home, err := resolveCopilotHomeDir(ctx)
	if err != nil {
		adapterLog.Error("cannot resolve home dir", "operation", "compile", "tool", types.ToolIdCopilot, "error", err)
		return ctx.FileRecords, nil
	}
	copilotDir := filepath.Join(home, ".copilot")
	_ = files.EnsureDir(copilotDir)
	cfgPath := filepath.Join(copilotDir, "mcp-config.json")

	patch := toCopilotCLIMcp(servers)
	if _, err := configmerge.MergeJSONFile(cfgPath, patch); err != nil {
		return ctx.FileRecords, fmt.Errorf("merge %s: %w", cfgPath, err)
	}

	relPath := cfgPath
	if rel, err := filepath.Rel(mcpWorkspaceRoot(ctx), cfgPath); err == nil {
		relPath = rel
	}
	hash, _ := files.FileHash(cfgPath)
	return append(ctx.FileRecords, types.TrackedFile{
		Path: relPath, Hash: hash, Source: "compiled:mcp:copilot-cli", Owner: types.FileOwnerLibrary,
	}), nil
}

// toCopilotServerEntries translates the canonical server map into the
// per-server shape shared by both the VS Code extension and the standalone
// Copilot CLI. It also returns the set of ${VAR} placeholder IDs discovered
// in env values (used by VS Code for its "inputs" prompt UI).
func toCopilotServerEntries(servers map[string]McpServer) (entries map[string]any, placeholderIDs map[string]bool) {
	entries = make(map[string]any)
	placeholderIDs = make(map[string]bool)

	for name, server := range servers {
		if server.URL != "" {
			entry := map[string]any{
				"type": "sse",
				"url":  server.URL,
			}
			if server.Headers != nil {
				entry["headers"] = server.Headers
			}
			entries[name] = entry
			continue
		}
		entry := map[string]any{
			"type":    "stdio",
			"command": server.Command,
			"args":    server.Args,
		}
		if server.Env != nil {
			entry["env"] = server.Env
			for _, val := range server.Env {
				matches := envPattern.FindAllStringSubmatch(val, -1)
				for _, match := range matches {
					if len(match) > 1 {
						placeholderIDs[match[1]] = true
					}
				}
			}
		}
		entries[name] = entry
	}
	return entries, placeholderIDs
}

// toCopilotVSCodeMcp builds the .vscode/mcp.json payload: uses the "servers"
// top-level key and adds an "inputs" prompt array for each placeholder.
func toCopilotVSCodeMcp(servers map[string]McpServer) map[string]any {
	entries, placeholderIDs := toCopilotServerEntries(servers)
	output := map[string]any{"servers": entries}

	if len(placeholderIDs) > 0 {
		inputs := make([]map[string]any, 0, len(placeholderIDs))
		ids := make([]string, 0, len(placeholderIDs))
		for id := range placeholderIDs {
			ids = append(ids, id)
		}
		sort.Strings(ids)

		for _, id := range ids {
			inputs = append(inputs, map[string]any{
				"type":        "promptString",
				"id":          id,
				"description": id,
				"password":    true,
			})
		}
		output["inputs"] = inputs
	}
	return output
}

// toCopilotCLIMcp builds the ~/.copilot/mcp-config.json payload: uses the
// "mcpServers" top-level key. Unlike VS Code, the standalone CLI reads
// env variables from the process environment directly — no "inputs" prompts.
func toCopilotCLIMcp(servers map[string]McpServer) map[string]any {
	entries, _ := toCopilotServerEntries(servers)
	return map[string]any{"mcpServers": entries}
}

// ---------------------------------------------------------------------------
// Shared helpers
// ---------------------------------------------------------------------------

var envPattern = regexp.MustCompile(`\$\{(\w+)\}`)

func mcpWorkspaceRoot(ctx CompileContext) string {
	if ctx.SetupScope == types.SetupScopeWorkspace && ctx.WorkspaceRoot != "" {
		return ctx.WorkspaceRoot
	}
	return ctx.TargetDir
}

// transformEnvSyntax replaces ${VAR} patterns with the target pattern.
func transformEnvSyntax(envObj map[string]string, targetPattern string) map[string]string {
	result := make(map[string]string, len(envObj))
	for key, value := range envObj {
		result[key] = envPattern.ReplaceAllString(value, targetPattern)
	}
	return result
}
