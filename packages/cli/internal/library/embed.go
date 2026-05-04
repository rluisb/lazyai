// Package library provides access to the embedded library data used by
// scaffold, adapter, compiler, and generator packages.
//
// In development (running from repo root), the filesystem library/ directory
// is used directly. In production (installed binary), the embedded FS is used.
package library

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
)

// embeddedFS is set from main.go via SetEmbeddedFS() using the project-root
// embed directive. This avoids path traversal issues with go:embed.
var embeddedFS fs.FS

// mu protects the libraryDir cache.
var mu sync.Once

// cachedDir is the resolved library directory (filesystem mode).
var cachedDir string

// cachedErr is any error from resolving the library directory.
var cachedErr error

// SetEmbeddedFS sets the embedded filesystem. Called from main.go during init.
func SetEmbeddedFS(fsys fs.FS) {
	embeddedFS = fsys
}

// GetLibraryFS returns an fs.FS for the library data.
//
// In development mode (running from the repo), it returns the filesystem
// directly via os.DirFS for live editing. In production mode (installed
// binary or no library/ dir found), it returns a sub-FS of the embedded FS
// with the "library/" prefix stripped so that paths like "constitution/..."
// work without the library/ prefix.
func GetLibraryFS() fs.FS {
	dir, err := FindLibraryDir()
	if err == nil && dir != "" {
		return os.DirFS(dir)
	}
	// embeddedFS already has the "library/" prefix stripped by main.go
	// (which calls fs.Sub(libraryEmbed, "library") before SetEmbeddedFS).
	// Return it directly — do NOT fs.Sub again.
	if embeddedFS != nil {
		return embeddedFS
	}
	// Fallback: return current directory (won't work in production).
	return os.DirFS(".")
}

// FindLibraryDir resolves the library directory by walking up from the
// current working directory and then from the executable location.
// Returns empty string and an error if not found (caller should fall back to embedded FS).
func FindLibraryDir() (string, error) {
	mu.Do(func() {
		cachedDir, cachedErr = findLibraryDirInternal()
	})
	return cachedDir, cachedErr
}

// ResetLibraryDir clears the cached library directory (useful for testing).
func ResetLibraryDir() {
	mu = sync.Once{}
	cachedDir = ""
	cachedErr = nil
}

func findLibraryDirInternal() (string, error) {
	// Strategy 1: Walk up from current working directory.
	// We look for a directory named "library" that contains "mcp/catalog.json"
	// to distinguish our library directory from macOS /Library.
	if dir, err := walkUpFindLibrary(); err == nil {
		return dir, nil
	}

	// Strategy 2: Walk up from the executable location.
	if exe, err := os.Executable(); err == nil {
		exeDir := filepath.Dir(exe)
		if dir, err := walkUpFromLibrary(exeDir); err == nil {
			return dir, nil
		}
	}

	return "", fs.ErrNotExist
}

// walkUpFindLibrary walks up from the current working directory looking for
// a directory named "library" that contains "mcp/catalog.json".
func walkUpFindLibrary() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	return walkUpFromLibrary(dir)
}

// walkUpFromLibrary walks up from startDir looking for a directory named "library"
// that contains "mcp/catalog.json" (to distinguish from system /Library on macOS).
func walkUpFromLibrary(startDir string) (string, error) {
	dir := startDir
	for i := 0; i < 20; i++ {
		candidate := filepath.Join(dir, "library")
		if info, err := os.Stat(candidate); err == nil && info.IsDir() {
			// Verify this is our library directory by checking for mcp/catalog.json
			if files.FileExists(filepath.Join(candidate, "mcp", "catalog.json")) {
				return candidate, nil
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", fs.ErrNotExist
}

// Root returns the root path within the library FS.
// When using filesystem mode, this is the absolute path to library/.
// When using embedded FS, this is "library".
func Root() string {
	dir, err := FindLibraryDir()
	if err == nil && dir != "" {
		return dir
	}
	return "library"
}

// AgentsDir returns the path to the agents directory.
func AgentsDir() string { return filepath.Join(Root(), "agents") }

// SkillsDir returns the path to the skills directory.
func SkillsDir() string { return filepath.Join(Root(), "skills") }

// TemplatesDir returns the path to the templates directory.
func TemplatesDir() string { return filepath.Join(Root(), "templates") }

// RulesDir returns the path to the rules directory.
func RulesDir() string { return filepath.Join(Root(), "rules") }

// PromptsDir returns the path to the prompts directory.
func PromptsDir() string { return filepath.Join(Root(), "prompts") }

// ConstitutionDir returns the path to the constitution directory.
func ConstitutionDir() string { return filepath.Join(Root(), "constitution") }

// FragmentsDir returns the path to the fragments directory.
func FragmentsDir() string { return filepath.Join(Root(), "fragments") }

// InfraDir returns the path to the infra directory.
func InfraDir() string { return filepath.Join(Root(), "infra") }

// MCPDir returns the path to the MCP directory.
func MCPDir() string { return filepath.Join(Root(), "mcp") }

// OrchestrationDir returns the path to the orchestration directory.
func OrchestrationDir() string { return filepath.Join(Root(), "orchestration") }

// RootDir returns the path to the root templates directory.
func RootDir() string { return filepath.Join(Root(), "root") }

// SpecsAgentsDir returns the path to the specs-agents directory.
func SpecsAgentsDir() string { return filepath.Join(Root(), "specs-agents") }

// ToolAgentsDir returns the path to the tool-agents directory.
func ToolAgentsDir() string { return filepath.Join(Root(), "tool-agents") }

// ToolTemplatesDir returns the path to the tool-templates directory.
func ToolTemplatesDir() string { return filepath.Join(Root(), "tool-templates") }

// CopilotAgentsDir returns the path to the copilot agents directory.
func CopilotAgentsDir() string { return filepath.Join(Root(), "copilot", "agents") }

// CopilotInstructionsDir returns the path to the copilot instructions directory.
func CopilotInstructionsDir() string { return filepath.Join(Root(), "copilot", "instructions") }
