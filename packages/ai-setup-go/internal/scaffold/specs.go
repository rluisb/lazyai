package scaffold

import (
	"io/fs"
	"os"
	"path/filepath"
	"strconv"

	"github.com/ricardoborges-teachable/ai-setup/internal/files"
	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

// specifyTemplates names the speckit-aligned templates copied into
// .specify/templates/ during scaffold. Spec 022 / E2.1 added the actual
// copy implementation (it was previously a placeholder). When a template
// is missing from the embedded library FS the file is silently skipped —
// scaffold must work in test environments that ship a minimal library.
var specifyTemplates = []string{
	"spec-template.md",
	"plan-template.md",
	"tasks-template.md",
	"checklist-template.md",
	"task-harness-template.md",
	"spike-template.md",
	"poc-template.md",
	"housekeeping-template.md",
	"audit-template.md",
	"ledger-template.md",
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

	entries, err := fs.ReadDir(os.DirFS(specsDir), ".")
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
	hasSpecs, _ := DetectExistingSpecs(targetDir)

	// 1. Create .specify/ directory (speckit core) — only if not already present.
	if !hasSpecify {
		specifyDir := filepath.Join(targetDir, ".specify")

		// 1a. .specify/templates/
		templatesDir := filepath.Join(specifyDir, "templates")
		_ = files.EnsureDir(templatesDir)
		// Spec 022 / E2.1: actually copy the speckit-aligned templates from
		// the embedded library FS into .specify/templates/. The earlier
		// placeholder loop (which did nothing) left .specify/templates/
		// empty, breaking /speckit.specify and /speckit.plan as soon as
		// they tried to read the template. We copy what's present and skip
		// what isn't, so minimal-library tests still pass.
		if libFS != nil {
			for _, tpl := range specifyTemplates {
				src := "templates/" + tpl
				data, err := fs.ReadFile(libFS, src)
				if err != nil {
					continue // template not in this build's library — skip
				}
				dest := filepath.Join(templatesDir, tpl)
				if err := files.WriteFile(dest, data, 0o644); err != nil {
					return err
				}
				if fileRecords != nil {
					hash, _ := files.FileHash(dest)
					relPath, _ := filepath.Rel(targetDir, dest)
					*fileRecords = append(*fileRecords, types.TrackedFile{
						Path:   filepath.ToSlash(relPath),
						Hash:   hash,
						Source: "speckit:" + src,
						Owner:  types.FileOwnerLibrary,
					})
				}
			}
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

// osDirFS wraps os.DirFS so scaffold helpers can swap in a fake FS in tests
// without touching the real filesystem.
var osDirFS = func(path string) fs.FS {
	return os.DirFS(path)
}
