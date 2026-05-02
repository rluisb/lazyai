package scaffold

import (
	"os"
	"path/filepath"
	"testing"
	"testing/fstest"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

var expectedStarterStandards = []string{
	"agent-security.md",
	"context-loading.md",
	"error-handling.md",
	"orchestration-patterns.md",
	"test-patterns.md",
}

func TestScaffoldSpecs_CopiesAllStarterStandardsIntoEmptyTarget(t *testing.T) {
	targetDir := t.TempDir()
	var records []types.TrackedFile

	if err := ScaffoldSpecs(targetDir, types.SetupScopeProject, starterStandardsLibraryFS(), nil, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldSpecs: %v", err)
	}

	assertStarterStandardsExist(t, targetDir)
	for _, standard := range expectedStarterStandards {
		if !hasTrackedPath(records, filepath.ToSlash(filepath.Join("specs", "standards", "starter", standard))) {
			t.Fatalf("missing tracked record for starter standard %q in %#v", standard, records)
		}
	}
}

func TestScaffoldSpecs_MetadataFilesDoNotBlockStarterStandards(t *testing.T) {
	targetDir := t.TempDir()
	starterDir := filepath.Join(targetDir, "specs", "standards", "starter")
	if err := os.MkdirAll(starterDir, 0o755); err != nil {
		t.Fatalf("create starter dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(starterDir, ".gitkeep"), []byte("keep\n"), 0o644); err != nil {
		t.Fatalf("write .gitkeep: %v", err)
	}
	if err := os.WriteFile(filepath.Join(starterDir, "README.md"), []byte("# Local standards\n"), 0o644); err != nil {
		t.Fatalf("write README.md: %v", err)
	}

	var records []types.TrackedFile
	if err := ScaffoldSpecs(targetDir, types.SetupScopeProject, starterStandardsLibraryFS(), nil, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldSpecs: %v", err)
	}

	assertStarterStandardsExist(t, targetDir)
}

func TestScaffoldSpecs_PreservesUserAuthoredStarterStandard(t *testing.T) {
	targetDir := t.TempDir()
	starterDir := filepath.Join(targetDir, "specs", "standards", "starter")
	if err := os.MkdirAll(starterDir, 0o755); err != nil {
		t.Fatalf("create starter dir: %v", err)
	}
	const userContent = "# My local error handling standard\n\nDo not replace me.\n"
	userFile := filepath.Join(starterDir, "error-handling.md")
	if err := os.WriteFile(userFile, []byte(userContent), 0o644); err != nil {
		t.Fatalf("write existing starter standard: %v", err)
	}

	var records []types.TrackedFile
	if err := ScaffoldSpecs(targetDir, types.SetupScopeProject, starterStandardsLibraryFS(), nil, &records, types.ConflictStrategyAlign, nil); err != nil {
		t.Fatalf("ScaffoldSpecs: %v", err)
	}

	got, err := os.ReadFile(userFile)
	if err != nil {
		t.Fatalf("read existing starter standard: %v", err)
	}
	if string(got) != userContent {
		t.Fatalf("existing starter standard content = %q, want preserved %q", string(got), userContent)
	}
	assertStarterStandardsExist(t, targetDir)
	if hasTrackedPath(records, filepath.ToSlash(filepath.Join("specs", "standards", "starter", "error-handling.md"))) {
		t.Fatal("existing user-authored starter standard should not be tracked as a copied library file")
	}
}

func starterStandardsLibraryFS() fstest.MapFS {
	files := fstest.MapFS{}
	for _, standard := range expectedStarterStandards {
		files[filepath.ToSlash(filepath.Join("standards", "starter", standard))] = &fstest.MapFile{
			Data: []byte("# " + standard + "\n"),
		}
	}
	return files
}

func assertStarterStandardsExist(t *testing.T, targetDir string) {
	t.Helper()
	for _, standard := range expectedStarterStandards {
		path := filepath.Join(targetDir, "specs", "standards", "starter", standard)
		if _, err := os.Stat(path); err != nil {
			t.Fatalf("expected starter standard %q to exist: %v", standard, err)
		}
	}
}

func hasTrackedPath(records []types.TrackedFile, path string) bool {
	for _, record := range records {
		if record.Path == path {
			return true
		}
	}
	return false
}
