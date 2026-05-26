// Package session provides dispatch tracking for the LazyAI runtime.
package session

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// Dispatch represents an agent dispatch record.
type Dispatch struct {
	ID           int
	SessionID    string
	Seq          int
	ParentID     *int
	Agent        string
	Model        string
	Task         string
	Phase        string
	Workflow     string
	Mode         string
	StartedAt    *time.Time
	EndedAt      *time.Time
	Result       string
	TokenUsed    int
	ErrorMessage string
	Summary      string
	FilesTouched []string
}

// DispatchOptions configures a new dispatch.
type DispatchOptions struct {
	ParentID     *int
	Agent        string
	Model        string
	Task         string
	Phase        string
	Workflow     string
	Mode         string
	FilesTouched []string
}

// Dispatch records a new agent dispatch for the given session.
// Automatically assigns the next sequence number.
func (m *Manager) Dispatch(sessionID string, opts DispatchOptions) (*Dispatch, error) {
	startedAt := runtime.Now()

	// Get next sequence number for this session
	var nextSeq int
	err := m.db.QueryRow(
		"SELECT COALESCE(MAX(seq), 0) + 1 FROM dispatches WHERE session_id = ?",
		sessionID,
	).Scan(&nextSeq)
	if err != nil {
		return nil, fmt.Errorf("get next seq: %w", err)
	}

	filesTouched := ""
	if len(opts.FilesTouched) > 0 {
		filesTouched = strings.Join(opts.FilesTouched, ",")
	}

	result, err := m.db.Exec(
		"INSERT INTO dispatches (session_id, seq, parent_id, agent, model, task, phase, workflow, mode, started_at, files_touched) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)",
		sessionID, nextSeq, opts.ParentID, opts.Agent, opts.Model, opts.Task, opts.Phase, opts.Workflow, opts.Mode, startedAt, filesTouched,
	)
	if err != nil {
		return nil, fmt.Errorf("create dispatch: %w", err)
	}

	id, _ := result.LastInsertId()

	return m.GetDispatch(int(id))
}

// CompleteDispatch marks a dispatch as completed.
func (m *Manager) CompleteDispatch(dispatchID int, result string, tokenUsed int) error {
	endedAt := runtime.Now()

	_, err := m.db.Exec(
		"UPDATE dispatches SET ended_at = ?, result = ?, token_used = ? WHERE id = ?",
		endedAt, result, tokenUsed, dispatchID,
	)
	if err != nil {
		return fmt.Errorf("complete dispatch: %w", err)
	}

	return nil
}

// FailDispatch marks a dispatch as failed.
func (m *Manager) FailDispatch(dispatchID int, errorMessage string) error {
	endedAt := runtime.Now()

	_, err := m.db.Exec(
		"UPDATE dispatches SET ended_at = ?, error_message = ? WHERE id = ?",
		endedAt, errorMessage, dispatchID,
	)
	if err != nil {
		return fmt.Errorf("fail dispatch: %w", err)
	}

	return nil
}

// GetDispatch retrieves a dispatch by ID.
func (m *Manager) GetDispatch(dispatchID int) (*Dispatch, error) {
	var d Dispatch
	var startedAt, endedAt sql.NullString
	var parentID sql.NullInt64
	var filesTouched string

	err := m.db.QueryRow(
		"SELECT id, session_id, seq, parent_id, agent, model, task, phase, workflow, mode, started_at, ended_at, result, token_used, error_message, summary, files_touched FROM dispatches WHERE id = ?",
		dispatchID,
	).Scan(
		&d.ID, &d.SessionID, &d.Seq, &parentID, &d.Agent, &d.Model, &d.Task,
		&d.Phase, &d.Workflow, &d.Mode, &startedAt, &endedAt,
		&d.Result, &d.TokenUsed, &d.ErrorMessage, &d.Summary, &filesTouched,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("dispatch not found: %d", dispatchID)
	}
	if err != nil {
		return nil, fmt.Errorf("get dispatch: %w", err)
	}

	if parentID.Valid {
		pid := int(parentID.Int64)
		d.ParentID = &pid
	}

	if startedAt.Valid {
		t, _ := time.Parse(time.RFC3339, startedAt.String)
		d.StartedAt = &t
	}
	if endedAt.Valid {
		t, _ := time.Parse(time.RFC3339, endedAt.String)
		d.EndedAt = &t
	}

	if filesTouched != "" {
		d.FilesTouched = strings.Split(filesTouched, ",")
	}

	return &d, nil
}

// ListDispatches returns all dispatches for a session.
func (m *Manager) ListDispatches(sessionID string) ([]Dispatch, error) {
	rows, err := m.db.Query(
		"SELECT id, session_id, seq, parent_id, agent, model, task, phase, workflow, mode, started_at, ended_at, result, token_used, error_message, summary, files_touched FROM dispatches WHERE session_id = ? ORDER BY seq",
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("list dispatches: %w", err)
	}
	defer rows.Close()

	var dispatches []Dispatch
	for rows.Next() {
		var d Dispatch
		var startedAt, endedAt sql.NullString
		var parentID sql.NullInt64
		var filesTouched string

		if err := rows.Scan(
			&d.ID, &d.SessionID, &d.Seq, &parentID, &d.Agent, &d.Model, &d.Task,
			&d.Phase, &d.Workflow, &d.Mode, &startedAt, &endedAt,
			&d.Result, &d.TokenUsed, &d.ErrorMessage, &d.Summary, &filesTouched,
		); err != nil {
			continue
		}

		if parentID.Valid {
			pid := int(parentID.Int64)
			d.ParentID = &pid
		}

		if startedAt.Valid {
			t, _ := time.Parse(time.RFC3339, startedAt.String)
			d.StartedAt = &t
		}
		if endedAt.Valid {
			t, _ := time.Parse(time.RFC3339, endedAt.String)
			d.EndedAt = &t
		}

		if filesTouched != "" {
			d.FilesTouched = strings.Split(filesTouched, ",")
		}

		dispatches = append(dispatches, d)
	}

	return dispatches, nil
}

// GetLastDispatch returns the most recent dispatch for a session.
func (m *Manager) GetLastDispatch(sessionID string) (*Dispatch, error) {
	var id int
	err := m.db.QueryRow(
		"SELECT id FROM dispatches WHERE session_id = ? ORDER BY seq DESC LIMIT 1",
		sessionID,
	).Scan(&id)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get last dispatch: %w", err)
	}

	return m.GetDispatch(id)
}
