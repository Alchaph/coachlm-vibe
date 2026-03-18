package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type CloudSyncState struct {
	Provider       string
	LastSyncedAt   time.Time
	LastChatSyncAt time.Time
	RemoteEtag     string
}

func (db *DB) GetCloudSyncState() (*CloudSyncState, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	s := &CloudSyncState{}
	var lastSynced, lastChat sql.NullTime
	err := db.conn.QueryRow(`
		SELECT provider, last_synced_at, last_chat_sync_at, remote_etag
		FROM cloud_sync_state
		WHERE id = 1`).Scan(
		&s.Provider,
		&lastSynced,
		&lastChat,
		&s.RemoteEtag,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get cloud sync state: %w", err)
	}
	if lastSynced.Valid {
		s.LastSyncedAt = lastSynced.Time
	}
	if lastChat.Valid {
		s.LastChatSyncAt = lastChat.Time
	}
	return s, nil
}

func (db *DB) SaveCloudSyncState(s *CloudSyncState) error {
	if s == nil {
		return errors.New("cloud sync state is nil")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	var lastSynced, lastChat interface{}
	if !s.LastSyncedAt.IsZero() {
		lastSynced = s.LastSyncedAt
	}
	if !s.LastChatSyncAt.IsZero() {
		lastChat = s.LastChatSyncAt
	}

	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO cloud_sync_state
			(id, provider, last_synced_at, last_chat_sync_at, remote_etag,
			 created_at, updated_at)
		VALUES
			(1, ?, ?, ?, ?,
			 COALESCE((SELECT created_at FROM cloud_sync_state WHERE id = 1), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)`,
		s.Provider,
		lastSynced,
		lastChat,
		s.RemoteEtag,
	)
	if err != nil {
		return fmt.Errorf("save cloud sync state: %w", err)
	}
	return nil
}

func (db *DB) DeleteCloudSyncState() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`DELETE FROM cloud_sync_state WHERE id = 1`)
	if err != nil {
		return fmt.Errorf("delete cloud sync state: %w", err)
	}
	return nil
}
