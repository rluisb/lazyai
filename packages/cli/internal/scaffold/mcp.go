package scaffold

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/conflict"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
	buildversion "github.com/rluisb/lazyai/packages/cli/internal/version"
)

const orchestratorLatestReleaseURL = "https://api.github.com/repos/rluisb/lazyai/releases/latest"
const orchestratorReleaseByTagURLFormat = "https://api.github.com/repos/rluisb/lazyai/releases/tags/%s"

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

var orchestratorUserCacheDir = os.UserCacheDir

var orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
	cmd := exec.CommandContext(ctx, name, args...)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

var orchestratorReleaseFetcher = fetchOrchestratorRelease

var orchestratorAssetDownloader = downloadOrchestratorAsset

var orchestratorHTTPClient = http.DefaultClient

type orchestratorRelease struct {
	TagName string                     `json:"tag_name"`
	Assets  []orchestratorReleaseAsset `json:"assets"`
}

type orchestratorReleaseAsset struct {
	Name               string `json:"name"`
	BrowserDownloadURL string `json:"browser_download_url"`
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
		if binaryPath, err := orchestratorLookPath("lazyai-orchestrator"); err == nil {
			server.Command = binaryPath
			server.Args = orchestratorConnectArgs(targetDir)
			server.RequiresInstall = false
			server.InstallHint = ""
			return server, nil
		}

		binaryPath, err := prepareDownloadedOrchestratorBinary()
		if err != nil {
			return server, err
		}
		server.Command = binaryPath
		server.Args = orchestratorConnectArgs(targetDir)
		server.RequiresInstall = false
		server.InstallHint = ""
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

func prepareDownloadedOrchestratorBinary() (string, error) {
	assetName, err := orchestratorReleaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		return "", err
	}

	releaseURL := orchestratorReleaseURL(buildversion.Version)
	release, err := orchestratorReleaseFetcher(releaseURL)
	if err != nil {
		return "", err
	}
	if release == nil {
		return "", fmt.Errorf("prepare orchestrator MCP: release response was empty")
	}
	if release.TagName == "" {
		return "", fmt.Errorf("prepare orchestrator MCP: release did not include a tag_name")
	}

	binaryAsset, ok := findOrchestratorReleaseAsset(release.Assets, assetName)
	if !ok {
		return "", fmt.Errorf("prepare orchestrator MCP: release %s does not include asset %q", release.TagName, assetName)
	}
	checksumsAsset, ok := findOrchestratorReleaseAsset(release.Assets, "checksums.txt")
	if !ok {
		return "", fmt.Errorf("prepare orchestrator MCP: release %s does not include checksums.txt", release.TagName)
	}

	cacheDir, err := orchestratorReleaseCacheDir(release.TagName)
	if err != nil {
		return "", err
	}
	binaryPath := filepath.Join(cacheDir, assetName)
	checksumsPath := filepath.Join(cacheDir, "checksums.txt")

	if err := orchestratorAssetDownloader(checksumsAsset.BrowserDownloadURL, checksumsPath); err != nil {
		return "", fmt.Errorf("prepare orchestrator MCP: downloading checksums: %w", err)
	}

	if files.FileExists(binaryPath) {
		if err := verifyOrchestratorChecksum(binaryPath, checksumsPath, assetName); err == nil {
			return ensureExecutableOrchestratorBinary(binaryPath)
		}
		_ = os.Remove(binaryPath)
	}

	tempPath := binaryPath + ".download"
	_ = os.Remove(tempPath)
	if err := orchestratorAssetDownloader(binaryAsset.BrowserDownloadURL, tempPath); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("prepare orchestrator MCP: downloading binary: %w", err)
	}
	if err := verifyOrchestratorChecksum(tempPath, checksumsPath, assetName); err != nil {
		_ = os.Remove(tempPath)
		return "", err
	}
	if err := os.Rename(tempPath, binaryPath); err != nil {
		_ = os.Remove(tempPath)
		return "", fmt.Errorf("prepare orchestrator MCP: caching binary: %w", err)
	}

	return ensureExecutableOrchestratorBinary(binaryPath)
}

func orchestratorReleaseAssetName(goos, goarch string) (string, error) {
	supported := map[string]bool{
		"darwin/arm64":  true,
		"darwin/amd64":  true,
		"linux/amd64":   true,
		"linux/arm64":   true,
		"windows/amd64": true,
	}
	platform := goos + "/" + goarch
	if !supported[platform] {
		return "", fmt.Errorf("prepare orchestrator MCP: unsupported platform %s", platform)
	}

	name := fmt.Sprintf("lazyai-orchestrator-%s-%s", goos, goarch)
	if goos == "windows" {
		name += ".exe"
	}
	return name, nil
}

func orchestratorReleaseCacheDir(tagName string) (string, error) {
	cacheDir, err := orchestratorUserCacheDir()
	if err != nil || cacheDir == "" {
		cacheDir = os.TempDir()
	}
	outDir := filepath.Join(cacheDir, "lazyai", "orchestrator", "releases", safePathSegment(tagName))
	if err := files.EnsureDir(outDir); err != nil {
		return "", err
	}
	return outDir, nil
}

