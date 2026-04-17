package scaffold

import (
	"encoding/json"
	"io/fs"
	"log"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/conflict"
	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// mcpCatalog represents the MCP server catalog structure.
type mcpCatalog struct {
	Servers map[string]mcpServer `json:"servers"`
}

// mcpServer represents a single MCP server entry.
type mcpServer struct {
	Enabled *bool             `json:"enabled,omitempty"`
	Env     map[string]string `json:"env,omitempty"`
}

// ScaffoldMcp scaffolds .ai/mcp.json from the library catalog, enabling
// selected servers. Ported from src/scaffold/mcp.ts.
func ScaffoldMcp(targetDir string, libFS fs.FS, cliTools, enableServers []string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	aiDir := filepath.Join(targetDir, ".ai")
	if err := files.EnsureDir(aiDir); err != nil {
		return err
	}

	catalogRelPath := "mcp/catalog.json"
	if !files.ExistsFS(libFS, catalogRelPath) {
		return nil
	}

	dest := filepath.Join(aiDir, "mcp.json")
	relPath, err := filepath.Rel(targetDir, dest)
	if err != nil {
		relPath = dest
	}

	action, err := conflict.ApplyStrategy(dest, strategy, perFileOverrides, targetDir)
	if err != nil {
		return err
	}
	if action == "skip" {
		log.Printf("Skipping existing file: %s", relPath)
		return nil
	}

	data, err := files.ReadFS(libFS, catalogRelPath)
	if err != nil {
		return err
	}

	var catalog mcpCatalog
	if err := json.Unmarshal(data, &catalog.Servers); err != nil {
		// Try wrapping in a top-level servers key
		var wrapper struct {
			Servers map[string]mcpServer `json:"servers"`
		}
		if err2 := json.Unmarshal(data, &wrapper); err2 != nil {
			return err2
		}
		catalog = wrapper
	}

	enabledServerNames := make(map[string]bool)

	// Enable CLI tools that are also MCP servers (legacy behavior).
	for _, toolName := range cliTools {
		enabledServerNames[toolName] = true
	}

	// Enable explicitly selected MCP servers.
	for _, serverName := range enableServers {
		enabledServerNames[serverName] = true
	}

	for serverName := range enabledServerNames {
		if server, ok := catalog.Servers[serverName]; ok {
			enabled := true
			server.Enabled = &enabled
			catalog.Servers[serverName] = server
		}
	}

	content, err := json.MarshalIndent(catalog, "", "  ")
	if err != nil {
		return err
	}
	content = append(content, '\n')

	if err := files.WriteFile(dest, content, 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	*fileRecords = append(*fileRecords, types.TrackedFile{
		Path:   relPath,
		Hash:   hash,
		Source: catalogRelPath,
		Owner:  types.FileOwnerLibrary,
	})

	return nil
}