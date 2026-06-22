package cmd

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/aimanifest"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestCompileSuccessWritesToolConfigsAndTracksFiles(t *testing.T) {
	dir := t.TempDir()
	seedStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode, types.ToolIdClaudeCode}
	})
	writeCanonicalMCPConfig(t, dir)

	cmd := newCompileCommand(dir, "", false)
	if _, _ = captureOutput(t, func() {
		if err := runCompile(cmd, nil); err != nil {
			t.Fatalf("runCompile: %v", err)
		}
	}); false {
	}

	if !fileExists(filepath.Join(dir, "opencode.json")) {
		t.Fatal("expected opencode.json to be generated")
	}
	storeData := readSeededStoreData(t, dir)
	if !hasTrackedFile(storeData.Files, "opencode.json") {
		t.Fatal("expected opencode.json to be tracked")
	}
	if !hasTrackedFile(storeData.Files, ".mcp.json") {
		t.Fatal("expected .mcp.json to be tracked")
	}
}

func TestCompileWorkspaceUsesPersistedWorkspaceRootForCanonicalMCPAndOutputs(t *testing.T) {
	workspaceRoot := t.TempDir()
	planningRepo := t.TempDir()
	seedStoreData(t, planningRepo, func(data *types.StoreData) {
		data.Config.SetupScope = types.SetupScopeWorkspace
		data.Config.TargetDir = planningRepo
		data.Config.PlanningRepoPath = planningRepo
		data.Config.WorkspaceRoot = workspaceRoot
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode, types.ToolIdClaudeCode, types.ToolIdCopilot}
	})
	writeCanonicalMCPConfig(t, workspaceRoot)

	cmd := newCompileCommand(planningRepo, "", false)
	_ = cmd.Flags().Set("validate-contracts", "false")
	if _, _ = captureOutput(t, func() {
		if err := runCompile(cmd, nil); err != nil {
			t.Fatalf("runCompile: %v", err)
		}
	}); false {
	}

	for _, path := range []string{
		filepath.Join(workspaceRoot, "opencode.json"),
		filepath.Join(workspaceRoot, ".mcp.json"),
		filepath.Join(workspaceRoot, ".vscode", "mcp.json"),
	} {
		if !fileExists(path) {
			t.Fatalf("expected workspace output %q to be generated", path)
		}
	}
	for _, path := range []string{
		filepath.Join(planningRepo, "opencode.json"),
		filepath.Join(planningRepo, ".mcp.json"),
		filepath.Join(planningRepo, ".vscode", "mcp.json"),
	} {
		if fileExists(path) {
			t.Fatalf("did not expect planning repo output %q", path)
		}
	}
}

func TestCompileMissingConfigReturnsError(t *testing.T) {
	dir := t.TempDir()
	cmd := newCompileCommand(dir, "", false)
	if err := runCompile(cmd, nil); err == nil || err.Error() != "no MCP config found at .ai/mcp.json. Run 'lazyai-cli init' first" {
		t.Fatalf("runCompile error = %v, want missing-config error", err)
	}
}

func TestCompileWithUnsupportedToolFailsFastBeforeConfigValidation(t *testing.T) {
	dir := t.TempDir()
	cmd := newCompileCommand(dir, "gemini", true)

	stdout, stderr := captureOutput(t, func() {
		err := runCompile(cmd, nil)
		if err == nil || err.Error() != "unsupported tool \"gemini\" (supported tools: antigravity, claude-code, copilot, kiro, omp, opencode, pi)" {
			t.Fatalf("runCompile error = %v, want unsupported-tool error", err)
		}
	})
	combined := stdout + stderr
	if strings.Contains(combined, "contract") {
		t.Fatalf("output = %q, did not expect contract validation output", combined)
	}
	if strings.Contains(combined, "MCP config") {
		t.Fatalf("output = %q, did not expect MCP config validation output", combined)
	}
}

func TestCompileValidatesContractsBeforeMCPConfig(t *testing.T) {
	dir := t.TempDir()
	testFS := fstest.MapFS{
		"agents/alpha.md": &fstest.MapFile{Data: []byte("---\nname: reviewer\n---\n")},
		"agents/beta.md":  &fstest.MapFile{Data: []byte("---\nname: reviewer\n---\n")},
	}
	oldGetContractLibraryFS := getContractLibraryFS
	getContractLibraryFS = func() fs.FS { return testFS }
	t.Cleanup(func() { getContractLibraryFS = oldGetContractLibraryFS })

	cmd := newCompileCommand(dir, "", false)
	_, stderr := captureOutput(t, func() {
		err := runCompile(cmd, nil)
		if err == nil || err.Error() != "contract validation failed; pass --validate-contracts=false to override" {
			t.Fatalf("runCompile error = %v, want contract validation failure", err)
		}
	})

	if !strings.Contains(stderr, "contract errors: 1") || !strings.Contains(stderr, "[duplicate-name]") {
		t.Fatalf("stderr = %q, want duplicate-name contract error", stderr)
	}
	if strings.Contains(stderr, "MCP config") {
		t.Fatalf("stderr = %q, did not expect MCP config validation after contract failure", stderr)
	}
}

func TestCompileCanDisableContractValidation(t *testing.T) {
	dir := t.TempDir()
	testFS := fstest.MapFS{
		"agents/alpha.md": &fstest.MapFile{Data: []byte("---\nname: reviewer\n---\n")},
		"agents/beta.md":  &fstest.MapFile{Data: []byte("---\nname: reviewer\n---\n")},
	}
	oldGetContractLibraryFS := getContractLibraryFS
	getContractLibraryFS = func() fs.FS { return testFS }
	t.Cleanup(func() { getContractLibraryFS = oldGetContractLibraryFS })

	cmd := newCompileCommand(dir, "", false)
	_ = cmd.Flags().Set("validate-contracts", "false")
	_, stderr := captureOutput(t, func() {
		err := runCompile(cmd, nil)
		if err == nil || err.Error() != "no MCP config found at .ai/mcp.json. Run 'lazyai-cli init' first" {
			t.Fatalf("runCompile error = %v, want missing-config error", err)
		}
	})

	if strings.Contains(stderr, "contract") {
		t.Fatalf("stderr = %q, did not expect contract validation output", stderr)
	}
}

