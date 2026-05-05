package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func TestScaffoldHousekeepingWritesGraphifySyncState(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()
	cfg := &types.HousekeepingConfig{
		MemoryPath:        filepath.Join(".specify", "memory"),
		EnableQmd:         true,
		QmdIndexPath:      "",
		EnableCodegraph:   true,
		CodegraphDataPath: ".codegraph/",
		EnableGraphify:    true,
		GraphifyDataPath:  "graphify-out",
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
		Qmd struct {
			Enabled   bool   `json:"enabled"`
			IndexPath string `json:"indexPath"`
		} `json:"qmd"`
		Codegraph struct {
			Enabled  bool   `json:"enabled"`
			DataPath string `json:"dataPath"`
		} `json:"codegraph"`
		Graphify struct {
			Enabled     bool   `json:"enabled"`
			DataPath    string `json:"dataPath"`
			DriftStatus string `json:"driftStatus"`
		} `json:"graphify"`
		StaleAcked struct {
			Graphify []any `json:"graphify"`
		} `json:"staleAcked"`
	}
	if err := json.Unmarshal(data, &syncState); err != nil {
		t.Fatalf("unmarshal sync-state.json: %v", err)
	}

	if !syncState.Qmd.Enabled {
		t.Fatalf("qmd.enabled = false, want true")
	}
	if syncState.Qmd.IndexPath != "" {
		t.Fatalf("qmd.indexPath = %q, want empty", syncState.Qmd.IndexPath)
	}
	if !syncState.Codegraph.Enabled {
		t.Fatalf("codegraph.enabled = false, want true")
	}
	if syncState.Codegraph.DataPath != ".codegraph/" {
		t.Fatalf("codegraph.dataPath = %q, want .codegraph/", syncState.Codegraph.DataPath)
	}
	if !syncState.Graphify.Enabled {
		t.Fatalf("graphify.enabled = false, want true")
	}
	if syncState.Graphify.DataPath != "graphify-out" {
		t.Fatalf("graphify.dataPath = %q, want graphify-out", syncState.Graphify.DataPath)
	}
	if syncState.Graphify.DriftStatus != "unknown" {
		t.Fatalf("graphify.driftStatus = %q, want unknown", syncState.Graphify.DriftStatus)
	}
	if syncState.StaleAcked.Graphify == nil {
		t.Fatalf("staleAcked.graphify missing")
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
