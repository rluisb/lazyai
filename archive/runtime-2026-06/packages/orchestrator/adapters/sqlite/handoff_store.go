package sqlite

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/rluisb/lazyai/packages/orchestrator/internal/db"
	"github.com/rluisb/lazyai/packages/orchestrator/internal/types"
	"github.com/rluisb/lazyai/packages/orchestrator/ports"
)

var _ ports.HandoffStore = (*HandoffStore)(nil)
var _ ports.HandoffQueryStore = (*HandoffStore)(nil)

// HandoffStore persists resumable handoff documents in SQLite.
type HandoffStore struct {
	database *db.DB
}

// NewHandoffStore creates a SQLite-backed handoff document store adapter.
func NewHandoffStore(database *db.DB) *HandoffStore {
	return &HandoffStore{database: database}
}

// SaveHandoffDocument records a resumable handoff document.
func (s *HandoffStore) SaveHandoffDocument(doc *types.HandoffDocument) error {
	encoded, err := json.Marshal(doc)
	if err != nil {
		return fmt.Errorf("marshal handoff: %w", err)
	}
	_, err = s.database.Exec(`INSERT OR IGNORE INTO handoffs (id, run_id, run_kind, doc_json, created_at) VALUES (?, ?, ?, ?, ?)`, doc.ID, doc.RunID, doc.Kind, string(encoded), doc.CreatedAt)
	if err != nil {
		return fmt.Errorf("save handoff %s: %w", doc.ID, err)
	}
	return nil
}

// ListHandoffDocuments returns decodable handoff documents for a run, newest first.
func (s *HandoffStore) ListHandoffDocuments(ctx context.Context, kind types.RunKind, runID string) ([]types.HandoffDocument, error) {
	rows, err := s.database.QueryContext(ctx, `SELECT doc_json FROM handoffs WHERE run_id = ? AND run_kind = ? ORDER BY created_at DESC`, runID, string(kind))
	if err != nil {
		return nil, fmt.Errorf("list handoffs for %s/%s: %w", kind, runID, err)
	}
	defer rows.Close()

	var docs []types.HandoffDocument
	for rows.Next() {
		var docJSON string
		if err := rows.Scan(&docJSON); err != nil {
			return nil, fmt.Errorf("scan handoff for %s/%s: %w", kind, runID, err)
		}
		var doc types.HandoffDocument
		if err := json.Unmarshal([]byte(docJSON), &doc); err != nil {
			continue
		}
		docs = append(docs, doc)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate handoffs for %s/%s: %w", kind, runID, err)
	}
	return docs, nil
}
