package cmd

import (
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
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

	if !fileExists(filepath.Join(dir, ".opencode", "opencode.jsonc")) {
		t.Fatal("expected opencode.jsonc to be generated")
	}
	if !fileExists(filepath.Join(dir, ".mcp.json")) {
		t.Fatal("expected .mcp.json to be generated")
	}

	storeData := readSeededStoreData(t, dir)
	if !hasTrackedFile(storeData.Files, ".opencode/opencode.jsonc") {
		t.Fatal("expected opencode.jsonc to be tracked")
	}
	if !hasTrackedFile(storeData.Files, ".mcp.json") {
		t.Fatal("expected .mcp.json to be tracked")
	}
}

func TestCompileMissingConfigReturnsError(t *testing.T) {
	dir := t.TempDir()
	cmd := newCompileCommand(dir, "", false)
	if err := runCompile(cmd, nil); err == nil || err.Error() != "no MCP config found at .ai/mcp.json. Run 'ai-setup init' first" {
		t.Fatalf("runCompile error = %v, want missing-config error", err)
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

	if fileExists(filepath.Join(dir, ".opencode", "opencode.jsonc")) {
		t.Fatal("did not expect opencode.jsonc in dry-run")
	}
	if fileExists(filepath.Join(dir, ".mcp.json")) {
		t.Fatal("did not expect .mcp.json in dry-run")
	}

	storeData := readSeededStoreData(t, dir)
	if len(storeData.Files) != 0 {
		t.Fatalf("tracked files = %d, want 0", len(storeData.Files))
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
