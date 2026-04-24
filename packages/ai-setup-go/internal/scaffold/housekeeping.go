package scaffold

import (
	"encoding/json"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func ScaffoldHousekeeping(targetDir string, cfg *types.HousekeepingConfig) error {
	if cfg == nil {
		return nil
	}

	memoryPath := cfg.MemoryPath
	if memoryPath == "" {
		memoryPath = filepath.Join("specs", "memory")
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
		"qmd": map[string]any{
			"enabled":     cfg.EnableQmd,
			"indexPath":   cfg.QmdIndexPath,
			"driftStatus": "unknown",
		},
		"codegraph": map[string]any{
			"enabled":     cfg.EnableCodegraph,
			"dataPath":    cfg.CodegraphDataPath,
			"driftStatus": "unknown",
		},
		"staleAcked": map[string]any{
			"qmd":       []any{},
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
