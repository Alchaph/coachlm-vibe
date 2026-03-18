package storage

import (
	"database/sql"
	"fmt"
	"sync"

	_ "modernc.org/sqlite"
)

type DB struct {
	conn *sql.DB
	mu   sync.RWMutex
}

func New(dbPath string) (*DB, error) {
	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open database: %w", err)
	}

	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set journal mode: %w", err)
	}

	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	db := &DB{conn: conn}
	if err := db.migrate(); err != nil {
		conn.Close()
		return nil, fmt.Errorf("run migrations: %w", err)
	}

	return db, nil
}

func (db *DB) Close() error {
	return db.conn.Close()
}

func (db *DB) Conn() *sql.DB {
	return db.conn
}

// ResetAll deletes all user data from every table, returning the app to first-run state.
func (db *DB) ResetAll() error {
	db.mu.Lock()
	defer db.mu.Unlock()

	tx, err := db.conn.Begin()
	if err != nil {
		return fmt.Errorf("begin reset transaction: %w", err)
	}
	defer tx.Rollback()

	tables := []string{
		"plan_sessions",
		"plan_weeks",
		"training_plans",
		"races",
		"activity_streams",
		"activities",
		"chat_messages",
		"chat_sessions",
		"pinned_insights",
		"athlete_stats",
		"gear",
		"cloud_sync_state",
		"oauth_tokens",
		"athlete_profile",
		"settings",
	}

	for _, t := range tables {
		if _, err := tx.Exec("DELETE FROM " + t); err != nil {
			return fmt.Errorf("reset table %s: %w", t, err)
		}
	}

	return tx.Commit()
}
