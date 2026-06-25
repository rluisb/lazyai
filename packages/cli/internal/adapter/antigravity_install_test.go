package adapter

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

const antigravityHooksJSON = `{
  "lazyai-block-destructive-shell": {
    "PreToolUse": [
      {
        "matcher": "run_command",
        "hooks": [
          {
            "type": "command",
            "command": ".gemini/hooks/lazyai/block-destructive-shell.sh",
            "timeout": 10
          }
        ]
      }
    ]
  },
  "lazyai-objective-workflow-gate": {
    "Stop": [
      {
        "type": "command",
        "command": ".gemini/hooks/lazyai/objective-workflow-gate.sh"
      }
    ]
  }
}
`

func TestAntigravityAdapter_Install_ProducesAgentSkillsSurfaceAtAgentsDir(t *testing.T) {
	targetDir := t.TempDir()
	libDir := t.TempDir()

	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{}\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks", "lazyai", "block-destructive-shell.sh"), "#!/usr/bin/env bash\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks", "lazyai", "objective-workflow-gate.sh"), "#!/usr/bin/env bash\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)
	writeFile(t, filepath.Join(libDir, "skills", "diagnose.md"), "# diagnose\n")
	writeFile(t, filepath.Join(libDir, "skills", "issue-triage.md"), "# issue triage\n")

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryDir: libDir,
		LibraryFS:  nil,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Skills: []types.SkillId{types.SkillIdDiagnose, types.SkillIdIssueTriage},
		},
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Antigravity Install failed: %v", err)
	}

	assertFileExists(t, filepath.Join(targetDir, ".gemini", "settings.json"))
	assertFileExists(t, filepath.Join(targetDir, ".gemini", "hooks", "lazyai", "block-destructive-shell.sh"))
	assertFileExists(t, filepath.Join(targetDir, ".gemini", "hooks", "lazyai", "objective-workflow-gate.sh"))
	assertFileExists(t, filepath.Join(targetDir, ".agents", "hooks.json"))
	assertFileExists(t, filepath.Join(targetDir, ".agents", "skills", "diagnose", "SKILL.md"))
	assertFileExists(t, filepath.Join(targetDir, ".agents", "skills", "issue-triage", "SKILL.md"))

	content, err := os.ReadFile(filepath.Join(targetDir, ".agents", "hooks.json"))
	if err != nil {
		t.Fatalf("read hooks.json failed: %v", err)
	}
	if !strings.Contains(string(content), ".gemini/hooks/lazyai/block-destructive-shell.sh") {
		t.Fatalf("hooks.json missing block-destructive-shell hook command")
	}
	if _, err := os.Stat(filepath.Join(targetDir, ".agents", "agents")); err == nil {
		t.Fatalf("unexpected .agents/agents directory")
	}
}

func TestAntigravityAdapter_Install_UsesWorkspaceRootForAgentsAndGemini(t *testing.T) {
	workspaceRoot := t.TempDir()
	targetRepo := filepath.Join(workspaceRoot, "repo")
	if err := os.MkdirAll(targetRepo, 0o755); err != nil {
		t.Fatalf("mkdir repo failed: %v", err)
	}
	libDir := t.TempDir()

	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{}\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)
	writeFile(t, filepath.Join(libDir, "skills", "review.md"), "# review\n")

	ctx := &AdapterContext{
		TargetDir:     targetRepo,
		WorkspaceRoot: workspaceRoot,
		SetupScope:    types.SetupScopeWorkspace,
		LibraryDir:    libDir,
		LibraryFS:     nil,
		Strategy:      types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Skills: []types.SkillId{types.SkillIdReview},
		},
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Antigravity Install failed: %v", err)
	}

	assertFileExists(t, filepath.Join(workspaceRoot, ".gemini", "settings.json"))
	assertFileExists(t, filepath.Join(workspaceRoot, ".agents", "skills", "review", "SKILL.md"))
}

func TestAntigravityAdapter_CompileMCP_WritesGeminiUserConfig(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()
	writeFile(t, filepath.Join(targetDir, ".ai", "mcp.json"), `{"servers":{"filesystem":{"command":"npx","args":["-y","@modelcontextprotocol/server-filesystem","."],"enabled":true}}}`)

	adapter := &AntigravityAdapter{}
	records, err := adapter.CompileMCP(CompileContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeProject,
	})
	if err != nil {
		t.Fatalf("Antigravity CompileMCP failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected two tracked files (mcp_config.json + settings.json), got %d", len(records))
	}
	if records[0].Source != "compiled:mcp:antigravity" {
		t.Fatalf("unexpected source for first record %q", records[0].Source)
	}
	if records[1].Source != "compiled:mcp:antigravity:settings" {
		t.Fatalf("unexpected source for second record %q", records[1].Source)
	}

	configPath := filepath.Join(homeDir, ".gemini", "config", "mcp_config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected %s: %v", configPath, err)
	}
	content := string(data)
	for _, want := range []string{"mcpServers", "filesystem", "npx", "@modelcontextprotocol/server-filesystem"} {
		if !strings.Contains(content, want) {
			t.Fatalf("mcp_config.json missing %q:\n%s", want, content)
		}
	}
}

