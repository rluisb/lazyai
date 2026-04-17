// Package adapter provides the MCP compiler that generates per-tool MCP
// configuration files from the canonical .ai/mcp.json.
// Ported from the TypeScript mcp-compiler.ts.
package adapter

import (
	"encoding/json"
	"log"
	"path/filepath"
	"regexp"
	"strings"

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
// tool-specific MCP configuration file. It returns updated file records.
func CompileMCPForTool(toolId types.ToolId, targetDir string, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	catalog := readCanonicalMcp(targetDir)
	if catalog == nil {
		return fileRecords, nil
	}

	enabledServers := getEnabledServers(catalog)
	if len(enabledServers) == 0 {
		return fileRecords, nil
	}

	switch toolId {
	case types.ToolIdOpenCode:
		return compileOpenCodeMCP(targetDir, catalog, fileRecords)
	case types.ToolIdClaudeCode:
		return compileClaudeCodeMCP(targetDir, enabledServers, fileRecords)
	case types.ToolIdCopilot:
		return compileCopilotMCP(targetDir, enabledServers, fileRecords)
	case types.ToolIdGemini:
		return compileGeminiMCP(targetDir, enabledServers, fileRecords)
	default:
		// Tool has no MCP config format — return unchanged.
		return fileRecords, nil
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

func compileOpenCodeMCP(targetDir string, catalog *McpCatalog, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	// In global scope, targetDir is already ~/.config/opencode, so write directly.
	// In project scope, targetDir is the project root, so write to .opencode/.
	configPath := filepath.Join(targetDir, ".opencode", "opencode.jsonc")
	if isGlobalOpenCodeDir(targetDir) {
		configPath = filepath.Join(targetDir, "opencode.jsonc")
	}
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
	existingConfig["mcp"] = ocMcp

	if err := WriteJSONFile(configPath, existingConfig); err != nil {
		return fileRecords, err
	}

	hash, _ := files.FileHash(configPath)
	recordPath := ".opencode/opencode.jsonc"
	if isGlobalOpenCodeDir(targetDir) {
		recordPath = "opencode.jsonc"
	}
	fileRecords = append(fileRecords, types.TrackedFile{
		Path: recordPath, Hash: hash, Source: "compiled:mcp:opencode", Owner: types.FileOwnerLibrary,
	})
	return fileRecords, nil
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

func compileClaudeCodeMCP(targetDir string, servers map[string]McpServer, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	mcpPath := filepath.Join(targetDir, ".mcp.json")
	content := toClaudeCodeMcp(servers)

	if err := WriteJSONFile(mcpPath, content); err != nil {
		return fileRecords, err
	}

	hash, _ := files.FileHash(mcpPath)
	fileRecords = append(fileRecords, types.TrackedFile{
		Path: ".mcp.json", Hash: hash, Source: "compiled:mcp", Owner: types.FileOwnerLibrary,
	})
	return fileRecords, nil
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

func compileCopilotMCP(targetDir string, servers map[string]McpServer, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	vscodeDir := filepath.Join(targetDir, ".vscode")
	_ = files.EnsureDir(vscodeDir)
	vscodeMcpPath := filepath.Join(vscodeDir, "mcp.json")
	content := toCopilotMcp(servers)

	if err := WriteJSONFile(vscodeMcpPath, content); err != nil {
		return fileRecords, err
	}

	hash, _ := files.FileHash(vscodeMcpPath)
	fileRecords = append(fileRecords, types.TrackedFile{
		Path: ".vscode/mcp.json", Hash: hash, Source: "compiled:mcp:copilot", Owner: types.FileOwnerLibrary,
	})
	return fileRecords, nil
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

func compileGeminiMCP(targetDir string, servers map[string]McpServer, fileRecords []types.TrackedFile) ([]types.TrackedFile, error) {
	geminiDir := filepath.Join(targetDir, ".gemini")
	_ = files.EnsureDir(geminiDir)
	settingsPath := filepath.Join(geminiDir, "settings.json")
	content := toGeminiSettings(servers)

	if err := WriteJSONFile(settingsPath, content); err != nil {
		return fileRecords, err
	}

	hash, _ := files.FileHash(settingsPath)
	fileRecords = append(fileRecords, types.TrackedFile{
		Path: ".gemini/settings.json", Hash: hash, Source: "compiled:mcp:gemini", Owner: types.FileOwnerLibrary,
	})
	return fileRecords, nil
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

// isGlobalOpenCodeDir returns true if the given directory looks like the global
// OpenCode config directory (~/.config/opencode).
func isGlobalOpenCodeDir(dir string) bool {
	return filepath.Base(dir) == "opencode" && filepath.Base(filepath.Dir(dir)) == ".config"
}

// Ensure all format strings with %s in log statements are correct.
// This blank import ensures strings is available.
var _ = strings.TrimSpace
