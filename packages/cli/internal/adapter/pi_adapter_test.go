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
}