func TestCompileWithSupportedToolDryRunSucceeds(t *testing.T) {
	dir := t.TempDir()
	seedStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
	})
	writeCanonicalMCPConfig(t, dir)

	cmd := newCompileCommand(dir, "opencode", true)
	stdout, _ := captureOutput(t, func() {
		if err := runCompile(cmd, nil); err != nil {
			t.Fatalf("runCompile dry-run: %v", err)
		}
	})

	if !strings.Contains(stdout, "Would compile MCP config for OpenCode") {
		t.Fatalf("stdout = %q, want OpenCode dry-run compile output", stdout)
	}
}

func TestCompileDryRunDoesNotWriteFilesOrStoreRecords(t *testing.T) {
	dir := t.TempDir()
	seedStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode, types.ToolIdClaudeCode}
	})
	writeCanonicalMCPConfig(t, dir)

	cmd := newCompileCommand(dir, "", true)
	if _, _ = captureOutput(t, func() {
		if err := runCompile(cmd, nil); err != nil {
			t.Fatalf("runCompile dry-run: %v", err)
		}
	}); false {
	}

	if fileExists(filepath.Join(dir, "opencode.json")) {
		t.Fatal("did not expect opencode.json in dry-run")
	}
	if fileExists(filepath.Join(dir, ".opencode", "opencode.json")) {
		t.Fatal("did not expect .opencode/opencode.json in dry-run")
	}
	if fileExists(filepath.Join(dir, ".mcp.json")) {
		t.Fatal("did not expect .mcp.json in dry-run")
	}

	storeData := readSeededStoreData(t, dir)
	if len(storeData.Files) != 0 {
		t.Fatalf("tracked files = %d, want 0", len(storeData.Files))
	}
}

func TestCompileManifestDrivesTargetsAndWritesLock(t *testing.T) {
	dir := t.TempDir()
	// Store says claude only; manifest says opencode only. Manifest must win.
	seedStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdClaudeCode}
	})
	writeCanonicalMCPConfig(t, dir)
	if err := (&aimanifest.Manifest{Version: aimanifest.SchemaVersion, Targets: []string{"opencode"}}).Save(filepath.Join(dir, ".ai")); err != nil {
		t.Fatalf("save manifest: %v", err)
	}

	cmd := newCompileCommand(dir, "", false)
	_ = cmd.Flags().Set("validate-contracts", "false")
	if _, _ = captureOutput(t, func() {
		if err := runCompile(cmd, nil); err != nil {
			t.Fatalf("runCompile: %v", err)
		}
	}); false {
	}

	if !fileExists(filepath.Join(dir, "opencode.json")) {
		t.Fatal("expected opencode.json (manifest target) to be generated")
	}

	data, err := os.ReadFile(filepath.Join(dir, ".ai", "lock.json"))
	if err != nil {
		t.Fatalf("read lock: %v", err)
	}
	var lock struct {
		CompiledAt string `json:"compiledAt"`
	}
	if err := json.Unmarshal(data, &lock); err != nil {
		t.Fatalf("unmarshal lock: %v", err)
	}
	if _, err := time.Parse(time.RFC3339, lock.CompiledAt); err != nil {
		t.Fatalf("compiledAt = %q, want RFC3339 timestamp: %v", lock.CompiledAt, err)
	}
	if !fileExists(filepath.Join(dir, ".ai", "lock.json")) {
		t.Fatal("expected .ai/lock.json to be written")
	}
}

func TestCompileRejectsInvalidManifest(t *testing.T) {
	dir := t.TempDir()
	seedStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOpenCode}
	})
	writeCanonicalMCPConfig(t, dir)
	// Codex is rejected in V2.
	if err := (&aimanifest.Manifest{Version: aimanifest.SchemaVersion, Targets: []string{"codex"}}).Save(filepath.Join(dir, ".ai")); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	cmd := newCompileCommand(dir, "", false)
	_ = cmd.Flags().Set("validate-contracts", "false")
	err := runCompile(cmd, nil)
	if err == nil || !strings.Contains(err.Error(), "lazyai.json") {
		t.Fatalf("want invalid-manifest error, got %v", err)
	}
}

func TestCompileSurfacesBetaAdapters(t *testing.T) {
	dir := t.TempDir()
	seedStoreData(t, dir, func(data *types.StoreData) {
		data.Config.Tools = []types.ToolId{types.ToolIdOmp}
	})
	writeCanonicalMCPConfig(t, dir)
	if err := (&aimanifest.Manifest{Version: aimanifest.SchemaVersion, Targets: []string{"omp"}}).Save(filepath.Join(dir, ".ai")); err != nil {
		t.Fatalf("save manifest: %v", err)
	}
	cmd := newCompileCommand(dir, "", true) // dry-run
	_ = cmd.Flags().Set("validate-contracts", "false")
	stdout, _ := captureOutput(t, func() {
		if err := runCompile(cmd, nil); err != nil {
			t.Fatalf("runCompile: %v", err)
		}
	})
	if !strings.Contains(stdout, "beta") {
		t.Fatalf("expected beta label for OMP in compile output; got:\n%s", stdout)
	}
}

func hasTrackedFile(files []types.TrackedFile, path string) bool {
	for _, file := range files {
		if file.Path == path {
			return true
		}
	}
	return false
}
