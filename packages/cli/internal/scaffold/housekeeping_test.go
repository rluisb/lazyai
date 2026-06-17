package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestScaffoldHousekeepingWritesCodegraphSyncState(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()
	cfg := &types.HousekeepingConfig{
		MemoryPath:        filepath.Join(".specify", "memory"),
		EnableCodegraph:   true,
		CodegraphDataPath: ".codegraph/",
	}

	if err := ScaffoldHousekeeping(targetDir, cfg); err != nil {
		t.Fatalf("ScaffoldHousekeeping: %v", err)
	}
	if !dirExists(filepath.Join(targetDir, ".specify", "memory")) {
		t.Fatal("expected .specify/memory directory to exist")
	}

	data, err := os.ReadFile(filepath.Join(targetDir, ".ai", "housekeeping", "sync-state.json"))
	if err != nil {
		t.Fatalf("read sync-state.json: %v", err)
	}

	var syncState struct {
		Codegraph struct {
			Enabled     bool   `json:"enabled"`
			DataPath    string `json:"dataPath"`
			DriftStatus string `json:"driftStatus"`
		} `json:"codegraph"`
		StaleAcked struct {
			Codegraph []any `json:"codegraph"`
		} `json:"staleAcked"`
	}
	if err := json.Unmarshal(data, &syncState); err != nil {
		t.Fatalf("unmarshal sync-state.json: %v", err)
	}

	if !syncState.Codegraph.Enabled {
		t.Fatalf("codegraph.enabled = false, want true")
	}
	if syncState.Codegraph.DataPath != ".codegraph/" {
		t.Fatalf("codegraph.dataPath = %q, want .codegraph/", syncState.Codegraph.DataPath)
	}
	if syncState.Codegraph.DriftStatus != "unknown" {
		t.Fatalf("codegraph.driftStatus = %q, want unknown", syncState.Codegraph.DriftStatus)
	}
	if syncState.StaleAcked.Codegraph == nil {
		t.Fatalf("staleAcked.codegraph missing")
	}
}

func TestScaffoldHousekeepingEmptyMemoryPathDefaultsToSpecifyMemory(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()
	cfg := &types.HousekeepingConfig{}

	if err := ScaffoldHousekeeping(targetDir, cfg); err != nil {
		t.Fatalf("ScaffoldHousekeeping: %v", err)
	}
	if !dirExists(filepath.Join(targetDir, ".specify", "memory")) {
		t.Fatal("expected empty memory path to create .specify/memory")
	}
}

func TestScaffoldHousekeepingLegacySpecsMemoryRemapsInSpecKitMode(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(targetDir, ".specify"), 0o755); err != nil {
		t.Fatalf("create .specify: %v", err)
	}
	cfg := &types.HousekeepingConfig{MemoryPath: filepath.Join("specs", "memory")}

	if err := ScaffoldHousekeeping(targetDir, cfg); err != nil {
		t.Fatalf("ScaffoldHousekeeping: %v", err)
	}
	if !dirExists(filepath.Join(targetDir, ".specify", "memory")) {
		t.Fatal("expected legacy specs/memory to remap to .specify/memory in speckit mode")
	}
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	return err == nil && info.IsDir()
}
