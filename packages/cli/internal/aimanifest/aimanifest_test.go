package aimanifest

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestLoadMissingReturnsErrNotFound(t *testing.T) {
	_, err := Load(t.TempDir())
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("want ErrNotFound, got %v", err)
	}
}

func TestSaveLoadRoundTrip(t *testing.T) {
	dir := t.TempDir()
	m := Default()
	if err := m.Save(dir); err != nil {
		t.Fatalf("save: %v", err)
	}
	raw, err := os.ReadFile(Path(dir))
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	if len(raw) == 0 || raw[len(raw)-1] != '\n' {
		t.Fatalf("manifest must end with newline")
	}
	got, err := Load(dir)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got.Version != m.Version || len(got.Targets) != len(m.Targets) {
		t.Fatalf("round-trip mismatch: %+v vs %+v", got, m)
	}
}

func TestResolveTargets(t *testing.T) {
	tests := []struct {
		name    string
		targets []string
		want    []types.ToolId
		wantErr bool
	}{
		{"claude alias", []string{"claude"}, []types.ToolId{types.ToolIdClaudeCode}, false},
		{"claude-code canonical", []string{"claude-code"}, []types.ToolId{types.ToolIdClaudeCode}, false},
		{"dedup", []string{"claude", "claude-code"}, []types.ToolId{types.ToolIdClaudeCode}, false},
		{"all seven", []string{"opencode", "claude", "copilot", "pi", "omp", "antigravity", "kiro"}, nil, false},
		{"codex rejected", []string{"codex"}, nil, true},
		{"unknown rejected", []string{"vim"}, nil, true},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			m := &Manifest{Version: SchemaVersion, Targets: tc.targets}
			got, err := m.ResolveTargets()
			if tc.wantErr {
				if err == nil {
					t.Fatalf("want error, got %v", got)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.want != nil {
				if len(got) != len(tc.want) {
					t.Fatalf("got %v want %v", got, tc.want)
				}
				for i := range got {
					if got[i] != tc.want[i] {
						t.Fatalf("got %v want %v", got, tc.want)
					}
				}
			}
		})
	}
}

func TestDefaultValidates(t *testing.T) {
	if err := Default().Validate(); err != nil {
		t.Fatalf("default manifest must validate: %v", err)
	}
}

func TestValidateRejectsEmptyTargetsAndBadProfile(t *testing.T) {
	m := &Manifest{Version: SchemaVersion, Profile: "wild", Targets: nil}
	if err := m.Validate(); err == nil {
		t.Fatal("want validation error for empty targets + bad profile")
	}
	m2 := &Manifest{Version: "", Targets: []string{"opencode"}}
	if err := m2.Validate(); err == nil {
		t.Fatal("want validation error for missing version")
	}
	m3 := &Manifest{Version: "0.9", Targets: []string{"opencode"}}
	if err := m3.Validate(); err == nil {
		t.Fatal("want validation error for unsupported version")
	}
}

func TestEnabledTargetsHonorsDisabled(t *testing.T) {
	m := &Manifest{
		Version: SchemaVersion,
		Targets: []string{"opencode", "claude"},
		Adapters: map[string]map[string]any{
			"claude": {"enabled": false},
		},
	}
	got, err := m.EnabledTargets()
	if err != nil {
		t.Fatalf("enabled: %v", err)
	}
	if len(got) != 1 || got[0] != types.ToolIdOpenCode {
		t.Fatalf("want only opencode enabled, got %v", got)
	}
}

func TestForToolsMapsTokens(t *testing.T) {
	m := ForTools([]types.ToolId{types.ToolIdClaudeCode, types.ToolIdOpenCode})
	if err := m.Validate(); err != nil {
		t.Fatalf("ForTools manifest must validate: %v", err)
	}
	resolved, err := m.ResolveTargets()
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	if len(resolved) != 2 {
		t.Fatalf("want 2 targets, got %v", resolved)
	}
}

func TestForToolsEmptyFallsBackToDefault(t *testing.T) {
	if got := ForTools(nil); len(got.Targets) != 7 {
		t.Fatalf("want 7 default targets, got %v", got.Targets)
	}
}

func TestDefaultManifestFileShape(t *testing.T) {
	dir := t.TempDir()
	if err := Default().Save(dir); err != nil {
		t.Fatalf("save: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dir, FileName)); err != nil {
		t.Fatalf("manifest file missing: %v", err)
	}
}
