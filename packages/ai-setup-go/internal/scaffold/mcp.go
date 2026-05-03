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
	"runtime"
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
	if err := json.Unmarshal(data, &catalog); err != nil {
		return err
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
		server, err := prepareManagedOrchestratorServer(targetDir, libraryDir, catalog.Servers["orchestrator"])
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
		rootContent, err := json.MarshalIndent(struct {
			MCPServers map[string]mcpServer `json:"mcpServers"`
		}{MCPServers: catalog.Servers}, "", "  ")
		if err != nil {
			return err
		}
		rootContent = append(rootContent, '\n')
		if err := files.WriteFile(rootDest, rootContent, 0o644); err == nil {
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

func prepareManagedOrchestratorServer(targetDir, libraryDir string, server mcpServer) (mcpServer, error) {
	packageDir, ok := resolveOrchestratorPackageDir(libraryDir)
	if !ok {
		if binaryPath, err := orchestratorLookPath("ai-setup-orchestrator"); err == nil {
			server.Command = binaryPath
			server.Args = orchestratorConnectArgs(targetDir)
			server.RequiresInstall = false
			server.InstallHint = ""
		}
		return server, nil
	}

	goPath, err := orchestratorLookPath("go")
	if err != nil {
		return server, fmt.Errorf("prepare orchestrator MCP: go executable not found: %w", err)
	}

	binaryPath, err := orchestratorBinaryPath()
	if err != nil {
		return server, err
	}
	if err := runOrchestratorBuildStep(goPath, packageDir, binaryPath); err != nil {
		return server, err
	}
	if !files.FileExists(binaryPath) {
		return server, fmt.Errorf("prepare orchestrator MCP: expected build output %s", binaryPath)
	}

	if strings.EqualFold(os.Getenv("AI_SETUP_ORCHESTRATOR_SMOKE"), "1") || strings.EqualFold(os.Getenv("AI_SETUP_ORCHESTRATOR_SMOKE"), "true") {
		if err := runOrchestratorSmokeTest(binaryPath); err != nil {
			return server, err
		}
	}

	server.Command = binaryPath
	server.Args = orchestratorConnectArgs(targetDir)
	server.RequiresInstall = false
	server.InstallHint = ""
	return server, nil
}

func resolveOrchestratorPackageDir(libraryDir string) (string, bool) {
	if libraryDir == "" {
		return "", false
	}
	repoRoot := filepath.Dir(libraryDir)
	packageDir := filepath.Join(repoRoot, "packages", "orchestrator-go")
	if !files.FileExists(filepath.Join(packageDir, "go.mod")) {
		return "", false
	}
	return packageDir, true
}

func orchestratorConnectArgs(targetDir string) []string {
	args := []string{"connect"}
	if targetDir != "" {
		args = append(args, "--project", targetDir)
	}
	return args
}

func orchestratorBinaryPath() (string, error) {
	cacheDir, err := os.UserCacheDir()
	if err != nil || cacheDir == "" {
		cacheDir = os.TempDir()
	}
	outDir := filepath.Join(cacheDir, "ai-setup", "orchestrator")
	if err := files.EnsureDir(outDir); err != nil {
		return "", err
	}
	binaryName := "ai-setup-orchestrator"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	return filepath.Join(outDir, binaryName), nil
}

func runOrchestratorBuildStep(goPath, packageDir, binaryPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	runArgs := []string{"-C", packageDir, "build", "-o", binaryPath, "./cmd/orchestrator"}
	if err := orchestratorCommandRunner(ctx, goPath, runArgs...); err != nil {
		return fmt.Errorf("prepare orchestrator MCP: go build failed: %w", err)
	}
	return nil
}

func runOrchestratorSmokeTest(binaryPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := orchestratorCommandRunner(ctx, binaryPath, "status"); err != nil {
		return fmt.Errorf("prepare orchestrator MCP: smoke test failed: %w", err)
	}
	return nil
}
