// Package session provides inter-agent messaging for the LazyAI runtime.
package session

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/rluisb/lazyai/packages/cli/internal/runtime"
)

// Message represents an inter-agent message.
type Message struct {
	ID        int
	SessionID string
	FromAgent string
	ToAgent   *string
	Subject   string
	Body      string
	Priority  string
	Status    string
	CreatedAt time.Time
	ReadAt    *time.Time
}

// SendMessage creates a new message in the session.
func (m *Manager) SendMessage(sessionID string, fromAgent string, toAgent *string, subject string, body string, priority string) (*Message, error) {
	createdAt := runtime.Now()

	result, err := m.db.Exec(
		"INSERT INTO messages (session_id, from_agent, to_agent, subject, body, priority, status, created_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		sessionID, fromAgent, toAgent, subject, body, priority, "unread", createdAt,
	)
	if err != nil {
		return nil, fmt.Errorf("send message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("getting last insert id: %w", err)
	}
	return m.GetMessage(int(id))
}

// GetMessage retrieves a message by ID.
func (m *Manager) GetMessage(messageID int) (*Message, error) {
	var msg Message
	var toAgent sql.NullString
	var readAt sql.NullString

	err := m.db.QueryRow(
		"SELECT id, session_id, from_agent, to_agent, subject, body, priority, status, created_at, read_at FROM messages WHERE id = ?",
		messageID,
	).Scan(
		&msg.ID, &msg.SessionID, &msg.FromAgent, &toAgent, &msg.Subject, &msg.Body,
		&msg.Priority, &msg.Status, &msg.CreatedAt, &readAt,
	)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("message not found: %d", messageID)
	}
	if err != nil {
		return nil, fmt.Errorf("get message: %w", err)
	}

	if toAgent.Valid {
		msg.ToAgent = &toAgent.String
	}
	if readAt.Valid {
		t, _ := time.Parse(time.RFC3339, readAt.String)
		msg.ReadAt = &t
	}

	return &msg, nil
}

// GetMessages returns messages for a session, optionally filtered by recipient.
func (m *Manager) GetMessages(sessionID string, toAgent *string) ([]Message, error) {
	var rows *sql.Rows
	var err error

	if toAgent != nil {
		rows, err = m.db.Query(
			"SELECT id, session_id, from_agent, to_agent, subject, body, priority, status, created_at, read_at FROM messages WHERE session_id = ? AND (to_agent = ? OR to_agent IS NULL) ORDER BY created_at DESC",
			sessionID, *toAgent,
		)
	} else {
		rows, err = m.db.Query(
			"SELECT id, session_id, from_agent, to_agent, subject, body, priority, status, created_at, read_at FROM messages WHERE session_id = ? ORDER BY created_at DESC",
			sessionID,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	defer rows.Close()

	var messages []Message
	for rows.Next() {
		var msg Message
		var toAgent sql.NullString
		var readAt sql.NullString

		if err := rows.Scan(
			&msg.ID, &msg.SessionID, &msg.FromAgent, &toAgent, &msg.Subject, &msg.Body,
			&msg.Priority, &msg.Status, &msg.CreatedAt, &readAt,
		); err != nil {
			continue
		}

		if toAgent.Valid {
			msg.ToAgent = &toAgent.String
		}
		if readAt.Valid {
			t, _ := time.Parse(time.RFC3339, readAt.String)
			msg.ReadAt = &t
		}

		messages = append(messages, msg)
	}

	return messages, nil
}

// MarkRead marks a message as read.
func (m *Manager) MarkRead(messageID int) error {
	readAt := runtime.Now()

	_, err := m.db.Exec(
		"UPDATE messages SET status = ?, read_at = ? WHERE id = ?",
		"read", readAt, messageID,
	)
	if err != nil {
		return fmt.Errorf("mark read: %w", err)
	}

	return nil
}

// GetUnreadCount returns the number of unread messages for an agent.
func (m *Manager) GetUnreadCount(sessionID string, agent string) (int, error) {
	var count int
	err := m.db.QueryRow(
		"SELECT COUNT(*) FROM messages WHERE session_id = ? AND to_agent = ? AND status = ?",
		sessionID, agent, "unread",
	).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("get unread count: %w", err)
	}

	return count, nil
}