func safePathSegment(value string) string {
	value = strings.TrimSpace(value)
	if value == "" {
		return "latest"
	}
	replacer := strings.NewReplacer("/", "_", "\\", "_", ":", "_", "..", "_")
	return replacer.Replace(value)
}

func orchestratorReleaseURL(currentVersion string) string {
	tagName, ok := buildversion.ReleaseTag(currentVersion)
	if !ok {
		return orchestratorLatestReleaseURL
	}
	return fmt.Sprintf(orchestratorReleaseByTagURLFormat, url.PathEscape(tagName))
}

func fetchOrchestratorRelease(releaseURL string) (*orchestratorRelease, error) {
	req, err := http.NewRequest(http.MethodGet, releaseURL, nil)
	if err != nil {
		return nil, fmt.Errorf("prepare orchestrator MCP: creating release request: %w", err)
	}
	addOrchestratorGitHubHeaders(req)

	resp, err := orchestratorHTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("prepare orchestrator MCP: fetching release: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return nil, fmt.Errorf("prepare orchestrator MCP: fetching release: GitHub returned %s", resp.Status)
	}

	var release orchestratorRelease
	if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
		return nil, fmt.Errorf("prepare orchestrator MCP: parsing release: %w", err)
	}
	return &release, nil
}

func downloadOrchestratorAsset(url, destination string) error {
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("creating download request: %w", err)
	}
	addOrchestratorGitHubHeaders(req)

	resp, err := orchestratorHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("download returned %s", resp.Status)
	}

	if err := files.EnsureDir(filepath.Dir(destination)); err != nil {
		return err
	}
	out, err := os.Create(destination)
	if err != nil {
		return err
	}
	_, copyErr := io.Copy(out, resp.Body)
	closeErr := out.Close()
	if copyErr != nil {
		return copyErr
	}
	return closeErr
}

func addOrchestratorGitHubHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	if token := os.Getenv("GITHUB_TOKEN"); token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
}

func findOrchestratorReleaseAsset(assets []orchestratorReleaseAsset, name string) (orchestratorReleaseAsset, bool) {
	for _, asset := range assets {
		if asset.Name == name {
			return asset, true
		}
	}
	return orchestratorReleaseAsset{}, false
}

func verifyOrchestratorChecksum(binaryPath, checksumsPath, assetName string) error {
	expectedChecksum, err := orchestratorChecksumForAsset(checksumsPath, assetName)
	if err != nil {
		return err
	}

	actualChecksum, err := sha256ForOrchestratorFile(binaryPath)
	if err != nil {
		return fmt.Errorf("prepare orchestrator MCP: computing checksum: %w", err)
	}
	if !strings.EqualFold(actualChecksum, expectedChecksum) {
		return fmt.Errorf("prepare orchestrator MCP: checksum mismatch for %s: expected %s, got %s", assetName, expectedChecksum, actualChecksum)
	}
	return nil
}

func orchestratorChecksumForAsset(checksumsPath, assetName string) (string, error) {
	contents, err := os.ReadFile(checksumsPath)
	if err != nil {
		return "", fmt.Errorf("prepare orchestrator MCP: reading checksums: %w", err)
	}
	for _, line := range strings.Split(string(contents), "\n") {
		fields := strings.Fields(strings.TrimSpace(line))
		if len(fields) < 2 {
			continue
		}
		filename := strings.TrimPrefix(fields[len(fields)-1], "*")
		if filepath.Base(filename) == assetName {
			return fields[0], nil
		}
	}
	return "", fmt.Errorf("prepare orchestrator MCP: checksums.txt does not include %s", assetName)
}

func sha256ForOrchestratorFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

func ensureExecutableOrchestratorBinary(binaryPath string) (string, error) {
	if runtime.GOOS != "windows" {
		if err := os.Chmod(binaryPath, 0o755); err != nil {
			return "", fmt.Errorf("prepare orchestrator MCP: chmod binary: %w", err)
		}
	}
	absPath, err := filepath.Abs(binaryPath)
	if err != nil {
		return "", err
	}
	return absPath, nil
}

func resolveOrchestratorPackageDir(libraryDir string) (string, bool) {
	if libraryDir == "" {
		return "", false
	}
	repoRoot := filepath.Dir(libraryDir)
	packageDir := filepath.Join(repoRoot, "packages", "orchestrator")
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
	outDir := filepath.Join(cacheDir, "lazyai", "orchestrator")
	if err := files.EnsureDir(outDir); err != nil {
		return "", err
	}
	binaryName := "lazyai-orchestrator"
	if runtime.GOOS == "windows" {
		binaryName += ".exe"
	}
	return filepath.Join(outDir, binaryName), nil
}

func runOrchestratorBuildStep(goPath, packageDir, binaryPath string) error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	runArgs := []string{"-C", packageDir, "build", "-o", binaryPath, "./cmd/lazyai-orchestrator"}
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
