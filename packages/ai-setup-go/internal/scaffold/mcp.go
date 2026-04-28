package scaffold

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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
	Description     string            `json:"description,omitempty"`
	Command         string            `json:"command,omitempty"`
	Args            []string          `json:"args,omitempty"`
	Enabled         *bool             `json:"enabled,omitempty"`
	Env             map[string]string `json:"env,omitempty"`
	URL             string            `json:"url,omitempty"`
	Headers         map[string]string `json:"headers,omitempty"`
	Tools           []string          `json:"tools,omitempty"`
	RequiresInstall bool              `json:"requiresInstall,omitempty"`
	InstallHint     string            `json:"installHint,omitempty"`
}

var orchestratorLookPath = exec.LookPath

var orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

// ScaffoldMcp scaffolds .ai/mcp.json from the library catalog, enabling
// selected servers. Ported from src/scaffold/mcp.ts.
func ScaffoldMcp(targetDir, libraryDir string, libFS fs.FS, cliTools, enableServers []string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
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

	if enabledServerNames["orchestrator"] {
		server, err := prepareManagedOrchestratorServer(libraryDir, catalog.Servers["orchestrator"])
		if err != nil {
			return err
		}
		catalog.Servers["orchestrator"] = server
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

	// Also write .mcp.json at project root for tools that expect it there.
	rootDest := filepath.Join(targetDir, ".mcp.json")
	if rootAction, err := conflict.ApplyStrategy(rootDest, strategy, perFileOverrides, targetDir); err == nil && rootAction != "skip" {
		if err := files.WriteFile(rootDest, content, 0o644); err == nil {
			hash, _ := files.FileHash(rootDest)
			rootRel, _ := filepath.Rel(targetDir, rootDest)
			*fileRecords = append(*fileRecords, types.TrackedFile{
				Path:   filepath.ToSlash(rootRel),
				Hash:   hash,
				Source: catalogRelPath,
				Owner:  types.FileOwnerLibrary,
			})
		}
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

func prepareManagedOrchestratorServer(libraryDir string, server mcpServer) (mcpServer, error) {
	packageDir, ok := resolveOrchestratorPackageDir(libraryDir)
	if !ok {
		return server, nil
	}

	nodePath, err := orchestratorLookPath("node")
	if err != nil {
		return server, fmt.Errorf("prepare orchestrator MCP: node executable not found: %w", err)
	}

	entryPath := filepath.Join(packageDir, "dist", "index.js")
	if !files.FileExists(entryPath) {
		npmPath, err := orchestratorLookPath("npm")
		if err != nil {
			return server, fmt.Errorf("prepare orchestrator MCP: npm executable not found: %w", err)
		}
		if err := runOrchestratorBuildStep(npmPath, packageDir, "install"); err != nil {
			return server, err
		}
		if err := runOrchestratorBuildStep(npmPath, packageDir, "run", "build"); err != nil {
			return server, err
		}
	}
	if !files.FileExists(entryPath) {
		return server, fmt.Errorf("prepare orchestrator MCP: expected build output %s", entryPath)
	}

	if strings.EqualFold(os.Getenv("AI_SETUP_ORCHESTRATOR_SMOKE"), "1") || strings.EqualFold(os.Getenv("AI_SETUP_ORCHESTRATOR_SMOKE"), "true") {
		if err := runOrchestratorSmokeTest(nodePath, entryPath); err != nil {
			return server, err
		}
	}

	server.Command = nodePath
	server.Args = []string{entryPath}
	server.RequiresInstall = false
	server.InstallHint = ""
	return server, nil
}

func resolveOrchestratorPackageDir(libraryDir string) (string, bool) {
	if libraryDir == "" {
		return "", false
	}
	repoRoot := filepath.Dir(libraryDir)
	packageDir := filepath.Join(repoRoot, "orchestrator")
	if !files.FileExists(filepath.Join(packageDir, "package.json")) {
		return "", false
	}
	return packageDir, true
}

func runOrchestratorBuildStep(npmPath, packageDir string, args ...string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	runArgs := append([]string{"--prefix", packageDir}, args...)
	if err := orchestratorCommandRunner(ctx, npmPath, runArgs...); err != nil {
		return fmt.Errorf("prepare orchestrator MCP: npm %s failed: %w", strings.Join(args, " "), err)
	}
	return nil
}

func runOrchestratorSmokeTest(nodePath, entryPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := orchestratorCommandRunner(ctx, nodePath, entryPath, "catalog"); err != nil {
		return fmt.Errorf("prepare orchestrator MCP: smoke test failed: %w", err)
	}
	return nil
}
