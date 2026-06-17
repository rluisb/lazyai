package scaffold

import (
	"encoding/json"
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

func ScaffoldHousekeeping(targetDir string, cfg *types.HousekeepingConfig) error {
	if cfg == nil {
		return nil
	}

	memoryPath := cfg.MemoryPath
	// In speckit mode, remap legacy specs/memory and empty defaults to .specify/memory.
	if HasSpecKitStructure(targetDir) && (memoryPath == "" || memoryPath == filepath.Join("specs", "memory") || memoryPath == filepath.Join(".specify", "memory")) {
		memoryPath = filepath.Join(".specify", "memory")
	} else if memoryPath == "" {
		memoryPath = filepath.Join(".specify", "memory")
	}
	if err := files.EnsureDir(filepath.Join(targetDir, memoryPath)); err != nil {
		return err
	}

	housekeepingDir := filepath.Join(targetDir, ".ai", "housekeeping")
	if err := files.EnsureDir(housekeepingDir); err != nil {
		return err
	}

	initialState := map[string]any{
		"schemaVersion": 1,
		"updatedAt":     "",
		"codegraph": map[string]any{
			"enabled":     cfg.EnableCodegraph,
			"dataPath":    cfg.CodegraphDataPath,
			"driftStatus": "unknown",
		},
		"staleAcked": map[string]any{
			"codegraph": []any{},
		},
		"repairProposals": []any{},
	}

	data, err := json.MarshalIndent(initialState, "", "  ")
	if err != nil {
		return err
	}

	return files.WriteFile(filepath.Join(housekeepingDir, "sync-state.json"), data, 0o644)
}