// TestAntigravityAdapter_CompileMCP_HTTPUsesServerUrl pins the Antigravity
// desktop-IDE MCP schema (serverUrl key, mcp_config.json) per
// https://antigravity.google/docs/mcp (docs/adapters/snapshots/beta-adapter-verification-2026-06.md).
// The Gemini CLI discoverable form is asserted separately in
// TestAntigravityAdapter_CompileMCP_RemoteServerGeminiCLIDiscoverable.
func TestAntigravityAdapter_CompileMCP_HTTPUsesServerUrl(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()
	writeFile(t, filepath.Join(targetDir, ".ai", "mcp.json"), `{"servers":{"remote":{"url":"https://example.com/mcp/","enabled":true}}}`)

	adapter := &AntigravityAdapter{}
	if _, err := adapter.CompileMCP(CompileContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("Antigravity CompileMCP failed: %v", err)
	}

	// Desktop-IDE path: mcp_config.json must use serverUrl.
	configPath := filepath.Join(homeDir, ".gemini", "config", "mcp_config.json")
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatalf("expected %s: %v", configPath, err)
	}
	content := string(data)
	if !strings.Contains(content, `"serverUrl"`) || !strings.Contains(content, "https://example.com/mcp/") {
		t.Fatalf("Antigravity desktop mcp_config.json must serialize as serverUrl:\n%s", content)
	}
	if strings.Contains(content, `"url"`) {
		t.Fatalf("Antigravity desktop mcp_config.json must not emit \"url\" for HTTP transport:\n%s", content)
	}
}

// TestAntigravityAdapter_CompileMCP_RemoteServerGeminiCLIDiscoverable asserts
// that remote MCP servers are written to the Gemini CLI-discoverable
// settings.json with httpUrl (Streamable HTTP) per the official schema:
// https://raw.githubusercontent.com/google-gemini/gemini-cli/main/schemas/settings.schema.json
// Fixes #546: previously only mcp_config.json+serverUrl was written, which the
// open-source Gemini CLI does not read.
func TestAntigravityAdapter_CompileMCP_RemoteServerGeminiCLIDiscoverable(t *testing.T) {
	targetDir := t.TempDir()
	homeDir := t.TempDir()
	writeFile(t, filepath.Join(targetDir, ".ai", "mcp.json"), `{"servers":{"myserver":{"url":"https://api.example.com/mcp","headers":{"Authorization":"Bearer tok"},"enabled":true}}}`)

	adapter := &AntigravityAdapter{}
	if _, err := adapter.CompileMCP(CompileContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeProject,
	}); err != nil {
		t.Fatalf("Antigravity CompileMCP failed: %v", err)
	}

	// Gemini CLI reads mcpServers from settings.json, not mcp_config.json.
	settingsPath := filepath.Join(targetDir, ".gemini", "settings.json")
	data, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("expected settings.json at %s: %v", settingsPath, err)
	}
	content := string(data)

	for _, want := range []string{"mcpServers", "myserver", "httpUrl", "https://api.example.com/mcp", "Authorization"} {
		if !strings.Contains(content, want) {
			t.Fatalf("settings.json missing Gemini CLI MCP key %q:\n%s", want, content)
		}
	}
	// Gemini CLI settings.json must NOT use serverUrl (Antigravity desktop-only key).
	if strings.Contains(content, `"serverUrl"`) {
		t.Fatalf("settings.json must not emit \"serverUrl\" (Antigravity desktop key, not Gemini CLI):\n%s", content)
	}
}

// TestAntigravityAdapter_Install_GlobalSkillsUseGeminiConfigDir pins #486 gap 1:
// global-scope skills must be written to the documented ~/.gemini/config/skills
// root, NOT ~/.agents/skills (which Antigravity does not discover globally).
func TestAntigravityAdapter_Install_GlobalSkillsUseGeminiConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	libDir := t.TempDir()

	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{}\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)
	writeFile(t, filepath.Join(libDir, "skills", "diagnose.md"), "# diagnose\n")

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeGlobal,
		LibraryDir: libDir,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Antigravity Install (global) failed: %v", err)
	}

	assertFileExists(t, filepath.Join(homeDir, ".gemini", "config", "skills", "diagnose", "SKILL.md"))
	if _, err := os.Stat(filepath.Join(homeDir, ".agents", "skills", "diagnose", "SKILL.md")); err == nil {
		t.Fatalf("global skills must not be written to ~/.agents/skills")
	}
}

