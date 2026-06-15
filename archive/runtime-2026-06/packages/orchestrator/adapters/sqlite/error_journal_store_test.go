package sqlite

import (
	"context"
	"testing"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
)

func TestErrorJournalStoreListsAndCountsEntries(t *testing.T) {
	store, database := newErrorJournalStoreTestAdapter(t)
	mustExecErrorJournal(t, database, `INSERT INTO error_journal (id, run_id, run_kind, definition_name, step_id, category, code, message, entry_json, created_at) VALUES ('err-1', 'chain-1', 'chain', 'release', 'build', 'fatal', 'boom', 'boom', '{}', '2026-05-05T10:01:00Z')`)
	mustExecErrorJournal(t, database, `INSERT INTO error_journal (id, run_id, run_kind, definition_name, step_id, category, code, message, entry_json, created_at) VALUES ('err-2', 'chain-1', 'chain', 'release', 'test', 'transient', 'retry', 'retry', '{}', '2026-05-05T10:02:00Z')`)
	mustExecErrorJournal(t, database, `INSERT INTO error_journal (id, run_id, run_kind, definition_name, step_id, category, code, message, entry_json, created_at) VALUES ('err-3', 'team-1', 'team', 'launch', NULL, 'fatal', 'team_failed', 'team failed', '{}', '2026-05-05T10:03:00Z')`)

	entries, err := store.ListErrorJournalEntries(context.Background(), "chain-1", 10)
	if err != nil {
		t.Fatalf("list entries: %v", err)
	}
	if len(entries) != 2 || entries[0].ID != "err-2" || entries[1].ID != "err-1" {
		t.Fatalf("entries not filtered or ordered newest first: %+v", entries)
	}
	if entries[0].RunKind != "chain" || entries[0].StepID != "test" {
		t.Fatalf("entry fields not decoded: %+v", entries[0])
	}

	count, err := store.CountErrorJournalEntry(context.Background(), types.RunKindChain, "chain-1")
	if err != nil {
		t.Fatalf("count entry: %v", err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}

	counts, err := store.CountErrorJournalEntriesByRun(context.Background(), []domain.RunRef{{Kind: "chain", ID: "chain-1"}, {Kind: "team", ID: "team-1"}})
	if err != nil {
		t.Fatalf("count entries by run: %v", err)
	}
	if counts[domain.RunRef{Kind: "chain", ID: "chain-1"}] != 2 || counts[domain.RunRef{Kind: "team", ID: "team-1"}] != 1 {
		t.Fatalf("counts by run mismatch: %+v", counts)
	}
}

func newErrorJournalStoreTestAdapter(t *testing.T) (*ErrorJournalStore, *db.DB) {
	t.Helper()
	database, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = database.Close() })
	if err := database.RunMigrations(); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return NewErrorJournalStore(database), database
}

func mustExecErrorJournal(t *testing.T, database *db.DB, query string) {
	t.Helper()
	if _, err := database.Exec(query); err != nil {
		t.Fatalf("exec %q: %v", query, err)
	}
}
