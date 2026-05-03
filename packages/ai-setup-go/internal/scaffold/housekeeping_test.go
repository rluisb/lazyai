package scaffold

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestScaffoldHousekeepingWritesGraphifySyncState(t *testing.T) {
	t.Parallel()

	targetDir := t.TempDir()
	cfg := &types.HousekeepingConfig{
		MemoryPath:        filepath.Join("specs", "memory"),
		EnableQmd:         true,
		QmdIndexPath:      ".qmd-index",
		EnableCodegraph:   true,
		CodegraphDataPath: ".codegraph",
		EnableGraphify:    true,
		GraphifyDataPath:  "graphify-out",
	}

	if err := ScaffoldHousekeeping(targetDir, cfg); err != nil {
		t.Fatalf("ScaffoldHousekeeping: %v", err)
	}

	data, err := os.ReadFile(filepath.Join(targetDir, ".ai", "housekeeping", "sync-state.json"))
	if err != nil {
		t.Fatalf("read sync-state.json: %v", err)
	}

	var syncState struct {
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
