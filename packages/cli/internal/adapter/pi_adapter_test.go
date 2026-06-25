package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestPiAdapter_Install_EmitsAgentsSkillsAndPrompts(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdResearcher, types.AgentIdImplementer},
		Skills: []types.SkillId{types.SkillIdIssueTriage},
	}
	if testFS, ok := ctx.LibraryFS.(fstest.MapFS); ok {
		testFS["prompts/plan.md"] = &fstest.MapFile{Data: []byte("# plan")}
		testFS["prompts/research.md"] = &fstest.MapFile{Data: []byte("# research")}
	}

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	for _, rel := range []string{
		".pi/agents/researcher.md",
		".pi/agents/implementer.md",
		".pi/skills/issue-triage/SKILL.md",
		".pi/prompts/plan.md",
		".pi/prompts/research.md",
		".pi/extensions/block-destructive-shell.ts",
	} {
		assertExists(t, filepath.Join(targetDir, rel))
	}

	// Pi has no .pi/hooks path.
	assertMissing(t, filepath.Join(targetDir, ".pi", "hooks"))

	skillsDir := filepath.Join(targetDir, ".pi", "skills")
	entries, err := os.ReadDir(skillsDir)
	if err != nil {
		t.Fatalf("read skills dir: %v", err)
	}
	if len(entries) != 1 {
		t.Fatalf("expected exactly one skill emitted, found %d", len(entries))
	}
	if got := entries[0].Name(); got != string(types.SkillIdIssueTriage) {
		t.Fatalf("expected only issue-triage skill, found %q", got)
	}
}

// decodePiSettings reads and decodes the Pi settings.json at path into a map.
func decodePiSettings(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read settings.json: %v", err)
	}
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		t.Fatalf("decode settings.json: %v\n%s", err, string(data))
	}
	return m
}

// assertPiResourceArray asserts settings[key] is []any containing want values.
func assertPiResourceArray(t *testing.T, settings map[string]any, key string, wants []string) {
	t.Helper()
	raw, ok := settings[key]
	if !ok {
		t.Fatalf("settings.json missing %q key: %v", key, settings)
	}
	arr, ok := raw.([]any)
	if !ok {
		t.Fatalf("settings.json %q is not an array, got %T: %v", key, raw, raw)
	}
	if len(arr) != len(wants) {
		t.Fatalf("settings.json %q has %d entries, want %d: %v", key, len(arr), len(wants), arr)
	}
	for i, w := range wants {
		got, _ := arr[i].(string)
		if got != w {
			t.Fatalf("settings.json %q[%d] = %q, want %q", key, i, got, w)
		}
	}
}

func TestPiAdapter_Install_ProjectEmitsSettings(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.SetupScope = types.SetupScopeProject

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	settingsPath := filepath.Join(targetDir, ".pi", "settings.json")
	assertExists(t, settingsPath)
	m := decodePiSettings(t, settingsPath)

	assertPiResourceArray(t, m, "extensions", []string{"extensions"})
	assertPiResourceArray(t, m, "skills", []string{"skills"})
	assertPiResourceArray(t, m, "prompts", []string{"prompts"})
	assertPiResourceArray(t, m, "packages", []string{"."})

	// settings.json must be tracked as a TrackedFile.
	var found bool
	for _, tf := range ctx.FileRecords {
		if strings.HasSuffix(tf.Path, ".pi/settings.json") && tf.Source == "pi/settings.json" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("settings.json not tracked: %+v", ctx.FileRecords)
	}
}

func TestPiAdapter_Install_GlobalEmitsAgentSettings(t *testing.T) {
	ctx, _ := createTestAdapterContext(t)
	home := t.TempDir()
	ctx.HomeDir = home
	ctx.SetupScope = types.SetupScopeGlobal

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	// Global settings live under ~/.pi/agent/settings.json, not ~/.pi/settings.json.
	settingsPath := filepath.Join(home, ".pi", "agent", "settings.json")
	assertExists(t, settingsPath)
	assertMissing(t, filepath.Join(home, ".pi", "settings.json"))

	m := decodePiSettings(t, settingsPath)
	assertPiResourceArray(t, m, "extensions", []string{"extensions"})
	assertPiResourceArray(t, m, "packages", []string{".."})
}

func TestPiAdapter_Install_SettingsIdempotent(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.SetupScope = types.SetupScopeProject

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("first install: %v", err)
	}

	settingsPath := filepath.Join(targetDir, ".pi", "settings.json")
	first, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read first: %v", err)
	}

	// Second install should produce identical bytes (idempotent merge).
	ctx2, _ := createTestAdapterContext(t)
	ctx2.TargetDir = targetDir
	ctx2.SetupScope = types.SetupScopeProject
	if _, err := adapter.Install(ctx2); err != nil {
		t.Fatalf("second install: %v", err)
	}
	second, err := os.ReadFile(settingsPath)
	if err != nil {
		t.Fatalf("read second: %v", err)
	}
	if string(first) != string(second) {
		t.Fatalf("settings.json not idempotent:\nfirst:  %s\nsecond: %s", first, second)
	}
}

