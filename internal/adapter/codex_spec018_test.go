package adapter

import (
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"testing"
	"testing/fstest"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/library"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ---------------------------------------------------------------------------
// Spec 018 Task 007 — validation argv
// ---------------------------------------------------------------------------

func TestCodexExecValidationArgs_IncludesSkipGitRepoCheck(t *testing.T) {
	args := codexExecValidationArgs()
	if !slices.Contains(args, "--skip-git-repo-check") {
		t.Errorf("argv must contain --skip-git-repo-check, got %v", args)
	}
	// Sanity: `exec` must be the first arg, prompt must be last.
	if len(args) < 3 || args[0] != "exec" {
		t.Errorf("expected exec as first arg, got %v", args)
	}
	if args[len(args)-1] == "" {
		t.Errorf("prompt arg is empty")
	}
}

// ---------------------------------------------------------------------------
// Spec 018 Task 008 — AGENTS.override write from library template
// ---------------------------------------------------------------------------

// newCodexTemplateFS returns an fs.FS that includes the library/codex
// AGENTS.override template so the adapter's reader finds it.
func newCodexTemplateFS() fs.FS {
	return fstest.MapFS{
		library.CodexAgentsOverrideTemplate: &fstest.MapFile{
			Data: []byte("# AGENTS.override — Test Template\n\n[YOUR_ORG]\n"),
		},
	}
}

func TestWriteCodexAgentsOverride_UsesLibraryTemplate(t *testing.T) {
	target := t.TempDir()
	configRoot := filepath.Join(target, ".codex")
	_ = files.EnsureDir(configRoot)

	ctx := &AdapterContext{
		TargetDir: target,
		LibraryFS: newCodexTemplateFS(),
	}
	if err := writeCodexAgentsOverride(ctx, configRoot); err != nil {
		t.Fatalf("writeCodexAgentsOverride: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configRoot, "AGENTS.override.md"))
	if err != nil {
		t.Fatalf("read override: %v", err)
	}
	if string(data) != "# AGENTS.override — Test Template\n\n[YOUR_ORG]\n" {
		t.Errorf("template content not written verbatim; got:\n%s", string(data))
	}
	if len(ctx.FileRecords) != 1 {
		t.Errorf("expected 1 tracked file record, got %d", len(ctx.FileRecords))
	}
}

func TestWriteCodexAgentsOverride_DoesNotOverwriteExisting(t *testing.T) {
	target := t.TempDir()
	configRoot := filepath.Join(target, ".codex")
	_ = files.EnsureDir(configRoot)

	// Pre-seed a user-authored override.
	overridePath := filepath.Join(configRoot, "AGENTS.override.md")
	userContent := []byte("# User's own override\n")
	if err := os.WriteFile(overridePath, userContent, 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	ctx := &AdapterContext{
		TargetDir: target,
		LibraryFS: newCodexTemplateFS(),
	}
	if err := writeCodexAgentsOverride(ctx, configRoot); err != nil {
		t.Fatalf("writeCodexAgentsOverride: %v", err)
	}

	data, err := os.ReadFile(overridePath)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if string(data) != string(userContent) {
		t.Errorf("user-authored override was overwritten\ngot:  %q\nwant: %q", string(data), string(userContent))
	}
	if len(ctx.FileRecords) != 0 {
		t.Errorf("expected 0 records when override already exists, got %d", len(ctx.FileRecords))
	}
}

func TestWriteCodexAgentsOverride_FallsBackToInlineStubWhenTemplateMissing(t *testing.T) {
	target := t.TempDir()
	configRoot := filepath.Join(target, ".codex")
	_ = files.EnsureDir(configRoot)

	ctx := &AdapterContext{
		TargetDir: target,
		LibraryFS: fstest.MapFS{}, // no template available
	}
	if err := writeCodexAgentsOverride(ctx, configRoot); err != nil {
		t.Fatalf("writeCodexAgentsOverride: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(configRoot, "AGENTS.override.md"))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	// The fallback stub must still be a valid, non-empty scaffold.
	content := string(data)
	if len(content) < 10 || content[:len("# AGENTS Override")] != "# AGENTS Override" {
		t.Errorf("fallback stub unexpectedly malformed:\n%s", content)
	}
}

// ---------------------------------------------------------------------------
// Spec 018 Task 009 — post-install MCP summary (parse helpers)
// ---------------------------------------------------------------------------

func TestParseCodexMcpListJSON_ArrayForm(t *testing.T) {
	data := []byte(`[{"name":"a"},{"name":"b"},{"name":"c"}]`)
	n, ok := parseCodexMcpListJSON(data)
	if !ok {
		t.Fatal("expected ok=true for array form")
	}
	if n != 3 {
		t.Errorf("count = %d, want 3", n)
	}
}

func TestParseCodexMcpListJSON_ObjectForm(t *testing.T) {
	data := []byte(`{"servers":[{"name":"x"},{"name":"y"}]}`)
	n, ok := parseCodexMcpListJSON(data)
	if !ok {
		t.Fatal("expected ok=true for object form")
	}
	if n != 2 {
		t.Errorf("count = %d, want 2", n)
	}
}

func TestParseCodexMcpListJSON_ObjectWithMapServers(t *testing.T) {
	data := []byte(`{"servers":{"memory":{},"github":{}}}`)
	n, ok := parseCodexMcpListJSON(data)
	if !ok {
		t.Fatal("expected ok=true for object form with map")
	}
	if n != 2 {
		t.Errorf("count = %d, want 2", n)
	}
}

func TestParseCodexMcpListJSON_InvalidReturnsFalse(t *testing.T) {
	if _, ok := parseCodexMcpListJSON([]byte("not json")); ok {
		t.Error("expected ok=false for invalid JSON")
	}
	if _, ok := parseCodexMcpListJSON([]byte("")); ok {
		t.Error("expected ok=false for empty")
	}
	if _, ok := parseCodexMcpListJSON([]byte("  \n ")); ok {
		t.Error("expected ok=false for whitespace-only")
	}
}

func TestCountCodexMcpPlaintext_CountsServerLines(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want int
	}{
		{"empty", "", 0},
		{"header only", "SERVERS:\n", 0},
		{"single entry", "SERVERS:\nmemory\n", 1},
		{"multiple", "SERVERS:\nmemory\ngithub\nfilesystem\n", 3},
		{"indented details ignored", "memory\n  command: npx\n  args: [-y]\ngithub\n", 2},
		{"blank lines ignored", "memory\n\n\ngithub\n", 2},
		{"NAME header variant", "NAME    COMMAND\nmemory  npx\n", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			got := countCodexMcpPlaintext([]byte(c.in))
			if got != c.want {
				t.Errorf("got %d, want %d\ninput:\n%s", got, c.want, c.in)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Spec 018 library constant sanity
// ---------------------------------------------------------------------------

func TestCodexAgentsOverrideTemplate_ConstantPointsToExistingFile(t *testing.T) {
	// When running under `go test` from the repo, library.GetLibraryFS
	// resolves to the on-disk library/ dir. Asserting the template is
	// readable prevents accidental deletion of the asset file.
	libFS := library.GetLibraryFS()
	if _, err := fs.Stat(libFS, library.CodexAgentsOverrideTemplate); err != nil {
		t.Fatalf("%s missing from library FS: %v", library.CodexAgentsOverrideTemplate, err)
	}
}

// ---------------------------------------------------------------------------
// Integration: adapter install emits AGENTS.override with the library template
// ---------------------------------------------------------------------------

func TestCodexAdapter_Install_WritesLibraryAgentsOverride(t *testing.T) {
	target := t.TempDir()
	home := t.TempDir()

	ctx := &AdapterContext{
		TargetDir:  target,
		HomeDir:    home,
		SetupScope: types.SetupScopeProject,
		LibraryFS:  newCodexTemplateFSWithSkills(),
	}

	adapter := &CodexAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Install: %v", err)
	}

	overridePath := filepath.Join(target, ".codex", "AGENTS.override.md")
	data, err := os.ReadFile(overridePath)
	if err != nil {
		t.Fatalf("read override: %v", err)
	}
	// Spec 018: override content should match the library template body.
	if string(data) != "# AGENTS.override — Test Template\n\n[YOUR_ORG]\n" {
		t.Errorf("override content does not match library template; got:\n%s", string(data))
	}
}

// newCodexTemplateFSWithSkills extends the basic template FS with an
// empty skills/ tree so CopyLibraryDirectory doesn't error on missing dir.
func newCodexTemplateFSWithSkills() fs.FS {
	return fstest.MapFS{
		library.CodexAgentsOverrideTemplate: &fstest.MapFile{
			Data: []byte("# AGENTS.override — Test Template\n\n[YOUR_ORG]\n"),
		},
		"skills/implement.md": &fstest.MapFile{
			Data: []byte("# Implement\n"),
		},
	}
}
