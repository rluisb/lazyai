package sqlite

import (
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

func TestCatalogStoreCreatesListsAndReadsVersions(t *testing.T) {
	store := newCatalogStoreTestAdapter(t)

	var _ ports.CatalogStore = store

	created, err := store.CreateVersion(domain.CreateCatalogVersionInput{
		Kind:        "chain",
		Name:        "release",
		Frontmatter: map[string]any{"description": "Release chain"},
		Body:        "body v1",
		CreatedBy:   "catalog-store-test",
		SetActive:   true,
	})
	if err != nil {
		t.Fatalf("create catalog version: %v", err)
	}
	if created.Version != 1 || created.Checksum == "" || created.AlreadyExists {
		t.Fatalf("unexpected create result: %+v", created)
	}

	duplicate, err := store.CreateVersion(domain.CreateCatalogVersionInput{
		Kind:        "chain",
		Name:        "release",
		Frontmatter: map[string]any{"description": "Release chain"},
		Body:        "body v1",
		SetActive:   true,
	})
	if err != nil {
		t.Fatalf("create duplicate catalog version: %v", err)
	}
	if !duplicate.AlreadyExists || duplicate.Version != created.Version {
		t.Fatalf("unexpected duplicate result: %+v", duplicate)
	}

	items, err := store.List("")
	if err != nil {
		t.Fatalf("list catalog: %v", err)
	}
	if len(items) != 1 || items[0].Kind != "chain" || items[0].Name != "release" || items[0].ActiveVersion == nil || *items[0].ActiveVersion != 1 {
		t.Fatalf("unexpected catalog summaries: %+v", items)
	}

	version, err := store.GetVersion("chain", "release", 0)
	if err != nil {
		t.Fatalf("get active version: %v", err)
	}
	if version.Body != "body v1" || version.CreatedBy != "catalog-store-test" {
		t.Fatalf("unexpected version: %+v", version)
	}

	versions, err := store.ListVersions("chain", "release")
	if err != nil {
		t.Fatalf("list versions: %v", err)
	}
	if len(versions) != 1 || versions[0].Version != version.Version {
		t.Fatalf("unexpected versions: %+v", versions)
	}
}

func newCatalogStoreTestAdapter(t *testing.T) *CatalogStore {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewCatalogStore(database)
}
