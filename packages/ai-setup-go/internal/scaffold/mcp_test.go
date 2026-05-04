package scaffold

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestScaffoldMcp_PreparesManagedOrchestratorServerFromLocalBuild(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "packages", "orchestrator-go")
	t.Setenv("HOME", t.TempDir())
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "orchestrator": {
      "description": "orchestrator",
      "command": "ai-setup-orchestrator",
      "args": ["connect"],
      "enabled": false,
      "requiresInstall": false
    }
  }
}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "go.mod"), `module example.com/orchestrator`)

	originalLookPath := orchestratorLookPath
	originalRunner := orchestratorCommandRunner
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorCommandRunner = originalRunner
	})
	orchestratorLookPath = func(file string) (string, error) {
		switch file {
		case "go":
			return "/fake/bin/go", nil
		default:
			return "", os.ErrNotExist
		}
	}
	orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
		if name == "/fake/bin/go" {
			for i, arg := range args {
				if arg == "-o" && i+1 < len(args) {
					mustWriteTestFile(t, args[i+1], "binary")
				}
			}
		}
		return nil
	}

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	parsed := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json"))
	entry := parsed.Servers["orchestrator"]
	if !filepath.IsAbs(entry.Command) || filepath.Base(entry.Command) != "ai-setup-orchestrator" {
		t.Fatalf("command = %q, want prepared absolute ai-setup-orchestrator binary", entry.Command)
	}
	if want := []string{"connect", "--project", targetDir}; !reflect.DeepEqual(entry.Args, want) {
		t.Fatalf("args = %#v, want %#v", entry.Args, want)
	}
	if entry.RequiresInstall {
		t.Fatal("requiresInstall should be false for managed orchestrator")
	}
	if len(records) != 2 || records[0].Owner != types.FileOwnerLibrary {
		t.Fatalf("records = %#v, want two library-owned tracked files (one for .ai/mcp.json, one for .mcp.json)", records)
	}
}

func TestScaffoldMcp_DownloadsOrchestratorReleaseWhenLocalSourceAndPathMissing(t *testing.T) {
	targetDir := t.TempDir()
	libraryDir := t.TempDir()
	cacheDir := t.TempDir()
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "orchestrator": {
      "command": "ai-setup-orchestrator",
      "args": ["connect"],
      "requiresInstall": true,
      "installHint": "install me"
    }
  }
}`)

	assetName, err := orchestratorReleaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatalf("orchestratorReleaseAssetName: %v", err)
	}
	binaryContent := "downloaded orchestrator"
	var downloads []string
	stubOrchestratorReleaseDownload(t, cacheDir, assetName, binaryContent, checksumLine(assetName, binaryContent), &downloads)

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	entry := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json")).Servers["orchestrator"]
	if !filepath.IsAbs(entry.Command) {
		t.Fatalf("command = %q, want absolute cached binary path", entry.Command)
	}
	if !strings.Contains(entry.Command, filepath.Join("ai-setup", "orchestrator", "releases", "v9.9.9", assetName)) {
		t.Fatalf("command = %q, want cached release asset %q", entry.Command, assetName)
	}
	if got, err := os.ReadFile(entry.Command); err != nil || string(got) != binaryContent {
		t.Fatalf("cached binary content = %q, %v; want %q", string(got), err, binaryContent)
	}
	if want := []string{"connect", "--project", targetDir}; !reflect.DeepEqual(entry.Args, want) {
		t.Fatalf("args = %#v, want %#v", entry.Args, want)
	}
	if entry.RequiresInstall || entry.InstallHint != "" {
		t.Fatalf("managed entry should clear install fields: %#v", entry)
	}
	if want := []string{"checksums-url", "binary-url"}; !reflect.DeepEqual(downloads, want) {
		t.Fatalf("downloads = %#v, want %#v", downloads, want)
	}
}

func TestScaffoldMcp_PathOrchestratorPreventsReleaseDownload(t *testing.T) {
	targetDir := t.TempDir()
	libraryDir := t.TempDir()
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{"servers":{"orchestrator":{"command":"ai-setup-orchestrator","args":["connect"]}}}`)

	originalLookPath := orchestratorLookPath
	originalFetcher := orchestratorReleaseFetcher
	originalDownloader := orchestratorAssetDownloader
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorReleaseFetcher = originalFetcher
		orchestratorAssetDownloader = originalDownloader
	})
	orchestratorLookPath = func(file string) (string, error) {
		if file == "ai-setup-orchestrator" {
			return "/usr/local/bin/ai-setup-orchestrator", nil
		}
		return "", os.ErrNotExist
	}
	orchestratorReleaseFetcher = func(releaseURL string) (*orchestratorRelease, error) {
		t.Fatal("release fetcher should not be called when PATH binary exists")
		return nil, nil
	}
	orchestratorAssetDownloader = func(url, destination string) error {
		t.Fatal("asset downloader should not be called when PATH binary exists")
		return nil
	}

	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	entry := readScaffoldedMcpCatalog(t, filepath.Join(targetDir, ".ai", "mcp.json")).Servers["orchestrator"]
	if entry.Command != "/usr/local/bin/ai-setup-orchestrator" {
		t.Fatalf("command = %q, want PATH binary", entry.Command)
	}
}

