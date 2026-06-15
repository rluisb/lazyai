package catalog

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
)

func TestCreateVersionDuplicateReturnsActualExistingVersion(t *testing.T) {
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}

	store := NewStore(database)
	first := createVersion(t, store, map[string]any{"description": "first"}, "same body")
	second := createVersion(t, store, map[string]any{"description": "second"}, "different body")
	duplicate := createVersion(t, store, map[string]any{"description": "first"}, "same body")

	if first.Version != 1 || second.Version != 2 {
		t.Fatalf("unexpected initial versions: first=%+v second=%+v", first, second)
	}
	if !duplicate.AlreadyExists {
		t.Fatalf("expected duplicate result, got %+v", duplicate)
	}
	if duplicate.Version != first.Version {
		t.Fatalf("duplicate version = %d, want %d", duplicate.Version, first.Version)
	}
}

func createVersion(t *testing.T, store *Store, frontmatter map[string]any, body string) *CreateVersionResult {
	t.Helper()
	result, err := store.CreateVersion(CreateVersionInput{
		Kind:        "chain",
		Name:        "release",
		Frontmatter: frontmatter,
		Body:        body,
		CreatedBy:   "catalog-test",
		SetActive:   true,
	})
	if err != nil {
		t.Fatalf("create version: %v", err)
	}
	return result
}
