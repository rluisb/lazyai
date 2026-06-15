package scaffold

import (
	"encoding/json"
	"io/fs"
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

type mcpCatalog struct {
	Servers map[string]mcpServer `json:"servers"`
}

type mcpServer struct {
	Description        string            `json:"description,omitempty"`
	Command            string            `json:"command,omitempty"`
	Args               []string          `json:"args,omitempty"`
	Enabled            *bool             `json:"enabled,omitempty"`
	Env                map[string]string `json:"env,omitempty"`
	URL                string            `json:"url,omitempty"`
	Headers            map[string]string `json:"headers,omitempty"`
	Tools              []string          `json:"tools,omitempty"`
	RequiresInstall    bool              `json:"requiresInstall,omitempty"`
	InstallHint        string            `json:"installHint,omitempty"`
	PreferredInterface string            `json:"preferred_interface,omitempty"`
	CliEquivalent      string            `json:"cli_equivalent,omitempty"`
	TokenPolicy        string            `json:"token_policy,omitempty"`
}

// ScaffoldMcp scaffolds .ai/mcp.json from the library catalog, enabling
// selected servers. It keeps the catalog generic: no server receives bespoke
// bootstrap logic.
func ScaffoldMcp(targetDir, libraryDir string, libFS fs.FS, cliTools, enableServers []string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	_ = libraryDir

	aiDir := filepath.Join(targetDir, ".ai")
	if err := files.EnsureDir(aiDir); err != nil {
		return err
	}

	const catalogRelPath = "mcp/catalog.json"
	if !files.ExistsFS(libFS, catalogRelPath) {
		return nil
	}

	catalogData, err := files.ReadFS(libFS, catalogRelPath)
	if err != nil {
		return err
	}

	var catalog mcpCatalog
	if err := json.Unmarshal(catalogData, &catalog); err != nil {
		return err
	}

	enabledServerNames := make(map[string]struct{})
	for _, toolName := range cliTools {
		enabledServerNames[toolName] = struct{}{}
	}
	for _, serverName := range enableServers {
		enabledServerNames[serverName] = struct{}{}
	}
	for serverName := range enabledServerNames {
		server, ok := catalog.Servers[serverName]
		if !ok {
			continue
		}
		enabled := true
		server.Enabled = &enabled
		catalog.Servers[serverName] = server
	}

	internalDest := filepath.Join(aiDir, "mcp.json")
	if err := writeManagedMcpFile(targetDir, internalDest, catalog, catalogRelPath, fileRecords, strategy, perFileOverrides, false); err != nil {
		return err
	}

	rootDest := filepath.Join(targetDir, ".mcp.json")
	return writeManagedMcpFile(targetDir, rootDest, catalog, catalogRelPath, fileRecords, strategy, perFileOverrides, true)
}

func writeManagedMcpFile(targetDir, dest string, catalog mcpCatalog, source string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy, wrapRoot bool) error {
	action, err := conflict.ApplyStrategy(dest, strategy, perFileOverrides, targetDir)
	if err != nil {
		return err
	}
	if action == "skip" {
		return nil
	}

	var content []byte
	if wrapRoot {
		content, err = json.MarshalIndent(struct {
			MCPServers map[string]mcpServer `json:"mcpServers"`
		}{MCPServers: catalog.Servers}, "", "  ")
	} else {
		content, err = json.MarshalIndent(catalog, "", "  ")
	}
	if err != nil {
		return err
	}
	content = append(content, '\n')
	if err := files.WriteFile(dest, content, 0o644); err != nil {
		return err
	}

	hash, _ := files.FileHash(dest)
	relPath, err := filepath.Rel(targetDir, dest)
	if err != nil {
		relPath = dest
	}
	*fileRecords = append(*fileRecords, types.TrackedFile{
		Path:   filepath.ToSlash(relPath),
		Hash:   hash,
		Source: source,
		Owner:  types.FileOwnerLibrary,
	})
	return nil
}
