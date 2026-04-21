package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

// TestBuildGeminiExtension_ReadsMcpCatalogFromCwd verifies that when the
// current working directory has a .ai/mcp.json, the subcommand's catalog
// reader picks it up and passes static servers through.
func TestBuildGeminiExtension_ReadsMcpCatalogFromCwd(t *testing.T) {
	// Prepare a fake project dir with .ai/mcp.json and chdir into it.
	projDir := t.TempDir()
	aiDir := filepath.Join(projDir, ".ai")
	if err := os.MkdirAll(aiDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	catalog := `{"servers":{"memory":{"command":"npx","args":["-y","@mcp/memory"]},"secret":{"command":"c","env":{"K":"${V}"}}}}`
	if err := os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(catalog), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	orig, _ := os.Getwd()
	if err := os.Chdir(projDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	got := readGeminiExtensionMcpCatalog()
	if got == nil {
		t.Fatal("expected catalog, got nil")
	}
	if _, ok := got["memory"]; !ok {
		t.Errorf("expected 'memory' server in catalog")
	}
	if _, ok := got["secret"]; !ok {
		t.Errorf("expected 'secret' server in catalog (filter happens in geminiext.Build, not here)")
	}
}

// TestBuildGeminiExtension_AbsentCatalogReturnsNil verifies graceful
// handling when no .ai/mcp.json is present.
func TestBuildGeminiExtension_AbsentCatalogReturnsNil(t *testing.T) {
	projDir := t.TempDir()
	orig, _ := os.Getwd()
	if err := os.Chdir(projDir); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if got := readGeminiExtensionMcpCatalog(); got != nil {
		t.Errorf("expected nil when .ai/mcp.json is absent, got: %v", got)
	}
}

// TestBuildGeminiExtension_InvalidJsonReturnsNil verifies we don't propagate
// parse errors through to the generator.
func TestBuildGeminiExtension_InvalidJsonReturnsNil(t *testing.T) {
	projDir := t.TempDir()
	aiDir := filepath.Join(projDir, ".ai")
	_ = os.MkdirAll(aiDir, 0o755)
	_ = os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte("{not valid}"), 0o644)

	orig, _ := os.Getwd()
	_ = os.Chdir(projDir)
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if got := readGeminiExtensionMcpCatalog(); got != nil {
		t.Errorf("expected nil on parse error, got: %v", got)
	}
}

// TestBuildGeminiExtension_EmptyCatalogReturnsNil verifies absence-equivalent
// handling for a well-formed but empty catalog.
func TestBuildGeminiExtension_EmptyCatalogReturnsNil(t *testing.T) {
	projDir := t.TempDir()
	aiDir := filepath.Join(projDir, ".ai")
	_ = os.MkdirAll(aiDir, 0o755)
	_ = os.WriteFile(filepath.Join(aiDir, "mcp.json"), []byte(`{"servers":{}}`), 0o644)

	orig, _ := os.Getwd()
	_ = os.Chdir(projDir)
	t.Cleanup(func() { _ = os.Chdir(orig) })

	if got := readGeminiExtensionMcpCatalog(); got != nil {
		t.Errorf("expected nil for empty catalog, got: %v", got)
	}
}

// TestBuildGeminiExtension_GoldenOutput runs the subcommand against the
// embedded library and asserts the manifest contains the locked-in fields.
// Runs from the go test cwd (cmd/) so the library walkup resolves.
func TestBuildGeminiExtension_GoldenOutput(t *testing.T) {
	outDir := filepath.Join(t.TempDir(), "extension")
	if err := preflightOutDir(outDir, false); err != nil {
		t.Fatalf("preflight: %v", err)
	}

	cmd := buildGeminiExtensionCmd
	_ = cmd.Flags().Set("out", outDir)
	_ = cmd.Flags().Set("force", "false")
	if err := runBuildGeminiExtension(cmd, nil); err != nil {
		t.Fatalf("run: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(outDir, "gemini-extension.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if manifest["name"] != "ai-setup-gemini" {
		t.Errorf("name = %v", manifest["name"])
	}
	if manifest["contextFileName"] != "GEMINI.md" {
		t.Errorf("contextFileName = %v", manifest["contextFileName"])
	}
	if _, err := os.Stat(filepath.Join(outDir, "GEMINI.md")); err != nil {
		t.Errorf("GEMINI.md missing from output: %v", err)
	}
}