func TestScaffoldMcp_FailsOnDownloadedOrchestratorChecksumMismatch(t *testing.T) {
	targetDir := t.TempDir()
	libraryDir := t.TempDir()
	cacheDir := t.TempDir()
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{"servers":{"orchestrator":{"command":"ai-setup-orchestrator","args":["connect"]}}}`)

	assetName, err := orchestratorReleaseAssetName(runtime.GOOS, runtime.GOARCH)
	if err != nil {
		t.Fatalf("orchestratorReleaseAssetName: %v", err)
	}
	var downloads []string
	stubOrchestratorReleaseDownload(t, cacheDir, assetName, "downloaded orchestrator", strings.Repeat("0", 64)+"  "+assetName+"\n", &downloads)

	err = ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil)
	if err == nil || !strings.Contains(err.Error(), "checksum mismatch") {
		t.Fatalf("err = %v, want checksum mismatch", err)
	}
}

func TestOrchestratorReleaseAssetName(t *testing.T) {
	tests := []struct {
		goos    string
		goarch  string
		want    string
		wantErr bool
	}{
		{goos: "darwin", goarch: "arm64", want: "ai-setup-orchestrator-darwin-arm64"},
		{goos: "darwin", goarch: "amd64", want: "ai-setup-orchestrator-darwin-amd64"},
		{goos: "linux", goarch: "amd64", want: "ai-setup-orchestrator-linux-amd64"},
		{goos: "linux", goarch: "arm64", want: "ai-setup-orchestrator-linux-arm64"},
		{goos: "windows", goarch: "amd64", want: "ai-setup-orchestrator-windows-amd64.exe"},
		{goos: "freebsd", goarch: "amd64", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.goos+"/"+tt.goarch, func(t *testing.T) {
			got, err := orchestratorReleaseAssetName(tt.goos, tt.goarch)
			if tt.wantErr {
				if err == nil {
					t.Fatal("expected unsupported platform error")
				}
				return
			}
			if err != nil || got != tt.want {
				t.Fatalf("orchestratorReleaseAssetName(%q, %q) = %q, %v; want %q", tt.goos, tt.goarch, got, err, tt.want)
			}
		})
	}
}

func TestOrchestratorReleaseURLUsesTaggedReleaseForKnownVersion(t *testing.T) {
	tests := []struct {
		name    string
		version string
		want    string
	}{
		{name: "plain semver", version: "1.2.3", want: "https://api.github.com/repos/ricardoborges-teachable/ai-setup/releases/tags/v1.2.3"},
		{name: "leading v", version: "v1.2.3", want: "https://api.github.com/repos/ricardoborges-teachable/ai-setup/releases/tags/v1.2.3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := orchestratorReleaseURL(tt.version); got != tt.want {
				t.Fatalf("orchestratorReleaseURL(%q) = %q, want %q", tt.version, got, tt.want)
			}
		})
	}
}

func TestOrchestratorReleaseURLFallsBackToLatestForDevOrEmptyVersion(t *testing.T) {
	for _, version := range []string{"", "0.0.0-dev", "v0.0.0-dev"} {
		t.Run(version, func(t *testing.T) {
			if got := orchestratorReleaseURL(version); got != orchestratorLatestReleaseURL {
				t.Fatalf("orchestratorReleaseURL(%q) = %q, want %q", version, got, orchestratorLatestReleaseURL)
			}
		})
	}
}

func TestScaffoldMcp_BuildsOrchestratorWhenDistMissing(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "packages", "orchestrator-go")
	t.Setenv("HOME", t.TempDir())
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "orchestrator": {
      "command": "ai-setup-orchestrator",
      "args": ["connect"]
    }
  }
}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "go.mod"), `module example.com/orchestrator`)

	originalLookPath := orchestratorLookPath
	originalRunner := orchestratorCommandRunner
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorCommandRunner = originalRunner
	})
	orchestratorLookPath = func(file string) (string, error) {
		switch file {
		case "go":
			return "/fake/bin/go", nil
		default:
			return "", os.ErrNotExist
		}
	}
	var calls []string
	orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
		calls = append(calls, name+" "+strings.Join(args, " "))
		if name == "/fake/bin/go" {
			for i, arg := range args {
				if arg == "-o" && i+1 < len(args) {
					mustWriteTestFile(t, args[i+1], "binary")
				}
			}
		}
		return nil
	}

	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	if len(calls) != 1 {
		t.Fatalf("runner calls = %#v, want one go build", calls)
	}
	if !strings.Contains(calls[0], "-C "+orchestratorDir+" build -o ") || !strings.Contains(calls[0], " ./cmd/orchestrator") {
		t.Fatalf("call = %q, want go build for orchestrator cmd", calls[0])
	}
}

func TestScaffoldMcp_OptionalOrchestratorSmokeTest(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "packages", "orchestrator-go")
	t.Setenv("HOME", t.TempDir())
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{"servers":{"orchestrator":{"command":"ai-setup-orchestrator","args":["connect"]}}}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "go.mod"), `module example.com/orchestrator`)
	t.Setenv("AI_SETUP_ORCHESTRATOR_SMOKE", "true")

	originalLookPath := orchestratorLookPath
	originalRunner := orchestratorCommandRunner
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorCommandRunner = originalRunner
	})
	orchestratorLookPath = func(file string) (string, error) {
		if file == "go" {
			return "/fake/bin/go", nil
		}
		return "", os.ErrNotExist
	}
	var calls []string
	orchestratorCommandRunner = func(ctx context.Context, name string, args ...string) error {
		calls = append(calls, name+" "+strings.Join(args, " "))
		if name == "/fake/bin/go" {
			for i, arg := range args {
				if arg == "-o" && i+1 < len(args) {
					mustWriteTestFile(t, args[i+1], "binary")
				}
			}
		}
		return nil
	}

	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	if got := calls[len(calls)-1]; !strings.HasSuffix(got, "ai-setup-orchestrator status") {
		t.Fatalf("last call = %q, want smoke test command", got)
	}
}

func TestScaffoldMcp_ReportsMissingGoForOrchestrator(t *testing.T) {
	targetDir := t.TempDir()
	repoRoot := t.TempDir()
	libraryDir := filepath.Join(repoRoot, "library")
	orchestratorDir := filepath.Join(repoRoot, "packages", "orchestrator-go")
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{"servers":{"orchestrator":{"command":"ai-setup-orchestrator","args":["connect"]}}}`)
	mustWriteTestFile(t, filepath.Join(orchestratorDir, "go.mod"), `module example.com/orchestrator`)

	originalLookPath := orchestratorLookPath
	t.Cleanup(func() { orchestratorLookPath = originalLookPath })
	orchestratorLookPath = func(file string) (string, error) { return "", os.ErrNotExist }

	err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"orchestrator"}, &[]types.TrackedFile{}, types.ConflictStrategyAlign, nil)
	if err == nil || !strings.Contains(err.Error(), "go executable not found") {
		t.Fatalf("err = %v, want missing go error", err)
	}
}

