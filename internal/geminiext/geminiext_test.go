package geminiext

import (
	"encoding/json"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"
)

// newTestLibraryFS returns an in-memory library mirror that exercises every
// branch the generator touches (preferred commands path, GEMINI template).
func newTestLibraryFS() fs.FS {
	return fstest.MapFS{
		"root/GEMINI.template.md": &fstest.MapFile{
			Data: []byte("# GEMINI.md\n\n[YOUR_ORG] / [YOUR_TEAM]\n"),
		},
		"gemini/commands/commit.toml": &fstest.MapFile{
			Data: []byte("description = \"Commit\"\n"),
		},
		"gemini/commands/gcs/sync.toml": &fstest.MapFile{
			Data: []byte("description = \"Sync to GCS\"\n"),
		},
	}
}

// newLegacyLibraryFS has no `gemini/commands/` but retains the old
// top-level `commands/` so the generator exercises its fallback path.
func newLegacyLibraryFS() fs.FS {
	return fstest.MapFS{
		"root/GEMINI.template.md": &fstest.MapFile{
			Data: []byte("# GEMINI.md\n"),
		},
		"commands/legacy.toml": &fstest.MapFile{
			Data: []byte("description = \"Legacy\"\n"),
		},
	}
}

func TestBuild_WritesManifestWithDefaults(t *testing.T) {
	outDir := t.TempDir()

	res, err := Build(newTestLibraryFS(), nil, outDir, "9.9.9")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if res.FileCount == 0 {
		t.Fatalf("expected files, got 0")
	}

	data, err := os.ReadFile(filepath.Join(outDir, "gemini-extension.json"))
	if err != nil {
		t.Fatalf("read manifest: %v", err)
	}
	var manifest map[string]any
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if manifest["name"] != ExtensionName {
		t.Errorf("name = %v, want %q", manifest["name"], ExtensionName)
	}
	if manifest["version"] != "9.9.9" {
		t.Errorf("version = %v", manifest["version"])
	}
	if manifest["contextFileName"] != ContextFileName {
		t.Errorf("contextFileName = %v", manifest["contextFileName"])
	}
	if _, hasMcp := manifest["mcpServers"]; hasMcp {
		t.Errorf("manifest must OMIT mcpServers when catalog is nil")
	}
}

func TestBuild_CopiesGeminiMdVerbatim(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	if _, err := Build(libFS, nil, outDir, "1.0.0"); err != nil {
		t.Fatalf("Build: %v", err)
	}
	src, _ := fs.ReadFile(libFS, "root/GEMINI.template.md")
	dst, err := os.ReadFile(filepath.Join(outDir, "GEMINI.md"))
	if err != nil {
		t.Fatalf("read dst: %v", err)
	}
	if string(src) != string(dst) {
		t.Errorf("GEMINI.md must be verbatim (raw template):\nsrc: %q\ndst: %q", src, dst)
	}
	if !strings.Contains(string(dst), "[YOUR_ORG]") {
		t.Errorf("placeholder markers were resolved; extension must ship raw template so recipients fill in their own values")
	}
}

func TestBuild_CopiesCommandsWithNamespacing(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	if _, err := Build(libFS, nil, outDir, "1.0.0"); err != nil {
		t.Fatalf("Build: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "commands", "commit.toml")); err != nil {
		t.Errorf("top-level command missing: %v", err)
	}
	// Namespaced command must preserve subdir.
	if _, err := os.Stat(filepath.Join(outDir, "commands", "gcs", "sync.toml")); err != nil {
		t.Errorf("namespaced command missing: %v", err)
	}
}

func TestBuild_FallsBackToLegacyCommandsDir(t *testing.T) {
	outDir := t.TempDir()
	libFS := newLegacyLibraryFS()

	res, err := Build(libFS, nil, outDir, "1.0.0")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if _, err := os.Stat(filepath.Join(outDir, "commands", "legacy.toml")); err != nil {
		t.Errorf("legacy command missing from fallback path: %v", err)
	}
	// Expected files: GEMINI.md + legacy.toml + manifest = 3.
	if res.FileCount != 3 {
		t.Errorf("expected 3 files (GEMINI.md + 1 command + manifest), got %d", res.FileCount)
	}
}

func TestBuild_SkipsPlaceholderMcpServers(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	catalog := map[string]McpServer{
		"static-ok": {
			Command: "npx",
			Args:    []string{"-y", "@mcp/memory"},
		},
		"needs-secret": {
			Command: "npx",
			Env:     map[string]string{"TOKEN": "${GITHUB_PAT}"},
		},
		"sse-with-placeholder-header": {
			URL:     "https://mcp.example.com",
			Headers: map[string]string{"Authorization": "Bearer ${API_KEY}"},
		},
	}

	res, err := Build(libFS, catalog, outDir, "1.0.0")
	if err != nil {
		t.Fatalf("Build: %v", err)
	}
	if len(res.SkippedMcpServers) != 2 {
		t.Errorf("expected 2 skipped servers, got %d: %v", len(res.SkippedMcpServers), res.SkippedMcpServers)
	}

	data, _ := os.ReadFile(filepath.Join(outDir, "gemini-extension.json"))
	var manifest map[string]any
	_ = json.Unmarshal(data, &manifest)
	mcpServers, _ := manifest["mcpServers"].(map[string]any)
	if len(mcpServers) != 1 {
		t.Fatalf("expected exactly 1 mcpServer in manifest, got %d", len(mcpServers))
	}
	if _, ok := mcpServers["static-ok"]; !ok {
		t.Errorf("static-ok server missing from manifest")
	}
}

func TestBuild_OmitsMcpWhenAllServersSkipped(t *testing.T) {
	outDir := t.TempDir()
	libFS := newTestLibraryFS()

	catalog := map[string]McpServer{
		"only-placeholder": {
			Command: "npx",
			Env:     map[string]string{"TOKEN": "${SECRET}"},
		},
	}

	if _, err := Build(libFS, catalog, outDir, "1.0.0"); err != nil {
		t.Fatalf("Build: %v", err)
	}
	data, _ := os.ReadFile(filepath.Join(outDir, "gemini-extension.json"))
	var manifest map[string]any
	_ = json.Unmarshal(data, &manifest)
	if _, hasMcp := manifest["mcpServers"]; hasMcp {
		t.Errorf("mcpServers key must be omitted when all entries are filtered out")
	}
}

func TestBuild_RejectsNilLibFS(t *testing.T) {
	if _, err := Build(nil, nil, t.TempDir(), "1.0.0"); err == nil {
		t.Error("expected error for nil libFS")
	}
}

func TestBuild_RejectsEmptyOutDir(t *testing.T) {
	if _, err := Build(newTestLibraryFS(), nil, "", "1.0.0"); err == nil {
		t.Error("expected error for empty outDir")
	}
}

func TestBuild_MissingGeminiTemplateIsAnError(t *testing.T) {
	libFS := fstest.MapFS{
		"gemini/commands/commit.toml": &fstest.MapFile{Data: []byte("x = 1\n")},
	}
	if _, err := Build(libFS, nil, t.TempDir(), "1.0.0"); err == nil {
		t.Error("expected error when GEMINI.template.md is absent")
	}
}
