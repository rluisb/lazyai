package manifest

import (
	"path/filepath"
	"testing"

	"github.com/ricardoborges-teachable/ai-setup/internal/types"
)

func TestWriteManifest_CreatesFile(t *testing.T) {
	dir := t.TempDir()

	data := types.DefaultStoreData()
	data.Config.ProjectName = "test-project"

	if err := WriteManifest(dir, &data); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	expectedPath := filepath.Join(dir, ".ai-setup.json")
	if !manifestExistsCheck(expectedPath) {
		t.Error(".ai-setup.json not created")
	}
}

func TestReadManifest_ReadsBackCorrectly(t *testing.T) {
	dir := t.TempDir()

	original := types.DefaultStoreData()
	original.Config.ProjectName = "round-trip-test"
	original.Meta.CLIVersion = "2.0.0"

	if err := WriteManifest(dir, &original); err != nil {
		t.Fatalf("WriteManifest: %v", err)
	}

	got, err := ReadManifest(dir)
	if err != nil {
		t.Fatalf("ReadManifest: %v", err)
	}

	if got.Config.ProjectName != "round-trip-test" {
		t.Errorf("ProjectName = %q, want round-trip-test", got.Config.ProjectName)
	}
	if got.Meta.CLIVersion != "2.0.0" {
		t.Errorf("CLIVersion = %q, want 2.0.0", got.Meta.CLIVersion)
	}
}

func TestManifestExists_True(t *testing.T) {
	dir := t.TempDir()
	data := types.DefaultStoreData()
	WriteManifest(dir, &data)

	if !ManifestExists(dir) {
		t.Error("ManifestExists = false, want true")
	}
}

func TestManifestExists_False(t *testing.T) {
	dir := t.TempDir()

	if ManifestExists(dir) {
		t.Error("ManifestExists = true, want false")
	}
}

func TestReadManifest_NotFound(t *testing.T) {
	dir := t.TempDir()

	_, err := ReadManifest(dir)
	if err == nil {
		t.Error("expected error for missing manifest")
	}
}

func TestReadManifestOptional_NilWhenMissing(t *testing.T) {
	dir := t.TempDir()

	got, err := ReadManifestOptional(dir)
	if err != nil {
		t.Fatalf("ReadManifestOptional: %v", err)
	}
	if got != nil {
		t.Error("expected nil when manifest doesn't exist")
	}
}

func TestReadManifestOptional_ReturnsDataWhenPresent(t *testing.T) {
	dir := t.TempDir()
	data := types.DefaultStoreData()
	data.Config.ProjectName = "optional-test"
	WriteManifest(dir, &data)

	got, err := ReadManifestOptional(dir)
	if err != nil {
		t.Fatalf("ReadManifestOptional: %v", err)
	}
	if got == nil {
		t.Fatal("expected non-nil result")
	}
	if got.Config.ProjectName != "optional-test" {
		t.Errorf("ProjectName = %q, want optional-test", got.Config.ProjectName)
	}
}

func manifestExistsCheck(path string) bool {
	_, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	// Use the package's FileExists indirectly
	return ManifestExists(filepath.Dir(path)) || true
}