func TestScaffoldMcp_WritesInternalAndRootSchemas(t *testing.T) {
	targetDir := t.TempDir()
	libraryDir := t.TempDir()
	mustWriteTestFile(t, filepath.Join(libraryDir, "mcp", "catalog.json"), `{
  "servers": {
    "context7": {
      "command": "npx",
      "args": ["-y", "@upstash/context7-mcp"]
    }
  }
}`)

	var records []types.TrackedFile
	if err := ScaffoldMcp(targetDir, libraryDir, os.DirFS(libraryDir), nil, []string{"context7"}, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldMcp: %v", err)
	}

	internalData, err := os.ReadFile(filepath.Join(targetDir, ".ai", "mcp.json"))
	if err != nil {
		t.Fatalf("read internal mcp: %v", err)
	}
	var internal map[string]json.RawMessage
	if err := json.Unmarshal(internalData, &internal); err != nil {
		t.Fatalf("unmarshal internal mcp: %v", err)
	}
	if _, ok := internal["servers"]; !ok {
		t.Fatalf("internal .ai/mcp.json keys = %#v, want top-level servers", internal)
	}
	if _, ok := internal["mcpServers"]; ok {
		t.Fatalf("internal .ai/mcp.json should not contain top-level mcpServers: %#v", internal)
	}

	rootData, err := os.ReadFile(filepath.Join(targetDir, ".mcp.json"))
	if err != nil {
		t.Fatalf("read root mcp: %v", err)
	}
	var root map[string]json.RawMessage
	if err := json.Unmarshal(rootData, &root); err != nil {
		t.Fatalf("unmarshal root mcp: %v", err)
	}
	if _, ok := root["mcpServers"]; !ok {
		t.Fatalf("root .mcp.json keys = %#v, want top-level mcpServers", root)
	}
	if _, ok := root["servers"]; ok {
		t.Fatalf("root .mcp.json should not contain top-level servers: %#v", root)
	}
}

