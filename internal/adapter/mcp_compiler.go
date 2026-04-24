// Package adapter provides the MCP compiler that generates per-tool MCP
// configuration files from the canonical .ai/mcp.json.
package adapter

import (
	"encoding/json"
	"log"
	"path/filepath"
	"regexp"

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
// tool-specific MCP configuration file.
func CompileMCPForTool(toolId types.ToolId, ctx CompileContext) ([]types.TrackedFile, error) {
	catalog := readCanonicalMcp(ctx.TargetDir)
	if catalog == nil {
		return ctx.FileRecords, nil
	}

	if !IsScopeSupported(toolId, ctx.SetupScope) {
		log.Printf("[compile] %s × %s scope is unsupported; skipping", toolId, ctx.SetupScope)
		return ctx.FileRecords, nil
	}

	switch toolId {
	case types.ToolIdOpenCode:
		return compileOpenCodeMCP(ctx, catalog)
	default:
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

// ---------------------------------------------------------------------------
// OpenCode MCP compilation
// ---------------------------------------------------------------------------

func compileOpenCodeMCP(ctx CompileContext, catalog *McpCatalog) ([]types.TrackedFile, error) {
	// Config always lives inside the tool directory:
	//   project/workspace → {targetDir}/.opencode/opencode.jsonc
	//   global            → ~/.config/opencode/opencode.jsonc
	root, err := ResolveToolRoot(types.ToolIdOpenCode, ctx.SetupScope, ctx.toAdapterContext())
	if err != nil {
		return ctx.FileRecords, err
	}
	configPath := filepath.Join(root, OpenCodeConfigFilename)

	ocMcp := toOpenCodeMcp(catalog.Servers)

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
// the user currently has under `mcp` in their opencode.jsonc. Managed servers
// win on key collision; user-authored servers not in the managed set are kept.
func mergeOpenCodeMcpServers(existingRaw any, managed map[string]any) map[string]any {
	merged := make(map[string]any, len(managed))
	if existing, ok := existingRaw.(map[string]any); ok {
		for name, entry := range existing {
			if _, isManaged := managed[name]; isManaged {
				continue
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
// Shared helpers
// ---------------------------------------------------------------------------

var envPattern = regexp.MustCompile(`\$\{(\w+)\}`)

func transformEnvSyntax(envObj map[string]string, targetPattern string) map[string]string {
	result := make(map[string]string, len(envObj))
	for key, value := range envObj {
		result[key] = envPattern.ReplaceAllString(value, targetPattern)
	}
	return result
}
