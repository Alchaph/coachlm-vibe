package storage

import (
	"errors"
	"fmt"
	"strings"
	"time"
)

// PinnedInsight represents a coaching insight saved from chat.
// CRITICAL: Pinned insights are NEVER compressed or dropped (AGENTS.md constraint).
type PinnedInsight struct {
	ID              int64
	Content         string
	SourceSessionID string
	CreatedAt       time.Time
}

// SaveInsight inserts a new pinned insight and returns the created record.
// Returns an error if content is empty.
func (db *DB) SaveInsight(content string, sourceSessionID string) (*PinnedInsight, error) {
	if strings.TrimSpace(content) == "" {
		return nil, errors.New("insight content must not be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.conn.Exec(
		`INSERT INTO pinned_insights (content, source_session_id) VALUES (?, ?)`,
		content, sourceSessionID,
	)
	if err != nil {
		return nil, fmt.Errorf("insert insight: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("get last insert id: %w", err)
	}

	var insight PinnedInsight
	row := db.conn.QueryRow(
		`SELECT id, content, source_session_id, created_at FROM pinned_insights WHERE id = ?`, id,
	)
	if err := row.Scan(&insight.ID, &insight.Content, &insight.SourceSessionID, &insight.CreatedAt); err != nil {
		return nil, fmt.Errorf("read back insight: %w", err)
	}

	return &insight, nil
}

// GetInsights returns all pinned insights ordered by created_at ASC (oldest first).
// Returns an empty slice (not nil) when no insights exist.
func (db *DB) GetInsights() ([]PinnedInsight, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(
		`SELECT id, content, source_session_id, created_at FROM pinned_insights ORDER BY created_at ASC, id ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query insights: %w", err)
	}
	defer rows.Close()

	insights := make([]PinnedInsight, 0)
	for rows.Next() {
		var i PinnedInsight
		if err := rows.Scan(&i.ID, &i.Content, &i.SourceSessionID, &i.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan insight: %w", err)
		}
		insights = append(insights, i)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate insights: %w", err)
	}

	return insights, nil
}

// DeleteInsight removes a pinned insight by ID.
// Returns an error if the insight does not exist.
func (db *DB) DeleteInsight(id int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	result, err := db.conn.Exec(`DELETE FROM pinned_insights WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete insight: %w", err)
	}

	n, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("rows affected: %w", err)
	}
	if n == 0 {
		return fmt.Errorf("insight with id %d not found", id)
	}

	return nil
}

// InsightExists checks whether an insight with the exact same content already exists.
func (db *DB) InsightExists(content string) (bool, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var count int
	err := db.conn.QueryRow(
		`SELECT COUNT(*) FROM pinned_insights WHERE content = ?`, content,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("check insight exists: %w", err)
	}

	return count > 0, nil
}

// ReplaceAllContext deletes all context data (profile, activities, insights) but keeps settings.
// Used for "replace all" import mode.
func (db *DB) ReplaceAllContext() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`DELETE FROM athlete_profile WHERE id = 1`)
	if err != nil {
		return fmt.Errorf("delete profile: %w", err)
	}

	_, err = db.conn.Exec(`DELETE FROM activities`)
	if err != nil {
		return fmt.Errorf("delete activities: %w", err)
	}

	_, err = db.conn.Exec(`DELETE FROM pinned_insights`)
	if err != nil {
		return fmt.Errorf("delete insights: %w", err)
	}

	return nil
}
