package library

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/rluisb/lazyai/packages/cli/internal/files"
)

// projectLibraryDir resolves the project library directory for tests.
// It walks up from the test binary or CWD to find a directory named "library"
// that also contains "mcp/catalog.json" (to avoid macOS /Library collision).
func projectLibraryDir() string {
	// Try from the executable first (works when running test binary directly)
	if exe, err := os.Executable(); err == nil {
		if dir := walkUpForLibrary(filepath.Dir(exe)); dir != "" {
			return dir
		}
	}

	// Try from CWD
	if dir, err := os.Getwd(); err == nil {
		if found := walkUpForLibrary(dir); found != "" {
			return found
		}
	}
	return ""
}

// walkUpForLibrary walks up looking for "library/mcp/catalog.json".
func walkUpForLibrary(startDir string) string {
	dir := startDir
	for i := 0; i < 20; i++ {
		candidate := filepath.Join(dir, "library")
		info, err := os.Stat(candidate)
		if err == nil && info.IsDir() {
			// Verify this is our library directory by checking for mcp/catalog.json.
			// This distinguishes our "library/" from macOS "/Library".
			if _, err := os.Stat(filepath.Join(candidate, "mcp", "catalog.json")); err == nil {
				return candidate
			}
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return ""
}

// testLibFS returns an fs.FS for reading library files in tests.
func testLibFS() fs.FS {
	dir := projectLibraryDir()
	if dir != "" {
		return os.DirFS(dir)
	}
	// Fallback: try the embedded FS (won't work in go test unless set)
	return GetLibraryFS()
}

func TestGetLibraryFS_ReturnsValidFS(t *testing.T) {
	ResetLibraryDir()
	fsys := GetLibraryFS()
	if fsys == nil {
		t.Fatal("GetLibraryFS returned nil")
	}
}

func TestGetLibraryFS_CanReadRootDir(t *testing.T) {
	fsys := testLibFS()
	entries, err := fs.ReadDir(fsys, ".")
	if err != nil {
		t.Fatalf("Failed to read root dir: %v", err)
	}
	if len(entries) == 0 {
		t.Error("Root dir has no entries")
	}
	t.Logf("Root dir entries: %d", len(entries))
}

func TestExistsFS_CatalogJSON(t *testing.T) {
	fsys := testLibFS()

	// Debug: check what's at the root of this FS
	entries, _ := fs.ReadDir(fsys, ".")
	for _, e := range entries {
		if e.IsDir() {
			subEntries, _ := fs.ReadDir(fsys, e.Name())
			t.Logf("%s/ (%d entries)", e.Name(), len(subEntries))
			for _, se := range subEntries[:min(len(subEntries), 3)] {
				t.Logf("  %s/%s", e.Name(), se.Name())
			}
			if len(subEntries) > 3 {
				t.Logf("  ... and %d more", len(subEntries)-3)
			}
		}
	}

	if !files.ExistsFS(fsys, "mcp/catalog.json") {
		t.Error("mcp/catalog.json not found in library FS")
	}
}

func TestReadFS_CatalogJSON(t *testing.T) {
	fsys := testLibFS()
	data, err := files.ReadFS(fsys, "mcp/catalog.json")
	if err != nil {
		t.Fatalf("ReadFS mcp/catalog.json: %v", err)
	}
	if len(data) == 0 {
		t.Error("mcp/catalog.json is empty")
	}
}

func TestReadFS_ConstitutionTemplate(t *testing.T) {
	fsys := testLibFS()
	data, err := files.ReadFS(fsys, "constitution/constitution.template.md")
	if err != nil {
		t.Fatalf("ReadFS constitution/constitution.template.md: %v", err)
	}
	if len(data) == 0 {
		t.Error("constitution template is empty")
	}
}

func TestReadFS_RootTemplate(t *testing.T) {
	fsys := testLibFS()
	data, err := files.ReadFS(fsys, "root/AGENTS.template.md")
	if err != nil {
		t.Fatalf("ReadFS root/AGENTS.template.md: %v", err)
	}
	if len(data) == 0 {
		t.Error("root/AGENTS.template.md is empty")
	}
}

func TestReadFS_ImplementerAndRootTemplateExposeFourPointContract(t *testing.T) {
	fsys := testLibFS()
	for _, path := range []string{
		"canonical/agents/implementer.md",
		"root/AGENTS.template.md",
	} {
		data, err := files.ReadFS(fsys, path)
		if err != nil {
			t.Fatalf("ReadFS %s: %v", path, err)
		}
		content := string(data)
		for _, want := range []string{"WHAT", "HOW", "DON'T WANT", "VALIDATE"} {
			if !strings.Contains(content, want) {
				t.Fatalf("%s missing four-point marker %q", path, want)
			}
		}
	}
}