func TestPiAdapter_Install_SettingsPreservesUserKeys(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.SetupScope = types.SetupScopeProject

	settingsPath := filepath.Join(targetDir, ".pi", "settings.json")
	_ = os.MkdirAll(filepath.Dir(settingsPath), 0o755)
	existing := `{"defaultModel": "claude-sonnet-4-20250514", "theme": "light"}`
	if err := os.WriteFile(settingsPath, []byte(existing), 0o644); err != nil {
		t.Fatalf("seed: %v", err)
	}

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	m := decodePiSettings(t, settingsPath)

	// User keys preserved.
	if got, _ := m["defaultModel"].(string); got != "claude-sonnet-4-20250514" {
		t.Fatalf("defaultModel not preserved: %v", m["defaultModel"])
	}
	if got, _ := m["theme"].(string); got != "light" {
		t.Fatalf("theme not preserved: %v", m["theme"])
	}
	// LazyAI-managed keys added.
	assertPiResourceArray(t, m, "extensions", []string{"extensions"})
	assertPiResourceArray(t, m, "packages", []string{"."})
}

func TestPiAdapter_Install_EmitsSystemPromptFiles(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	// SYSTEM.md and APPEND_SYSTEM.md land at the .pi root, not in a subdir.
	systemPath := filepath.Join(targetDir, ".pi", "SYSTEM.md")
	appendPath := filepath.Join(targetDir, ".pi", "APPEND_SYSTEM.md")
	assertExists(t, systemPath)
	assertExists(t, appendPath)

	got, err := os.ReadFile(systemPath)
	if err != nil {
		t.Fatalf("read SYSTEM.md: %v", err)
	}
	if !strings.Contains(string(got), "Pi System Prompt") {
		t.Fatalf("SYSTEM.md content mismatch: %q", got)
	}

	got, err = os.ReadFile(appendPath)
	if err != nil {
		t.Fatalf("read APPEND_SYSTEM.md: %v", err)
	}
	if !strings.Contains(string(got), "Pi Appended System Prompt") {
		t.Fatalf("APPEND_SYSTEM.md content mismatch: %q", got)
	}
}

func TestPiAdapter_Install_OmitsSystemPromptsWhenSourceAbsent(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	if testFS, ok := ctx.LibraryFS.(fstest.MapFS); ok {
		delete(testFS, "pi/SYSTEM.md")
		delete(testFS, "pi/APPEND_SYSTEM.md")
	}

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	assertMissing(t, filepath.Join(targetDir, ".pi", "SYSTEM.md"))
	assertMissing(t, filepath.Join(targetDir, ".pi", "APPEND_SYSTEM.md"))
}

func TestPiAdapter_Install_ExtensionDirectoryLayout(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	// Flat extension keeps working.
	assertExists(t, filepath.Join(targetDir, ".pi", "extensions", "block-destructive-shell.ts"))

	// Directory-layout extension ships intact with index.ts, helper modules,
	// and co-located package.json (Pi auto-discovers <name>/index.ts).
	extDir := filepath.Join(targetDir, ".pi", "extensions", "extension-dir")
	assertExists(t, filepath.Join(extDir, "index.ts"))
	assertExists(t, filepath.Join(extDir, "helper.ts"))
	assertExists(t, filepath.Join(extDir, "package.json"))

	// The directory entry must be present alongside the flat file entry.
	entries, err := os.ReadDir(filepath.Join(targetDir, ".pi", "extensions"))
	if err != nil {
		t.Fatalf("read extensions dir: %v", err)
	}
	names := map[string]bool{}
	for _, e := range entries {
		names[e.Name()] = true
	}
	if !names["block-destructive-shell.ts"] {
		t.Errorf("flat extension entry missing; got %v", names)
	}
	if !names["extension-dir"] {
		t.Errorf("directory extension entry missing; got %v", names)
	}
}

func TestPiAdapter_Install_DoesNotEmitExtensionsAsSystemPrompts(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	// The pi/ source directory also contains extensions/, but the system-prompt
	// copy must only pick up SYSTEM.md / APPEND_SYSTEM.md — never .ts files.
	assertMissing(t, filepath.Join(targetDir, ".pi", "block-destructive-shell.ts"))
	assertExists(t, filepath.Join(targetDir, ".pi", "extensions", "block-destructive-shell.ts"))
}

// TestPiAdapter_Install_DoesNotEmitThemes asserts that the Pi adapter does not
// create a .pi/themes/ directory, does not write a settings.json containing
// theme or themes keys, and does not emit any theme-related assets. Pi theme
// support is intentionally out-of-scope: Pi's `theme` setting is a user-owned
// UI preference, and the `themes` resource reference has no backing library
// assets. See docs/adapters/pi.md "Theme Behavior" and issue #535.
func TestPiAdapter_Install_DoesNotEmitThemes(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	ctx.Selections = AdapterSelections{
		Agents: []types.AgentId{types.AgentIdResearcher},
		Skills: []types.SkillId{types.SkillIdIssueTriage},
	}

	adapter := &PiAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Pi Install failed: %v", err)
	}

	// No .pi/themes directory should be created.
	assertMissing(t, filepath.Join(targetDir, ".pi", "themes"))

	// No settings.json with theme/themes keys should be emitted by Install.
	// (The settings write path is owned by #532; even if it exists, it must
	// not include theme or themes keys.)
	settingsPath := filepath.Join(targetDir, ".pi", "settings.json")
	if data, err := os.ReadFile(settingsPath); err == nil {
		body := string(data)
		if strings.Contains(body, `"theme"`) || strings.Contains(body, `"themes"`) {
			t.Fatalf("Pi settings.json must not contain theme or themes keys: %s", body)
		}
	}
}