// TestAntigravityAdapter_Install_GlobalHooksUseGeminiConfigDir pins #497:
// global-scope hooks.json must be written to the discoverable
// ~/.gemini/config/hooks.json root, NOT ~/.agents/hooks.json (which
// Antigravity does not discover globally), mirroring the scope-aware skills
// path pinned by TestAntigravityAdapter_Install_GlobalSkillsUseGeminiConfigDir.
func TestAntigravityAdapter_Install_GlobalHooksUseGeminiConfigDir(t *testing.T) {
	homeDir := t.TempDir()
	targetDir := t.TempDir()
	libDir := t.TempDir()

	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{}\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)
	writeFile(t, filepath.Join(libDir, "skills", "diagnose.md"), "# diagnose\n")

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		HomeDir:    homeDir,
		SetupScope: types.SetupScopeGlobal,
		LibraryDir: libDir,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}

	adapter := &AntigravityAdapter{}
	files, err := adapter.Install(ctx)
	if err != nil {
		t.Fatalf("Antigravity Install (global) failed: %v", err)
	}

	globalHooksPath := filepath.Join(homeDir, ".gemini", "config", "hooks.json")
	assertFileExists(t, globalHooksPath)

	// The legacy ~/.agents/hooks.json path must not be used at global scope.
	if _, err := os.Stat(filepath.Join(homeDir, ".agents", "hooks.json")); err == nil {
		t.Fatalf("global hooks.json must not be written to ~/.agents/hooks.json")
	}

	// Content must be the unmodified hooks asset payload.
	content, err := os.ReadFile(globalHooksPath)
	if err != nil {
		t.Fatalf("read global hooks.json failed: %v", err)
	}
	if !strings.Contains(string(content), "block-destructive-shell.sh") {
		t.Fatalf("global hooks.json missing hook command; got:\n%s", content)
	}

	// The TrackedFile record for hooks.json must reference the discoverable
	// global path (rel-slashed), not the legacy ~/.agents path.
	relExpected, err := filepath.Rel(targetDir, globalHooksPath)
	if err != nil {
		t.Fatalf("compute expected rel hooks path: %v", err)
	}
	relExpected = filepath.ToSlash(relExpected)
	found := false
	for _, tf := range files {
		if tf.Source == "antigravity/hooks.json" {
			found = true
			if tf.Path != relExpected {
				t.Fatalf("hooks.json record path = %q, want %q", tf.Path, relExpected)
			}
			if tf.Hash == "" {
				t.Fatalf("hooks.json record missing hash")
			}
			if tf.Owner != types.FileOwnerLibrary {
				t.Fatalf("hooks.json record owner = %q, want %q", tf.Owner, types.FileOwnerLibrary)
			}
		}
	}
	if !found {
		t.Fatalf("no TrackedFile record for antigravity/hooks.json")
	}
}

// TestAntigravityAdapter_Install_EmitsWorkspaceRules pins #486 gap 2: project /
// workspace installs must emit .agents/rules/lazyai.md importing the canonical
// AGENTS.md via @/AGENTS.md so Antigravity IDE discovers project instructions.
func TestAntigravityAdapter_Install_EmitsWorkspaceRules(t *testing.T) {
	targetDir := t.TempDir()
	libDir := t.TempDir()

	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{}\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)
	writeFile(t, filepath.Join(libDir, "skills", "diagnose.md"), "# diagnose\n")

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryDir: libDir,
		Strategy:   types.ConflictStrategyAlign,
		Selections: AdapterSelections{
			Skills: []types.SkillId{types.SkillIdDiagnose},
		},
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Antigravity Install failed: %v", err)
	}

	rulesPath := filepath.Join(targetDir, ".agents", "rules", "lazyai.md")
	assertFileExists(t, rulesPath)
	body, err := os.ReadFile(rulesPath)
	if err != nil {
		t.Fatalf("read rules file: %v", err)
	}
	if !strings.Contains(string(body), "@/AGENTS.md") {
		t.Fatalf(".agents/rules/lazyai.md must import @/AGENTS.md; got:\n%s", body)
	}
}

