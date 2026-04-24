package scaffold

import (
	"fmt"
	"io/fs"
	"path/filepath"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// ScaffoldSpecs creates the specs/ directory structure and copies the
// corresponding per-category AGENTS.md guide from `library/specs-agents/`
// into each category dir. Mirrors TS's `scaffoldSpecs`.
//
// For each `dir` in `specsDirs`, this function:
//   - creates `specs/<dir>/`
//   - if the library has `specs-agents/<dir>.md`, copies it to
//     `specs/<dir>/AGENTS.md` and records it in `fileRecords`
//   - special case: `memory` also creates a `handoffs/` subdirectory
//     (plus `decisions/`, `patterns/`, `projects/` in workspace scope)
func ScaffoldSpecs(targetDir string, setupScope types.SetupScope, libFS fs.FS, specsDirs []string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	_ = strategy
	_ = perFileOverrides
	specsDir := filepath.Join(targetDir, "specs")

	for _, dir := range specsDirs {
		dirPath := filepath.Join(specsDir, dir)
		if err := files.EnsureDir(dirPath); err != nil {
			return err
		}

		// Special case: memory also needs handoffs subdirectory.
		if dir == "memory" {
			if setupScope == types.SetupScopeWorkspace {
				_ = files.EnsureDir(filepath.Join(specsDir, "memory", "decisions"))
				_ = files.EnsureDir(filepath.Join(specsDir, "memory", "patterns"))
				_ = files.EnsureDir(filepath.Join(specsDir, "memory", "projects"))
			}
			_ = files.EnsureDir(filepath.Join(specsDir, "memory", "handoffs"))
		}

		// Copy the per-category AGENTS.md guide from library/specs-agents/<dir>.md.
		// Absent library entries are skipped silently (matches TS behavior).
		srcPath := filepath.Join("specs-agents", dir+".md")
		data, err := fs.ReadFile(libFS, srcPath)
		if err != nil {
			continue
		}

		destPath := filepath.Join(dirPath, "AGENTS.md")
		if err := files.WriteFile(destPath, data, 0o644); err != nil {
			return fmt.Errorf("write %s: %w", destPath, err)
		}

		if fileRecords != nil {
			hash, _ := files.FileHash(destPath)
			relPath, relErr := filepath.Rel(targetDir, destPath)
			if relErr != nil {
				relPath = destPath
			}
			*fileRecords = append(*fileRecords, types.TrackedFile{
				Path:   relPath,
				Hash:   hash,
				Source: "specs-agents/" + dir + ".md",
				Owner:  types.FileOwnerLibrary,
			})
		}
	}

	return nil
}
