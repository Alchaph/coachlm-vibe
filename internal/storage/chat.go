package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"
)

// ChatSession represents a chat conversation session.
type ChatSession struct {
	ID        string
	CreatedAt time.Time
}

// ChatMessage represents a single message within a chat session.
type ChatMessage struct {
	ID        int64
	SessionID string
	Role      string
	Content   string
	CreatedAt time.Time
}

// validRoles defines the allowed message roles.
var validRoles = map[string]bool{
	"user":      true,
	"assistant": true,
	"system":    true,
}

// CreateSession creates a new chat session with the given ID.
// Returns an error if the ID is empty.
func (db *DB) CreateSession(id string) (*ChatSession, error) {
	if strings.TrimSpace(id) == "" {
		return nil, errors.New("session id must not be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(
		`INSERT INTO chat_sessions (id) VALUES (?)`, id,
	)
	if err != nil {
		return nil, fmt.Errorf("create session: %w", err)
	}

	var session ChatSession
	row := db.conn.QueryRow(
		`SELECT id, created_at FROM chat_sessions WHERE id = ?`, id,
	)
	if err := row.Scan(&session.ID, &session.CreatedAt); err != nil {
		return nil, fmt.Errorf("read back session: %w", err)
	}

	return &session, nil
}

// GetSession returns a chat session by ID.
// Returns (nil, nil) if the session is not found.
func (db *DB) GetSession(id string) (*ChatSession, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var session ChatSession
	err := db.conn.QueryRow(
		`SELECT id, created_at FROM chat_sessions WHERE id = ?`, id,
	).Scan(&session.ID, &session.CreatedAt)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get session: %w", err)
	}

	return &session, nil
}

// ListSessions returns all chat sessions ordered by created_at DESC (newest first).
// Returns an empty slice (not nil) when no sessions exist.
func (db *DB) ListSessions() ([]ChatSession, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(
		`SELECT id, created_at FROM chat_sessions ORDER BY created_at DESC, id DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}
	defer rows.Close()

	sessions := make([]ChatSession, 0)
	for rows.Next() {
		var s ChatSession
		if err := rows.Scan(&s.ID, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan session: %w", err)
		}
		sessions = append(sessions, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate sessions: %w", err)
	}

	return sessions, nil
}

// SaveMessage inserts a new message into a chat session and returns the created record.
// Returns an error if sessionID is empty, role is invalid, or content is empty.
func (db *DB) SaveMessage(sessionID, role, content string) (*ChatMessage, error) {
	if strings.TrimSpace(sessionID) == "" {
		return nil, errors.New("session id must not be empty")
	}
	if !validRoles[role] {
		return nil, fmt.Errorf("invalid role %q: must be one of user, assistant, system", role)
	}
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("message content must not be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.conn.Exec(
		`INSERT INTO chat_messages (session_id, role, content) VALUES (?, ?, ?)`,
		sessionID, role, content,
	)
	if err != nil {
		return nil, fmt.Errorf("save message: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	var msg ChatMessage
	row := db.conn.QueryRow(
		`SELECT id, session_id, role, content, created_at FROM chat_messages WHERE id = ?`, id,
	)
	if err := row.Scan(&msg.ID, &msg.SessionID, &msg.Role, &msg.Content, &msg.CreatedAt); err != nil {
		return nil, fmt.Errorf("read back message: %w", err)
	}

	return &msg, nil
}

// GetMessages returns all messages for a chat session ordered by created_at ASC (oldest first).
// Returns an empty slice (not nil) when no messages exist.
func (db *DB) GetMessages(sessionID string) ([]ChatMessage, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(
		`SELECT id, session_id, role, content, created_at FROM chat_messages WHERE session_id = ? ORDER BY created_at ASC, id ASC`,
		sessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("get messages: %w", err)
	}
	defer rows.Close()

	messages := make([]ChatMessage, 0)
	for rows.Next() {
		var m ChatMessage
		if err := rows.Scan(&m.ID, &m.SessionID, &m.Role, &m.Content, &m.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan message: %w", err)
		}
		messages = append(messages, m)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate messages: %w", err)
	}

	return messages, nil
}

// DeleteSession removes a chat session and all its messages (via CASCADE).
// Returns an error if the session does not exist.
func (db *DB) DeleteSession(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.conn.Exec(`DELETE FROM chat_sessions WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete session: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("session with id %q not found", id)
	}

	return nil
}