// TestAntigravityAdapter_Install_UnsupportedScopeReturnsEarlyNoWrite pins the
// error-path contract (#500): an unsupported scope (e.g. an invalid scope
// value) must return early without error and without writing any files. The
// Install guard checks IsScopeSupported before touching the filesystem, so a
// bogus scope short-circuits to a nil error with no side effects.
func TestAntigravityAdapter_Install_UnsupportedScopeReturnsEarlyNoWrite(t *testing.T) {
	targetDir := t.TempDir()
	libDir := t.TempDir()

	// Provide a valid library asset set so the only failure mode is the scope.
	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{}\n")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScope("bogus"), // invalid → IsScopeSupported false
		LibraryDir: libDir,
		Strategy:   types.ConflictStrategyAlign,
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("unsupported scope must return nil error, got: %v", err)
	}
	// No Gemini surface should have been written. The short-circuit happens
	// before any EnsureDir, so not even the .gemini directory should exist.
	assertNoGeminiArtifacts(t, targetDir)
}

// TestAntigravityAdapter_Install_MissingSettingsAssetSurfacesError pins the
// error-path contract (#500): when the library is missing the required
// antigravity/settings.json asset, Install must surface the read error rather
// than silently skipping or writing an empty settings file.
func TestAntigravityAdapter_Install_MissingSettingsAssetSurfacesError(t *testing.T) {
	targetDir := t.TempDir()
	libDir := t.TempDir()

	// Deliberately omit antigravity/settings.json. Provide hooks.json so the
	// failure is isolated to the settings asset read.
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryDir: libDir,
		Strategy:   types.ConflictStrategyAlign,
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err == nil {
		t.Fatalf("expected error for missing antigravity/settings.json, got nil")
	}
	// The .gemini directory may be created by EnsureDir before the read fails,
	// but settings.json itself must not be written.
	assertFileNotWritten(t, filepath.Join(targetDir, ".gemini", "settings.json"))
}

// TestAntigravityAdapter_Install_InvalidSettingsJSONSurfacesError pins the
// error-path contract (#500): when the library's antigravity/settings.json is
// present but not valid JSON, Install must surface the parse error rather than
// emitting a malformed settings file.
func TestAntigravityAdapter_Install_InvalidSettingsJSONSurfacesError(t *testing.T) {
	targetDir := t.TempDir()
	libDir := t.TempDir()

	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{not valid json")
	writeFile(t, filepath.Join(libDir, "antigravity", "hooks.json"), antigravityHooksJSON)

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryDir: libDir,
		Strategy:   types.ConflictStrategyAlign,
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err == nil {
		t.Fatalf("expected error for invalid antigravity/settings.json, got nil")
	}
	// The invalid JSON must not have been propagated to the output settings.json.
	assertFileNotWritten(t, filepath.Join(targetDir, ".gemini", "settings.json"))
}

// TestAntigravityAdapter_Install_MissingHooksAssetSurfacesError pins the
// error-path contract (#500): when settings.json is valid but the
// antigravity/hooks.json asset is missing, Install must surface the read error
// for hooks.json rather than silently omitting the hook event configuration.
func TestAntigravityAdapter_Install_MissingHooksAssetSurfacesError(t *testing.T) {
	targetDir := t.TempDir()
	libDir := t.TempDir()

	writeFile(t, filepath.Join(libDir, "antigravity", "settings.json"), "{}\n")
	// Deliberately omit antigravity/hooks.json.

	ctx := &AdapterContext{
		TargetDir:  targetDir,
		SetupScope: types.SetupScopeProject,
		LibraryDir: libDir,
		Strategy:   types.ConflictStrategyAlign,
	}

	adapter := &AntigravityAdapter{}
	if _, err := adapter.Install(ctx); err == nil {
		t.Fatalf("expected error for missing antigravity/hooks.json, got nil")
	}
}

// assertNoGeminiArtifacts fails the test if any file was written under the
// .gemini directory or the .agents/hooks.json hook config in the given root.
func assertNoGeminiArtifacts(t *testing.T, root string) {
	t.Helper()
	if _, err := os.Stat(filepath.Join(root, ".gemini")); err == nil {
		t.Fatalf("expected no .gemini directory under %s, but it was created", root)
	}
	if _, err := os.Stat(filepath.Join(root, ".agents", "hooks.json")); err == nil {
		t.Fatalf("expected no .agents/hooks.json under %s, but it was created", root)
	}
}

// assertFileNotWritten fails the test if the given file exists on disk. Used to
// verify error-path contracts where a failed read/parse must not leave a
// (possibly malformed) output file behind.
func assertFileNotWritten(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %s to not be written, but it exists", path)
	}
}

func assertFileExists(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path, contents string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s failed: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
		t.Fatalf("write %s failed: %v", path, err)
	}
}
