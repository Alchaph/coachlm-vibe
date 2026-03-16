package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

func (db *DB) SaveTokens(accessToken, refreshToken []byte, expiresAt time.Time) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO oauth_tokens
			(id, access_token, refresh_token, token_expires_at, created_at, updated_at)
		VALUES
			(1, ?, ?, ?, COALESCE((SELECT created_at FROM oauth_tokens WHERE id = 1), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)`,
		accessToken,
		refreshToken,
		expiresAt,
	)
	if err != nil {
		return fmt.Errorf("save tokens: %w", err)
	}
	return nil
}

func (db *DB) GetTokens() (accessToken, refreshToken []byte, expiresAt time.Time, err error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	err = db.conn.QueryRow(`
		SELECT access_token, refresh_token, token_expires_at
		FROM oauth_tokens
		WHERE id = 1`).Scan(
		&accessToken,
		&refreshToken,
		&expiresAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil, time.Time{}, nil
	}
	if err != nil {
		return nil, nil, time.Time{}, fmt.Errorf("get tokens: %w", err)
	}
	return accessToken, refreshToken, expiresAt, nil
}

func (db *DB) DeleteTokens() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`DELETE FROM oauth_tokens WHERE id = 1`)
	if err != nil {
		return fmt.Errorf("delete tokens: %w", err)
	}
	return nil
}
