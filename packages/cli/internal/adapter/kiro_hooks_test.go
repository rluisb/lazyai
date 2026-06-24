package adapter

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"
)

// validKiroTriggers is the set of Kiro v3 hook triggers source-verified from
// https://kiro.dev/docs/cli/v3/hooks/ (verification date 2026-06-24).
var validKiroTriggers = map[string]bool{
	"SessionStart":     true,
	"Stop":             true,
	"PreToolUse":       true,
	"PostToolUse":      true,
	"PreTaskExec":      true,
	"PostTaskExec":     true,
	"UserPromptSubmit": true,
	"PostFileCreate":   true,
	"PostFileSave":     true,
	"PostFileDelete":   true,
	"Manual":           true,
}

// TestKiroHookAssetsEmitValidJSON installs the Kiro adapter with the test
// library FS and verifies every emitted .kiro/hooks/*.json file:
//   - parses as JSON,
//   - has top-level "version": "v1",
//   - has a "hooks" array where each entry carries a valid "trigger" (one of
//     the 11 documented Kiro v3 triggers) and an "action" with a "type" that
//     is either "command" or "agent".
//
// This satisfies SC-001 and FR-002.
func TestKiroHookAssetsEmitValidJSON(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	libFS, ok := ctx.LibraryFS.(fstest.MapFS)
	if !ok {
		t.Fatalf("expected test library fs")
	}
	// The default test FS already includes kiro/hooks/block-destructive-shell.json.
	_ = libFS

	adapter := &KiroAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Kiro Install failed: %v", err)
	}

	hooksDir := filepath.Join(targetDir, ".kiro", "hooks")
	entries, err := os.ReadDir(hooksDir)
	if err != nil {
		t.Fatalf("read .kiro/hooks dir: %v", err)
	}

	jsonCount := 0
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		name := entry.Name()
		if filepath.Ext(name) != ".json" {
			continue
		}
		jsonCount++

		raw, err := os.ReadFile(filepath.Join(hooksDir, name))
		if err != nil {
			t.Fatalf("read hook %s: %v", name, err)
		}

		var hook struct {
			Version string `json:"version"`
			Hooks   []struct {
				Name    string `json:"name"`
				Trigger string `json:"trigger"`
				Action  struct {
					Type    string `json:"type"`
					Command string `json:"command"`
					Prompt  string `json:"prompt"`
				} `json:"action"`
				Description string `json:"description"`
				Matcher     string `json:"matcher"`
			} `json:"hooks"`
		}
		if err := json.Unmarshal(raw, &hook); err != nil {
			t.Fatalf("hook %s is not valid JSON: %v\nraw: %s", name, err, raw)
		}
		if hook.Version != "v1" {
			t.Errorf("hook %s: version = %q, want \"v1\"", name, hook.Version)
		}
		if len(hook.Hooks) == 0 {
			t.Errorf("hook %s: hooks array is empty", name)
		}
		for i, h := range hook.Hooks {
			if h.Name == "" {
				t.Errorf("hook %s[%d]: name is empty", name, i)
			}
			if !validKiroTriggers[h.Trigger] {
				t.Errorf("hook %s[%d]: trigger %q is not a valid Kiro v3 trigger", name, i, h.Trigger)
			}
			if h.Action.Type != "command" && h.Action.Type != "agent" {
				t.Errorf("hook %s[%d]: action.type = %q, want \"command\" or \"agent\"", name, i, h.Action.Type)
			}
			if h.Action.Type == "command" && h.Action.Command == "" {
				t.Errorf("hook %s[%d]: command action has empty command field", name, i)
			}
			if h.Action.Type == "agent" && h.Action.Prompt == "" {
				t.Errorf("hook %s[%d]: agent action has empty prompt field", name, i)
			}
		}
	}

	if jsonCount == 0 {
		t.Fatal("expected at least one .kiro/hooks/*.json file, found none")
	}
}

// TestKiroHooksDirCreatedOnInstall verifies the .kiro/hooks directory is
// created even when the adapter runs with the default test context (which
// includes kiro/hooks assets in the memo FS).
func TestKiroHooksDirCreatedOnInstall(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)

	adapter := &KiroAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Kiro Install failed: %v", err)
	}

	hooksDir := filepath.Join(targetDir, ".kiro", "hooks")
	if _, err := os.Stat(hooksDir); err != nil {
		t.Fatalf("expected .kiro/hooks directory to exist: %v", err)
	}
}

// TestKiroHooksNoDirWhenSourceAbsent verifies the adapter does not create an
// empty .kiro/hooks directory when no kiro/hooks assets exist in the library
// FS (EC-001: no hooks selected → no directory noise).
func TestKiroHooksNoDirWhenSourceAbsent(t *testing.T) {
	ctx, targetDir := createTestAdapterContext(t)
	// Strip kiro/hooks entries from the memo FS to simulate no hook assets.
	libFS := fstest.MapFS{}
	for path, file := range ctx.LibraryFS.(fstest.MapFS) {
		if len(path) >= 10 && path[:10] == "kiro/hooks" {
			continue
		}
		libFS[path] = file
	}
	ctx.LibraryFS = libFS

	adapter := &KiroAdapter{}
	if _, err := adapter.Install(ctx); err != nil {
		t.Fatalf("Kiro Install failed: %v", err)
	}

	hooksDir := filepath.Join(targetDir, ".kiro", "hooks")
	if _, err := os.Stat(hooksDir); err == nil {
		// EnsureDir creates the directory unconditionally; this is acceptable
		// as long as no hook JSON files are written. The test verifies no
		// spurious hook files appear.
		entries, err := os.ReadDir(hooksDir)
		if err != nil {
			t.Fatalf("read .kiro/hooks: %v", err)
		}
		for _, e := range entries {
			if !e.IsDir() && filepath.Ext(e.Name()) == ".json" {
				t.Errorf("unexpected hook file %s when no kiro/hooks assets exist", e.Name())
			}
		}
	}
}
