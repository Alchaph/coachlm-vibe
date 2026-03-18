package storage

import (
	"fmt"
	"strings"
)

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

	`ALTER TABLE settings ADD COLUMN strava_client_id BLOB`,
	`ALTER TABLE settings ADD COLUMN strava_client_secret BLOB`,

	`ALTER TABLE settings ADD COLUMN claude_model TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE settings ADD COLUMN openai_model TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE settings ADD COLUMN ollama_model TEXT NOT NULL DEFAULT ''`,

	`ALTER TABLE athlete_profile ADD COLUMN experience_level TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE athlete_profile ADD COLUMN training_days_per_week INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE athlete_profile ADD COLUMN resting_hr INTEGER NOT NULL DEFAULT 0`,
	`ALTER TABLE athlete_profile ADD COLUMN preferred_terrain TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE settings ADD COLUMN custom_system_prompt TEXT NOT NULL DEFAULT ''`,

	`UPDATE settings SET active_llm = 'gemini' WHERE active_llm IN ('claude', 'openai', 'free')`,

	// Cloud sync (S21)
	`CREATE TABLE IF NOT EXISTS cloud_sync_state (
		id INTEGER PRIMARY KEY CHECK (id = 1),
		provider TEXT NOT NULL DEFAULT '',
		last_synced_at DATETIME,
		last_chat_sync_at DATETIME,
		remote_etag TEXT NOT NULL DEFAULT '',
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`ALTER TABLE settings ADD COLUMN cloud_provider TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE settings ADD COLUMN cloud_endpoint TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE settings ADD COLUMN cloud_bucket TEXT NOT NULL DEFAULT ''`,
	`ALTER TABLE settings ADD COLUMN cloud_access_key BLOB`,
	`ALTER TABLE settings ADD COLUMN cloud_secret_key BLOB`,
	`ALTER TABLE settings ADD COLUMN gdrive_access_token BLOB`,
	`ALTER TABLE settings ADD COLUMN gdrive_refresh_token BLOB`,
	`ALTER TABLE settings ADD COLUMN gdrive_token_expiry DATETIME`,
	`ALTER TABLE settings ADD COLUMN gdrive_client_id TEXT NOT NULL DEFAULT ''`,

	// Ollama-only migration: gemini backend removed
	`UPDATE settings SET active_llm = 'local' WHERE active_llm = 'gemini'`,

	// S41: Heart rate zones from Strava
	`ALTER TABLE athlete_profile ADD COLUMN heart_rate_zones TEXT NOT NULL DEFAULT ''`,

	// S22: Training plan generation
	`CREATE TABLE IF NOT EXISTS races (
		id TEXT PRIMARY KEY,
		name TEXT NOT NULL,
		distance_km REAL NOT NULL,
		race_date DATETIME NOT NULL,
		terrain TEXT NOT NULL DEFAULT 'road',
		elevation_m REAL,
		goal_time_s INTEGER,
		priority TEXT NOT NULL DEFAULT 'A',
		is_active INTEGER NOT NULL DEFAULT 0,
		created_at DATETIME DEFAULT CURRENT_TIMESTAMP
	)`,
	`CREATE TABLE IF NOT EXISTS training_plans (
		id TEXT PRIMARY KEY,
		race_id TEXT NOT NULL,
		generated_at DATETIME NOT NULL,
		llm_backend TEXT NOT NULL DEFAULT '',
		prompt_hash TEXT NOT NULL DEFAULT '',
		FOREIGN KEY (race_id) REFERENCES races(id) ON DELETE CASCADE
	)`,
	`CREATE TABLE IF NOT EXISTS plan_weeks (
		id TEXT PRIMARY KEY,
		plan_id TEXT NOT NULL,
		week_number INTEGER NOT NULL,
		week_start DATETIME NOT NULL,
		FOREIGN KEY (plan_id) REFERENCES training_plans(id) ON DELETE CASCADE,
		UNIQUE(plan_id, week_number)
	)`,
	`CREATE TABLE IF NOT EXISTS plan_sessions (
		id TEXT PRIMARY KEY,
		week_id TEXT NOT NULL,
		day_of_week INTEGER NOT NULL,
		session_type TEXT NOT NULL DEFAULT 'rest',
		duration_min INTEGER NOT NULL DEFAULT 0,
		distance_km REAL NOT NULL DEFAULT 0,
		hr_zone INTEGER NOT NULL DEFAULT 0,
		pace_min_low REAL NOT NULL DEFAULT 0,
		pace_min_high REAL NOT NULL DEFAULT 0,
		notes TEXT NOT NULL DEFAULT '',
		status TEXT NOT NULL DEFAULT 'planned',
		actual_duration_min INTEGER,
		actual_distance_km REAL,
		completed_at DATETIME,
		FOREIGN KEY (week_id) REFERENCES plan_weeks(id) ON DELETE CASCADE
	)`,
}

func (db *DB) migrate() error {
	for i, m := range migrations {
		if _, err := db.conn.Exec(m); err != nil {
			// ALTER TABLE ADD COLUMN fails harmlessly when column already exists.
			if strings.HasPrefix(strings.TrimSpace(strings.ToUpper(m)), "ALTER") && strings.Contains(err.Error(), "duplicate column") {
				continue
			}
			return fmt.Errorf("migration %d: %w", i, err)
		}
	}
	return nil
}
