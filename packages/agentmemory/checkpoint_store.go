package agentmemory

import (
	"context"
	"database/sql"
)

// CheckpointStore persists task checkpoints.
type CheckpointStore struct{ db *sql.DB }

func NewCheckpointStore(db *sql.DB) *CheckpointStore { return &CheckpointStore{db: db} }

func (s *CheckpointStore) CreateCheckpoint(ctx context.Context, cp Checkpoint) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO checkpoints (task_id, namespace, step_id, summary, state_json, created_at)
		VALUES (?, ?, ?, ?, ?, ?)`, cp.TaskID, cp.Namespace, cp.StepID, cp.Summary, Redact(cp.StateJSON), cp.CreatedAt)
	return err
}

func (s *CheckpointStore) LatestCheckpoint(ctx context.Context, taskID string) (Checkpoint, error) {
	return scanCheckpoint(s.db.QueryRowContext(ctx, `
		SELECT id, task_id, namespace, COALESCE(step_id, ''), COALESCE(summary, ''), COALESCE(state_json, ''), created_at
		FROM checkpoints WHERE task_id = ? ORDER BY id DESC LIMIT 1`, taskID))
}

func (s *CheckpointStore) ListCheckpoints(ctx context.Context, taskID string) ([]Checkpoint, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, task_id, namespace, COALESCE(step_id, ''), COALESCE(summary, ''), COALESCE(state_json, ''), created_at
		FROM checkpoints WHERE task_id = ? ORDER BY id ASC`, taskID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	checkpoints := []Checkpoint{}
	for rows.Next() {
		cp, err := scanCheckpoint(rows)
		if err != nil {
			return nil, err
		}
		checkpoints = append(checkpoints, cp)
	}
	return checkpoints, rows.Err()
}

type checkpointScanner interface{ Scan(dest ...any) error }

func scanCheckpoint(row checkpointScanner) (Checkpoint, error) {
	var cp Checkpoint
	err := row.Scan(&cp.ID, &cp.TaskID, &cp.Namespace, &cp.StepID, &cp.Summary, &cp.StateJSON, &cp.CreatedAt)
	return cp, err
}
