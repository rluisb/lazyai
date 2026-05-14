package dashboard

import (
	"context"
	"strings"
	"testing"

	sqliteadapter "github.com/rluisb/lazyai/packages/orchestrator/adapters/sqlite"
	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

func TestCatalogAdapterListsDefinitionsWithVersionCountsAndKindFilter(t *testing.T) {
	database := newDashboardTestDB(t)
	store := sqliteadapter.NewCatalogStore(database)
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

func TestCatalogAdapterListsAllKnownCatalogKinds(t *testing.T) {
	database := newDashboardTestDB(t)
	store := sqliteadapter.NewCatalogStore(database)
	for _, kind := range catalogKindStrings() {
		createCatalogVersion(t, store, kind, kind+"-definition", map[string]any{"kind": kind}, kind+" body", true)
	}

	adapter := NewCatalogAdapter(store)
	all, err := adapter.ListCatalog(context.Background(), "")
	if err != nil {
		t.Fatalf("list all catalog kinds: %v", err)
	}
	seen := map[string]bool{}
	for _, item := range all {
		seen[item.Kind] = true
		if item.ActiveVersion == nil || *item.ActiveVersion != 1 {
			t.Fatalf("%s active version mismatch: %+v", item.Kind, item)
		}
	}
	for _, kind := range catalogKindStrings() {
		if !seen[kind] {
			t.Fatalf("catalog list missing kind %q in %+v", kind, all)
		}

		filtered, err := adapter.ListCatalog(context.Background(), kind)
		if err != nil {
			t.Fatalf("list %s catalog: %v", kind, err)
		}
		if len(filtered) != 1 || filtered[0].Kind != kind || filtered[0].Name != kind+"-definition" {
			t.Fatalf("%s filtered summaries mismatch: %+v", kind, filtered)
		}
	}
}

func TestCatalogAdapterDetailReturnsActiveOrRequestedVersionReadOnly(t *testing.T) {
	database := newDashboardTestDB(t)
	store := sqliteadapter.NewCatalogStore(database)
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
	adapter := NewCatalogAdapter(sqliteadapter.NewCatalogStore(database))

	_, err := adapter.GetCatalogDetail(context.Background(), "chain", "missing", 0)
	if !IsNotFound(err) {
		t.Fatalf("expected not-found error, got %v", err)
	}

	response := ErrorResponse("not_found", err.Error())
	if response.Error.Code != "not_found" || response.Error.Message == "" {
		t.Fatalf("not-found response mismatch: %+v", response)
	}
}

func TestCatalogAdapterReturnsNotFoundForDefinitionWithoutActiveVersion(t *testing.T) {
	database := newDashboardTestDB(t)
	store := sqliteadapter.NewCatalogStore(database)
	createCatalogVersion(t, store, "agent", "draft-agent", map[string]any{"description": "Draft"}, "agent body", false)

	adapter := NewCatalogAdapter(store)
	_, err := adapter.GetCatalogDetail(context.Background(), "agent", "draft-agent", 0)
	if !IsNotFound(err) {
		t.Fatalf("expected not-found error for no active version, got %v", err)
	}
	if !strings.Contains(err.Error(), "agent/draft-agent") {
		t.Fatalf("not-found error should identify definition, got %q", err.Error())
	}

	versioned, err := adapter.GetCatalogDetail(context.Background(), "agent", "draft-agent", 1)
	if err != nil {
		t.Fatalf("versioned detail should still be available: %v", err)
	}
	if versioned.ActiveVersion != nil || versioned.Version != 1 {
		t.Fatalf("versioned no-active detail mismatch: %+v", versioned)
	}
}

func catalogKindStrings() []string {
	return []string{
		string(types.KindAgent),
		string(types.KindDomain),
		string(types.KindMode),
		string(types.KindChain),
		string(types.KindTeam),
		string(types.KindWorkflow),
	}
}

func createCatalogVersion(t *testing.T, store ports.CatalogStore, kind, name string, frontmatter map[string]any, body string, active bool) {
	t.Helper()
	result, err := store.CreateVersion(domain.CreateCatalogVersionInput{
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
}