func readScaffoldedMcpCatalog(t *testing.T, path string) mcpCatalog {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %q: %v", path, err)
	}
	var parsed mcpCatalog
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("unmarshal %q: %v", path, err)
	}
	return parsed
}

func stubOrchestratorReleaseDownload(t *testing.T, cacheDir, assetName, binaryContent, checksumsContent string, downloads *[]string) {
	t.Helper()
	originalLookPath := orchestratorLookPath
	originalCacheDir := orchestratorUserCacheDir
	originalFetcher := orchestratorReleaseFetcher
	originalDownloader := orchestratorAssetDownloader
	t.Cleanup(func() {
		orchestratorLookPath = originalLookPath
		orchestratorUserCacheDir = originalCacheDir
		orchestratorReleaseFetcher = originalFetcher
		orchestratorAssetDownloader = originalDownloader
	})
	orchestratorLookPath = func(file string) (string, error) { return "", os.ErrNotExist }
	orchestratorUserCacheDir = func() (string, error) { return cacheDir, nil }
	orchestratorReleaseFetcher = func(releaseURL string) (*orchestratorRelease, error) {
		return &orchestratorRelease{
			TagName: "v9.9.9",
			Assets: []orchestratorReleaseAsset{
				{Name: assetName, BrowserDownloadURL: "binary-url"},
				{Name: "checksums.txt", BrowserDownloadURL: "checksums-url"},
			},
		}, nil
	}
	orchestratorAssetDownloader = func(url, destination string) error {
		*downloads = append(*downloads, url)
		content := binaryContent
		if url == "checksums-url" {
			content = checksumsContent
		}
		if url != "binary-url" && url != "checksums-url" {
			return fmt.Errorf("unexpected download url %q", url)
		}
		mustWriteTestFile(t, destination, content)
		return nil
	}
}

func checksumLine(assetName, content string) string {
	sum := sha256.Sum256([]byte(content))
	return fmt.Sprintf("%x  %s\n", sum, assetName)
}

func mustWriteTestFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %q: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %q: %v", path, err)
	}
}
