package storage

import (
	"database/sql"
	"errors"
	"fmt"
)

type Settings struct {
	ClaudeAPIKey       []byte
	OpenAIAPIKey       []byte
	ActiveLLM          string
	OllamaEndpoint     string
	StravaClientID     []byte
	StravaClientSecret []byte
}

var validLLMs = map[string]bool{
	"claude": true,
	"openai": true,
	"local":  true,
}

func validateSettings(s *Settings) error {
	if s == nil {
		return errors.New("settings is nil")
	}
	if !validLLMs[s.ActiveLLM] {
		return fmt.Errorf("active_llm must be one of claude, openai, local; got %q", s.ActiveLLM)
	}
	return nil
}

func (db *DB) SaveSettings(s *Settings) error {
	if err := validateSettings(s); err != nil {
		return fmt.Errorf("validate settings: %w", err)
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO settings
			(id, claude_api_key, openai_api_key, active_llm, ollama_endpoint, strava_client_id, strava_client_secret, created_at, updated_at)
		VALUES
			(1, ?, ?, ?, ?, ?, ?, COALESCE((SELECT created_at FROM settings WHERE id = 1), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)`,
		s.ClaudeAPIKey,
		s.OpenAIAPIKey,
		s.ActiveLLM,
		s.OllamaEndpoint,
		s.StravaClientID,
		s.StravaClientSecret,
	)
	if err != nil {
		return fmt.Errorf("save settings: %w", err)
	}
	return nil
}

func (db *DB) GetSettings() (*Settings, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	s := &Settings{}
	err := db.conn.QueryRow(`
		SELECT claude_api_key, openai_api_key, active_llm, ollama_endpoint, strava_client_id, strava_client_secret
		FROM settings
		WHERE id = 1`).Scan(
		&s.ClaudeAPIKey,
		&s.OpenAIAPIKey,
		&s.ActiveLLM,
		&s.OllamaEndpoint,
		&s.StravaClientID,
		&s.StravaClientSecret,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	return s, nil
}
