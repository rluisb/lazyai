// Package session provides session lifecycle management for the LazyAI runtime.
package session

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// Manager handles session lifecycle operations
type Manager struct {
	db *runtime.DB
}

// NewManager creates a session manager backed by the given database.
func NewManager(db *runtime.DB) *Manager {
	return &Manager{db: db}
}

// SessionStatus represents the lifecycle state of a session.
type SessionStatus string

const (
	SessionActive SessionStatus = "active"
	SessionEnded  SessionStatus = "ended"
	SessionFailed SessionStatus = "failed"
)

// Session represents an AI agent session.
type Session struct {
	ID         string
	StartedAt  time.Time
	EndedAt    *time.Time
	Agent      string
	Model      string
	Goal       string
	Repo       string
	Worktree   string
	Status     SessionStatus
	TokenTotal int
	Summary    string
	Tags       []string
}

// StartOptions configures a new session.
type StartOptions struct {
	Agent    string
	Model    string
	Repo     string
	Worktree string
	Tags     []string
}

// Start creates a new session with the given goal.
func (m *Manager) Start(goal string, opts StartOptions) (*Session, error) {
	sessionID := fmt.Sprintf("ses_%d", time.Now().Unix())
	startedAt := runtime.Now()

	agent := opts.Agent
	if agent == "" {
		agent = "primary-agent"
	}

	tags := ""
	if len(opts.Tags) > 0 {
		tags = strings.Join(opts.Tags, ",")
	}

	_, err := m.db.Exec(
		"INSERT INTO sessions (id, started_at, agent, model, goal, repo, worktree, status, tags) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		sessionID, startedAt, agent, opts.Model, goal, opts.Repo, opts.Worktree, string(SessionActive), tags,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	return m.Get(sessionID)
}

// End terminates a session by ID.
func (m *Manager) End(sessionID string) error {
	endedAt := runtime.Now()

	result, err := m.db.Exec(
		"UPDATE sessions SET status = ?, ended_at = ? WHERE id = ?",
		string(SessionEnded), endedAt, sessionID,
	)
	if err != nil {
		return fmt.Errorf("end session: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return nil
}

// Get retrieves a session by ID.
func (m *Manager) Get(sessionID string) (*Session, error) {
	var s Session
	var endedAt sql.NullString
	var model sql.NullString
	var repo sql.NullString
	var worktree sql.NullString
	var summary sql.NullString
	var tags sql.NullString
	var startedAt string

	err := m.db.QueryRow(
		"SELECT id, started_at, ended_at, agent, model, goal, repo, worktree, status, token_total, summary, tags FROM sessions WHERE id = ?",
		sessionID,
	).Scan(
		&s.ID, &startedAt, &endedAt, &s.Agent, &model, &s.Goal, &repo, &worktree,
		&s.Status, &s.TokenTotal, &summary, &tags,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	parsedStartedAt, err := time.Parse(time.RFC3339, startedAt)
	if err != nil {
		return nil, fmt.Errorf("parse started_at: %w", err)
	}
	s.StartedAt = parsedStartedAt

	if endedAt.Valid {
		t, _ := time.Parse(time.RFC3339, endedAt.String)
		s.EndedAt = &t
	}
	if model.Valid {
		s.Model = model.String
	}
	if repo.Valid {
		s.Repo = repo.String
	}
	if worktree.Valid {
		s.Worktree = worktree.String
	}
	if summary.Valid {
		s.Summary = summary.String
	}

	if tags.Valid && tags.String != "" {
		s.Tags = strings.Split(tags.String, ",")
	}

	return &s, nil
}

// List returns all sessions ordered by start time (newest first).
func (m *Manager) List() ([]Session, error) {
	rows, err := m.db.Query(
		"SELECT id, started_at, ended_at, agent, model, goal, repo, worktree, status, token_total, summary, tags FROM sessions ORDER BY started_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		var endedAt sql.NullString
		var model sql.NullString
		var repo sql.NullString
		var worktree sql.NullString
		var summary sql.NullString
		var tags sql.NullString
		var startedAt string

		if err := rows.Scan(
			&s.ID, &startedAt, &endedAt, &s.Agent, &model, &s.Goal, &repo, &worktree,
			&s.Status, &s.TokenTotal, &summary, &tags,
		); err != nil {
			continue
		}

		parsedStartedAt, err := time.Parse(time.RFC3339, startedAt)
		if err != nil {
			continue
		}
		s.StartedAt = parsedStartedAt

		if endedAt.Valid {
			t, _ := time.Parse(time.RFC3339, endedAt.String)
			s.EndedAt = &t
		}
		if model.Valid {
			s.Model = model.String
		}
		if repo.Valid {
			s.Repo = repo.String
		}
		if worktree.Valid {
			s.Worktree = worktree.String
		}
		if summary.Valid {
			s.Summary = summary.String
		}

		if tags.Valid && tags.String != "" {
			s.Tags = strings.Split(tags.String, ",")
		}

		sessions = append(sessions, s)
	}

	return sessions, nil
}

// UpdateSummary updates the summary for a session.
func (m *Manager) UpdateSummary(sessionID string, summary string) error {
	_, err := m.db.Exec(
		"UPDATE sessions SET summary = ? WHERE id = ?",
		summary, sessionID,
	)
	if err != nil {
		return fmt.Errorf("update summary: %w", err)
	}
	return nil
}

// AddTags appends tags to a session.
func (m *Manager) AddTags(sessionID string, newTags []string) error {
	if len(newTags) == 0 {
		return nil
	}

	var existingTags string
	err := m.db.QueryRow("SELECT tags FROM sessions WHERE id = ?", sessionID).Scan(&existingTags)
	if err != nil {
		return fmt.Errorf("get existing tags: %w", err)
	}

	var tagSet []string
	if existingTags != "" {
		tagSet = strings.Split(existingTags, ",")
	}

	// Deduplicate
	seen := make(map[string]bool)
	for _, t := range tagSet {
		seen[t] = true
	}
	for _, t := range newTags {
		if !seen[t] {
			tagSet = append(tagSet, t)
			seen[t] = true
		}
	}

	_, err = m.db.Exec(
		"UPDATE sessions SET tags = ? WHERE id = ?",
		strings.Join(tagSet, ","), sessionID,
	)
	if err != nil {
		return fmt.Errorf("update tags: %w", err)
	}

	return nil
}
