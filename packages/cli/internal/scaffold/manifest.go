package scaffold

import (
	"path/filepath"

	"github.com/rluisb/lazyai/packages/cli/internal/aimanifest"
	"github.com/rluisb/lazyai/packages/cli/internal/files"
	"github.com/rluisb/lazyai/packages/cli/internal/types"
)

// ScaffoldManifest writes the canonical V2 manifest at <targetDir>/.ai/lazyai.json
// from the selected tools. It is idempotent: an existing manifest is preserved
// so user edits to targets/adapters survive re-runs. The written file is
// recorded in fileRecords when created.
func ScaffoldManifest(targetDir string, tools []types.ToolId, fileRecords *[]types.TrackedFile) error {
	aiDir := filepath.Join(targetDir, ".ai")
	manifestPath := aimanifest.Path(aiDir)
	if files.FileExists(manifestPath) {
		return nil
	}
	if err := files.EnsureDir(aiDir); err != nil {
		return err
	}
	if err := aimanifest.ForTools(tools).Save(aiDir); err != nil {
		return err
	}
	if fileRecords != nil {
		rel, err := filepath.Rel(targetDir, manifestPath)
		if err != nil {
			rel = manifestPath
		}
		*fileRecords = append(*fileRecords, types.TrackedFile{
			Path:  rel,
			Owner: types.FileOwnerLibrary,
			Kind:  types.FileKindFile,
		})
	}
	return nil
}
