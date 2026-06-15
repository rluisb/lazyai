package sqlite

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestHandoffStorePersistsDocument(t *testing.T) {
	store, database := newHandoffStoreTestAdapter(t)
	doc := &types.HandoffDocument{
		ID:        "handoff-1",
		RunID:     "chain-1",
		Kind:      types.RunKindChain,
		Summary:   "Continue later",
		Recipient: "next-agent",
		CreatedAt: "2026-05-14T10:00:00Z",
		Resumable: true,
		Status: map[string]any{
			"state": "handoff",
		},
		Plan: &types.ExecutionPlan{ID: "plan-1"},
	}

	if err := store.SaveHandoffDocument(doc); err != nil {
		t.Fatalf("save handoff document: %v", err)
	}

	var docJSON string
	if err := database.QueryRow(`SELECT doc_json FROM handoffs WHERE id = ?`, doc.ID).Scan(&docJSON); err != nil {
		t.Fatalf("load handoff row: %v", err)
	}
	var saved types.HandoffDocument
	if err := json.Unmarshal([]byte(docJSON), &saved); err != nil {
		t.Fatalf("decode saved handoff: %v", err)
	}
	if saved.ID != doc.ID || saved.RunID != doc.RunID || saved.Kind != doc.Kind || saved.Summary != doc.Summary || saved.Recipient != doc.Recipient || !saved.Resumable {
		t.Fatalf("saved handoff metadata did not round-trip: %+v", saved)
	}
	if saved.Plan == nil || saved.Plan.ID != "plan-1" {
		t.Fatalf("saved handoff plan did not round-trip: %+v", saved.Plan)
	}
}

func TestHandoffStoreListsDocumentsByRunNewestFirst(t *testing.T) {
	store, _ := newHandoffStoreTestAdapter(t)
	docs := []*types.HandoffDocument{
		{ID: "handoff-older", RunID: "chain-1", Kind: types.RunKindChain, Summary: "Older", CreatedAt: "2026-05-14T10:00:00Z", Resumable: true},
		{ID: "handoff-newer", RunID: "chain-1", Kind: types.RunKindChain, Summary: "Newer", CreatedAt: "2026-05-14T11:00:00Z", Resumable: true, Plan: &types.ExecutionPlan{ID: "plan-1"}},
		{ID: "handoff-other-run", RunID: "chain-2", Kind: types.RunKindChain, Summary: "Other run", CreatedAt: "2026-05-14T12:00:00Z", Resumable: true},
		{ID: "handoff-other-kind", RunID: "chain-1", Kind: types.RunKindTeam, Summary: "Other kind", CreatedAt: "2026-05-14T12:00:00Z", Resumable: true},
	}
	for _, doc := range docs {
		if err := store.SaveHandoffDocument(doc); err != nil {
			t.Fatalf("save %s: %v", doc.ID, err)
		}
	}

	listed, err := store.ListHandoffDocuments(context.Background(), types.RunKindChain, "chain-1")
	if err != nil {
		t.Fatalf("list handoff documents: %v", err)
	}
	if len(listed) != 2 {
		t.Fatalf("listed %d handoffs, want 2: %+v", len(listed), listed)
	}
	if listed[0].ID != "handoff-newer" || listed[1].ID != "handoff-older" {
		t.Fatalf("handoffs not ordered newest first: %+v", listed)
	}
	if listed[0].Plan == nil || listed[0].Plan.ID != "plan-1" {
		t.Fatalf("listed handoff did not decode plan: %+v", listed[0].Plan)
	}
}

func newHandoffStoreTestAdapter(t *testing.T) (*HandoffStore, *db.DB) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewHandoffStore(database), database
}
