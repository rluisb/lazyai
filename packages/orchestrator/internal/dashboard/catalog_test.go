package dashboard

import (
	"context"
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/catalog"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestCatalogAdapterListsDefinitionsWithVersionCountsAndKindFilter(t *testing.T) {
	database := newDashboardTestDB(t)
	store := catalog.NewStore(database)
	createCatalogVersion(t, store, "chain", "release", map[string]any{"description": "Release chain"}, "body v1", true)
	createCatalogVersion(t, store, "chain", "release", map[string]any{"description": "Release chain v2"}, "body v2", true)
	createCatalogVersion(t, store, "team", "launch", map[string]any{"description": "Launch team"}, "team body", true)

	adapter := NewCatalogAdapter(store)
	all, err := adapter.ListCatalog(context.Background(), "")
	if err != nil {
		t.Fatalf("list catalog: %v", err)
	}
	if len(all) != 2 || all[0].Kind != "chain" || all[0].Name != "release" || all[0].TotalVersions != 2 || all[0].ActiveVersion == nil || *all[0].ActiveVersion != 2 {
		t.Fatalf("all summaries mismatch: %+v", all)
	}

	chains, err := adapter.ListCatalog(context.Background(), "chain")
	if err != nil {
		t.Fatalf("list chain catalog: %v", err)
	}
	if len(chains) != 1 || chains[0].Kind != "chain" || chains[0].Name != "release" {
		t.Fatalf("chain summaries mismatch: %+v", chains)
	}
}

func TestCatalogAdapterDetailReturnsActiveOrRequestedVersionReadOnly(t *testing.T) {
	database := newDashboardTestDB(t)
	store := catalog.NewStore(database)
	createCatalogVersion(t, store, "chain", "release", map[string]any{"description": "Release chain"}, "body v1", true)
	createCatalogVersion(t, store, "chain", "release", map[string]any{"description": "Release chain v2", "owner": "ops"}, "body v2", true)

	adapter := NewCatalogAdapter(store)
	active, err := adapter.GetCatalogDetail(context.Background(), "chain", "release", 0)
	if err != nil {
		t.Fatalf("active detail: %v", err)
	}
	if active.Kind != "chain" || active.Name != "release" || active.Version != 2 || active.Body != "body v2" || active.Frontmatter["owner"] != "ops" {
		t.Fatalf("active detail mismatch: %+v", active)
	}
	if active.ActiveVersion == nil || *active.ActiveVersion != 2 || len(active.Versions) != 2 || active.Checksum == "" {
		t.Fatalf("active version metadata mismatch: %+v", active)
	}

	requested, err := adapter.GetCatalogDetail(context.Background(), "chain", "release", 1)
	if err != nil {
		t.Fatalf("requested detail: %v", err)
	}
	if requested.Version != 1 || requested.Body != "body v1" {
		t.Fatalf("requested version mismatch: %+v", requested)
	}
}

func TestCatalogAdapterReturnsNotFoundForMissingDefinition(t *testing.T) {
	database := newDashboardTestDB(t)
	adapter := NewCatalogAdapter(catalog.NewStore(database))

	_, err := adapter.GetCatalogDetail(context.Background(), "chain", "missing", 0)
	if !IsNotFound(err) {
		t.Fatalf("expected not-found error, got %v", err)
	}

	response := ErrorResponse("not_found", err.Error())
	if response.Error.Code != "not_found" || response.Error.Message == "" {
		t.Fatalf("not-found response mismatch: %+v", response)
	}
}

func createCatalogVersion(t *testing.T, store *catalog.Store, kind, name string, frontmatter map[string]any, body string, active bool) {
	t.Helper()
	result, err := store.CreateVersion(catalog.CreateVersionInput{
		Kind:        kind,
		Name:        name,
		Frontmatter: frontmatter,
		Body:        body,
		CreatedBy:   "dashboard-test",
		SetActive:   active,
	})
	if err != nil {
		t.Fatalf("create catalog version: %v", err)
	}
	if result.Version == 0 || result.Checksum == "" {
		t.Fatalf("unexpected create result: %+v", result)
	}
	if kind != string(types.KindChain) && kind != string(types.KindTeam) {
		t.Fatalf("test helper used unexpected kind %q", kind)
	}
}
