package scaffold

import (
	"io/fs"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

var specifyTemplates = []string{
	"spec-template.md",
	"plan-template.md",
	"tasks-template.md",
	"checklist-template.md",
	"task-harness-template.md",
}

// HasSpecKitStructure returns true if a .specify/ directory already exists.
func HasSpecKitStructure(targetDir string) bool {
	return files.IsDirectory(filepath.Join(targetDir, ".specify"))
}

// DetectExistingSpecs scans specs/ for ###-slug directories and returns
// the highest number found (0 if none).
func DetectExistingSpecs(targetDir string) (hasSpecs bool, highest int) {
	specsDir := filepath.Join(targetDir, "specs")
	if !files.IsDirectory(specsDir) {
		return false, 0
	}

	entries, err := fs.ReadDir(files.OSDirFS(specsDir), ".")
	if err != nil {
		return false, 0
	}

	highest = 0
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		// Match ###- pattern (e.g., "001-feature-name")
		name := entry.Name()
		if len(name) >= 4 && name[3] == '-' {
			if num, err := strconv.Atoi(name[:3]); err == nil && num > highest {
				highest = num
			}
		}
	}

	return highest > 0, highest
}

// ScaffoldSpecs creates the speckit-compatible .specify/ and specs/ directory structure.
//
// Behavior:
// - If .specify/ already exists: skip .specify/ scaffolding entirely (respect existing)
// - If specs/###-slug/ directories exist: skip spec directory creation
// - Greenfield: create specs/.gitkeep + full .specify/ with templates, memory, scripts
func ScaffoldSpecs(targetDir string, setupScope types.SetupScope, libFS fs.FS, specsDirs []string, fileRecords *[]types.TrackedFile, strategy types.ConflictStrategy, perFileOverrides map[string]types.ConflictStrategy) error {
	_ = libFS
	_ = fileRecords
	_ = strategy
	_ = perFileOverrides

	hasSpecify := HasSpecKitStructure(targetDir)
	_, hasSpecs := DetectExistingSpecs(targetDir)

	// 1. Create .specify/ directory (speckit core) — only if not already present.
	if !hasSpecify {
		specifyDir := filepath.Join(targetDir, ".specify")

		// 1a. .specify/templates/
		templatesDir := filepath.Join(specifyDir, "templates")
		_ = files.EnsureDir(templatesDir)
		// Templates are copied by the TS runtime via the library; Go creates
		// the directory structure. The actual file copy happens in the
		// compile/adapter layer which reads library/templates/.
		for _, tpl := range specifyTemplates {
			_ = tpl // placeholder — actual copy handled by adapter layer
		}

		// 1b. .specify/memory/
		memoryDir := filepath.Join(specifyDir, "memory")
		_ = files.EnsureDir(memoryDir)

		// Workspace: create repos/ ledger directory.
		if setupScope == types.SetupScopeWorkspace {
			_ = files.EnsureDir(filepath.Join(memoryDir, "repos"))
		}

		// 1c. .specify/scripts/bash/
		_ = files.EnsureDir(filepath.Join(specifyDir, "scripts", "bash"))
	}

	// 2. Create specs/ directory.
	specsDir := filepath.Join(targetDir, "specs")
	_ = files.EnsureDir(specsDir)

	// Greenfield: create .gitkeep.
	if !hasSpecs && len(specsDirs) == 0 {
		gitkeepPath := filepath.Join(specsDir, ".gitkeep")
		if err := files.WriteFile(gitkeepPath, []byte("# Specs\n\nFeature specifications are created by the `/speckit.specify` command.\n"), 0o644); err == nil {
			hash, _ := files.FileHash(gitkeepPath)
			if fileRecords != nil {
				relPath, _ := filepath.Rel(targetDir, gitkeepPath)
				*fileRecords = append(*fileRecords, types.TrackedFile{
					Path:   filepath.ToSlash(relPath),
					Hash:   hash,
					Source: "speckit:specs-root",
					Owner:  types.FileOwnerLibrary,
				})
			}
		}
	}

	// 3. Legacy: create selected specs directories.
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
	}

	return nil
}

// OSDirFS adapter for reading directory entries from the OS filesystem.
// This avoids importing os directly in scaffold helpers.
var osDirFS = func(path string) fs.FS {
	dirFS := files.OSDirFS(path)
	return dirFS
}
