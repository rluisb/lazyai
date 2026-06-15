package sqlite

import (
	"context"
	"database/sql"
	"strings"

	"github.com/rluisb/lazyai/packages/orchestrator/domain"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.ErrorJournalStore = (*ErrorJournalStore)(nil)

// ErrorJournalStore queries persisted error journal entries in SQLite.
type ErrorJournalStore struct {
	database *db.DB
}

// NewErrorJournalStore creates a SQLite-backed error journal store adapter.
func NewErrorJournalStore(database *db.DB) *ErrorJournalStore {
	return &ErrorJournalStore{database: database}
}

// ListErrorJournalEntries returns bounded error journal entries, optionally scoped to one run.
func (s *ErrorJournalStore) ListErrorJournalEntries(ctx context.Context, runID string, limit int) ([]domain.ErrorJournalEntry, error) {
	args := []any{}
	where := ""
	if runID != "" {
		where = `WHERE run_id = ?`
		args = append(args, runID)
	}
	args = append(args, limit)
	rows, err := s.database.QueryContext(ctx, `
		SELECT id, run_id, run_kind, definition_name, step_id, category, code, message, created_at
		FROM error_journal `+where+`
		ORDER BY created_at DESC, id DESC
		LIMIT ?`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []domain.ErrorJournalEntry
	for rows.Next() {
		var entry domain.ErrorJournalEntry
		var runID, runKind, stepID sql.NullString
		if err := rows.Scan(&entry.ID, &runID, &runKind, &entry.DefinitionName, &stepID, &entry.Category, &entry.Code, &entry.Message, &entry.CreatedAt); err != nil {
			return nil, err
		}
		if runID.Valid {
			entry.RunID = runID.String
		}
		if runKind.Valid {
			entry.RunKind = runKind.String
		}
		if stepID.Valid {
			entry.StepID = stepID.String
		}
		entries = append(entries, entry)
	}
	return entries, rows.Err()
}

// CountErrorJournalEntry returns the number of journal entries for one run.
func (s *ErrorJournalStore) CountErrorJournalEntry(ctx context.Context, kind types.RunKind, id string) (int, error) {
	var count int
	err := s.database.QueryRowContext(ctx, `SELECT COUNT(*) FROM error_journal WHERE run_id = ? AND run_kind = ?`, id, string(kind)).Scan(&count)
	return count, err
}

// CountErrorJournalEntriesByRun returns error counts keyed by run identity for the provided runs.
func (s *ErrorJournalStore) CountErrorJournalEntriesByRun(ctx context.Context, refs []domain.RunRef) (map[domain.RunRef]int, error) {
	counts := map[domain.RunRef]int{}
	if len(refs) == 0 {
		return counts, nil
	}
	runIDs := make([]string, 0, len(refs))
	seen := map[string]bool{}
	for _, ref := range refs {
		if ref.ID == "" || seen[ref.ID] {
			continue
		}
		runIDs = append(runIDs, ref.ID)
		seen[ref.ID] = true
	}
	if len(runIDs) == 0 {
		return counts, nil
	}

	query := `SELECT run_id, run_kind, COUNT(*) FROM error_journal WHERE run_id IN (` + sqlitePlaceholders(len(runIDs)) + `) AND run_kind IS NOT NULL GROUP BY run_id, run_kind`
	args := make([]any, 0, len(runIDs))
	for _, id := range runIDs {
		args = append(args, id)
	}
	rows, err := s.database.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id, kind string
		var count int
		if err := rows.Scan(&id, &kind, &count); err != nil {
			return nil, err
		}
		counts[domain.RunRef{Kind: kind, ID: id}] = count
	}
	return counts, rows.Err()
}

func sqlitePlaceholders(count int) string {
	return strings.TrimRight(strings.Repeat("?,", count), ",")
}
