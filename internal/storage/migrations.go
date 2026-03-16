package storage

import "fmt"

var migrations = []string{
	`CREATE TABLE IF NOT EXISTS athlete_profile (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		age INTEGER,
		max_hr INTEGER,
		threshold_pace_secs INTEGER,
		weekly_mileage_target REAL,
		race_goals TEXT,
		injury_history TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS oauth_tokens (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		access_token BLOB NOT NULL,
		refresh_token BLOB NOT NULL,
		token_expires_at DATETIME NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS activities (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		strava_id INTEGER UNIQUE,
		name TEXT,
		activity_type TEXT,
		start_date DATETIME,
		distance REAL,
		duration_secs INTEGER,
		avg_pace_secs INTEGER,
		avg_hr INTEGER,
		max_hr INTEGER,
		avg_cadence REAL,
		source TEXT NOT NULL DEFAULT 'strava',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS activity_streams (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		activity_id INTEGER NOT NULL,
		stream_type TEXT NOT NULL,
		data TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (activity_id) REFERENCES activities(id) ON DELETE CASCADE
	)`,

	`CREATE TABLE IF NOT EXISTS pinned_insights (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		content TEXT NOT NULL,
		source_session_id TEXT,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS chat_sessions (
		id TEXT PRIMARY KEY,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,

	`CREATE TABLE IF NOT EXISTS chat_messages (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		session_id TEXT NOT NULL,
		role TEXT NOT NULL,
		content TEXT NOT NULL,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		FOREIGN KEY (session_id) REFERENCES chat_sessions(id) ON DELETE CASCADE
	)`,

	`CREATE TABLE IF NOT EXISTS settings (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		claude_api_key BLOB,
		openai_api_key BLOB,
		active_llm TEXT NOT NULL DEFAULT 'claude',
		ollama_endpoint TEXT NOT NULL DEFAULT 'http://localhost:11434',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
}

func (db *DB) migrate() error {
	for i, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			return fmt.Errorf("migration %d: %w", i, err)
		}
	}
	return nil
}
