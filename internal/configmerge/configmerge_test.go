package configmerge

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/BurntSushi/toml"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
)

func TestMergeJSONFile_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	patch := map[string]any{
		"permissions": map[string]any{"allow": []any{"Read"}},
	}
	bak, err := MergeJSONFile(path, patch)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if bak != "" {
		t.Errorf("want empty backupPath for new file, got %q", bak)
	}
	got := readJSON(t, path)
	if !reflect.DeepEqual(got, patch) {
		t.Errorf("content mismatch: got %v, want %v", got, patch)
	}
}

func TestMergeJSONFile_PreservesUserKeys(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")

	// Pre-author a user-owned file with keys we should not touch.
	user := map[string]any{
		"experimental": map[string]any{"foo": true},
		"permissions":  map[string]any{"allow": []any{"UserTool"}},
	}
	writeJSON(t, path, user)

	patch := map[string]any{
		"mcpServers": map[string]any{
			"orchestrator": map[string]any{"command": "ai-setup", "args": []any{"server"}},
		},
	}
	bak, err := MergeJSONFile(path, patch)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if bak == "" {
		t.Fatal("expected backup path on first touch")
	}
	if !files.FileExists(bak) {
		t.Fatalf("backup not created at %s", bak)
	}

	got := readJSON(t, path)
	// User keys survive.
	if got["experimental"] == nil {
		t.Errorf("user experimental key lost")
	}
	if !reflect.DeepEqual(got["permissions"], user["permissions"]) {
		t.Errorf("user permissions mutated: got %v, want %v", got["permissions"], user["permissions"])
	}
	// Patch key added.
	if got["mcpServers"] == nil {
		t.Errorf("patch mcpServers not added")
	}

	// Backup holds original, not merged result.
	bakContent := readJSON(t, bak)
	if !reflect.DeepEqual(bakContent, user) {
		t.Errorf("backup mismatch: got %v, want %v", bakContent, user)
	}
}

func TestMergeJSONFile_Idempotent(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	user := map[string]any{"existing": "yes"}
	writeJSON(t, path, user)

	patch := map[string]any{"mcpServers": map[string]any{"x": map[string]any{"command": "c"}}}
	bak1, err := MergeJSONFile(path, patch)
	if err != nil {
		t.Fatalf("first merge: %v", err)
	}
	firstContent, _ := os.ReadFile(path)
	firstBak, _ := os.ReadFile(bak1)

	// Second identical run.
	bak2, err := MergeJSONFile(path, patch)
	if err != nil {
		t.Fatalf("second merge: %v", err)
	}
	secondContent, _ := os.ReadFile(path)
	secondBak, _ := os.ReadFile(bak2)

	if bak1 != bak2 {
		t.Errorf("backup path changed between runs: %q vs %q", bak1, bak2)
	}
	if string(firstContent) != string(secondContent) {
		t.Errorf("content changed on idempotent re-run")
	}
	if string(firstBak) != string(secondBak) {
		t.Errorf(".bak changed on re-run — must remain the original")
	}

	// Third run with a *different* patch: file updates, .bak unchanged.
	patch2 := map[string]any{"mcpServers": map[string]any{"y": map[string]any{"command": "d"}}}
	_, err = MergeJSONFile(path, patch2)
	if err != nil {
		t.Fatalf("third merge: %v", err)
	}
	thirdContent, _ := os.ReadFile(path)
	thirdBak, _ := os.ReadFile(bak1)
	if string(thirdContent) == string(secondContent) {
		t.Errorf("file did not update on patch change")
	}
	if string(thirdBak) != string(firstBak) {
		t.Errorf(".bak was overwritten on third run")
	}
}

func TestMergeJSONFile_SliceReplacedWholesale(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "settings.json")
	writeJSON(t, path, map[string]any{
		"permissions": map[string]any{"allow": []any{"A", "B"}},
	})
	patch := map[string]any{
		"permissions": map[string]any{"allow": []any{"C"}},
	}
	if _, err := MergeJSONFile(path, patch); err != nil {
		t.Fatalf("merge: %v", err)
	}
	got := readJSON(t, path)
	perms := got["permissions"].(map[string]any)
	allow := perms["allow"].([]any)
	if len(allow) != 1 || allow[0] != "C" {
		t.Errorf("slice not replaced wholesale: got %v, want [C]", allow)
	}
}

func TestMergeTOMLFile_NewFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	patch := map[string]any{
		"mcp_servers": map[string]any{
			"context7": map[string]any{"command": "npx", "args": []any{"-y", "@context7/server"}},
		},
	}
	bak, err := MergeTOMLFile(path, patch)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if bak != "" {
		t.Errorf("want empty backupPath for new file, got %q", bak)
	}
	got := readTOML(t, path)
	if got["mcp_servers"] == nil {
		t.Fatalf("mcp_servers missing after merge: %v", got)
	}
}

func TestMergeTOMLFile_PreservesUserTable(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.toml")
	original := "[profile]\nname = \"me\"\n\n[profile.options]\nverbose = true\n"
	if err := os.WriteFile(path, []byte(original), 0o644); err != nil {
		t.Fatal(err)
	}
	patch := map[string]any{
		"mcp_servers": map[string]any{
			"context7": map[string]any{"command": "npx"},
		},
	}
	bak, err := MergeTOMLFile(path, patch)
	if err != nil {
		t.Fatalf("merge: %v", err)
	}
	if bak == "" {
		t.Fatal("expected backup on first touch")
	}

	got := readTOML(t, path)
	if got["profile"] == nil {
		t.Errorf("user [profile] lost after merge")
	}
	if got["mcp_servers"] == nil {
		t.Errorf("patch mcp_servers missing after merge")
	}

	// Second run with same patch: idempotent.
	firstBytes, _ := os.ReadFile(path)
	if _, err := MergeTOMLFile(path, patch); err != nil {
		t.Fatalf("second merge: %v", err)
	}
	secondBytes, _ := os.ReadFile(path)
	if string(firstBytes) != string(secondBytes) {
		t.Errorf("TOML content changed on idempotent re-run:\nfirst:\n%s\nsecond:\n%s", firstBytes, secondBytes)
	}
}

// ---------- helpers ----------

func writeJSON(t *testing.T, path string, v any) {
	t.Helper()
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func readJSON(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		t.Fatalf("parse %s: %v\n%s", path, err, data)
	}
	return out
}

func readTOML(t *testing.T, path string) map[string]any {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	out := map[string]any{}
	if _, err := toml.Decode(string(data), &out); err != nil {
		t.Fatalf("parse %s: %v\n%s", path, err, data)
	}
	return out
}
