package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Settings struct {
	ActiveLLM          string
	OllamaEndpoint     string
	OllamaModel        string
	CustomSystemPrompt string

	CloudProvider      string
	CloudEndpoint      string
	CloudBucket        string
	CloudAccessKey     []byte
	CloudSecretKey     []byte
	GDriveAccessToken  []byte
	GDriveRefreshToken []byte
	GDriveTokenExpiry  time.Time
	GDriveClientID     string
}

var validLLMs = map[string]bool{
	"local": true,
}

func validateSettings(s *Settings) error {
	if s == nil {
		return errors.New("settings is nil")
	}
	if !validLLMs[s.ActiveLLM] {
		return fmt.Errorf("active_llm must be local; got %q", s.ActiveLLM)
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
			(id, active_llm, ollama_endpoint, ollama_model, custom_system_prompt,
			 cloud_provider, cloud_endpoint, cloud_bucket, cloud_access_key, cloud_secret_key,
			 gdrive_access_token, gdrive_refresh_token, gdrive_token_expiry, gdrive_client_id,
			 created_at, updated_at)
		VALUES
			(1, ?, ?, ?, ?,
			 ?, ?, ?, ?, ?,
			 ?, ?, ?, ?,
			 COALESCE((SELECT created_at FROM settings WHERE id = 1), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)`,
		s.ActiveLLM,
		s.OllamaEndpoint,
		s.OllamaModel,
		s.CustomSystemPrompt,
		s.CloudProvider,
		s.CloudEndpoint,
		s.CloudBucket,
		s.CloudAccessKey,
		s.CloudSecretKey,
		s.GDriveAccessToken,
		s.GDriveRefreshToken,
		s.GDriveTokenExpiry,
		s.GDriveClientID,
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
	var gdriveExpiry sql.NullTime
	err := db.conn.QueryRow(`
		SELECT active_llm, ollama_endpoint, ollama_model, custom_system_prompt,
		       cloud_provider, cloud_endpoint, cloud_bucket, cloud_access_key, cloud_secret_key,
		       gdrive_access_token, gdrive_refresh_token, gdrive_token_expiry, gdrive_client_id
		FROM settings
		WHERE id = 1`).Scan(
		&s.ActiveLLM,
		&s.OllamaEndpoint,
		&s.OllamaModel,
		&s.CustomSystemPrompt,
		&s.CloudProvider,
		&s.CloudEndpoint,
		&s.CloudBucket,
		&s.CloudAccessKey,
		&s.CloudSecretKey,
		&s.GDriveAccessToken,
		&s.GDriveRefreshToken,
		&gdriveExpiry,
		&s.GDriveClientID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get settings: %w", err)
	}
	if gdriveExpiry.Valid {
		s.GDriveTokenExpiry = gdriveExpiry.Time
	}
	return s, nil
}
