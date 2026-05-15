package agentmemory

import (
	"context"
	"database/sql"
)

// MemoryStore persists durable agent memories.
type MemoryStore struct{ db *sql.DB }

func NewMemoryStore(db *sql.DB) *MemoryStore { return &MemoryStore{db: db} }

func (s *MemoryStore) SaveMemory(ctx context.Context, memory Memory) (int64, error) {
	importance := memory.Importance
	if importance == "" {
		importance = ImportanceNormal
	}
	result, err := s.db.ExecContext(ctx, `
		INSERT INTO memories (namespace, content, source_task_id, source_step_id, tags, importance, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)`, memory.Namespace, Redact(memory.Content), memory.SourceTaskID,
		memory.SourceStepID, memory.Tags, importance, memory.CreatedAt)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func (s *MemoryStore) ListMemories(ctx context.Context, namespace string, tags string, limit int) ([]Memory, error) {
	query := `
		SELECT id, namespace, content, COALESCE(source_task_id, ''), COALESCE(source_step_id, ''), COALESCE(tags, ''), importance, created_at
		FROM memories WHERE namespace = ?`
	args := []any{namespace}
	if tags != "" {
		query += ` AND (tags = ? OR instr(',' || tags || ',', ',' || ? || ',') > 0)`
		args = append(args, tags, tags)
	}
	query += ` ORDER BY id DESC LIMIT ?`
	args = append(args, normalizeLimit(limit))

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMemories(rows)
}

func scanMemories(rows *sql.Rows) ([]Memory, error) {
	memories := []Memory{}
	for rows.Next() {
		memory, err := scanMemory(rows)
		if err != nil {
			return nil, err
		}
		memories = append(memories, memory)
	}
	return memories, rows.Err()
}

type memoryScanner interface{ Scan(dest ...any) error }

func scanMemory(row memoryScanner) (Memory, error) {
	var memory Memory
	err := row.Scan(&memory.ID, &memory.Namespace, &memory.Content, &memory.SourceTaskID,
		&memory.SourceStepID, &memory.Tags, &memory.Importance, &memory.CreatedAt)
	return memory, err
}

func normalizeLimit(limit int) int {
	if limit <= 0 {
		return 50
	}
	return limit
}
