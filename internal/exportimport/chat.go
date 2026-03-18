package exportimport

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"coachlm/internal/storage"
)

const chatSchemaVersion = 1

type ChatExportEnvelope struct {
	Version    string              `json:"version"`
	ExportedAt string              `json:"exported_at"`
	Sessions   []ChatSessionExport `json:"sessions"`
}

type ChatSessionExport struct {
	ID        string              `json:"id"`
	CreatedAt string              `json:"created_at"`
	Messages  []ChatMessageExport `json:"messages"`
}

type ChatMessageExport struct {
	Role      string `json:"role"`
	Content   string `json:"content"`
	Timestamp string `json:"timestamp"`
}

func ExportChatData(db *storage.DB) ([]byte, error) {
	if db == nil {
		return nil, errors.New("database is nil")
	}

	sessions, err := db.ListSessions()
	if err != nil {
		return nil, fmt.Errorf("list sessions: %w", err)
	}

	exported := make([]ChatSessionExport, 0, len(sessions))
	for _, sess := range sessions {
		messages, err := db.GetMessages(sess.ID)
		if err != nil {
			return nil, fmt.Errorf("get messages for session %s: %w", sess.ID, err)
		}

		msgs := make([]ChatMessageExport, 0, len(messages))
		for _, m := range messages {
			msgs = append(msgs, ChatMessageExport{
				Role:      m.Role,
				Content:   m.Content,
				Timestamp: m.CreatedAt.Format(time.RFC3339),
			})
		}

		exported = append(exported, ChatSessionExport{
			ID:        sess.ID,
			CreatedAt: sess.CreatedAt.Format(time.RFC3339),
			Messages:  msgs,
		})
	}

	envelope := ChatExportEnvelope{
		Version:    fmt.Sprintf("%d", chatSchemaVersion),
		ExportedAt: time.Now().Format(time.RFC3339),
		Sessions:   exported,
	}

	data, err := json.MarshalIndent(envelope, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("marshal chat export: %w", err)
	}

	return data, nil
}

func ImportChatData(db *storage.DB, data []byte, replaceAll bool) error {
	if db == nil {
		return errors.New("database is nil")
	}

	var envelope ChatExportEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		return fmt.Errorf("unmarshal chat import: %w", err)
	}

	if replaceAll {
		existing, err := db.ListSessions()
		if err != nil {
			return fmt.Errorf("list existing sessions: %w", err)
		}
		for _, sess := range existing {
			_ = db.DeleteSession(sess.ID)
		}
	}

	for _, sess := range envelope.Sessions {
		existing, err := db.GetSession(sess.ID)
		if err != nil {
			return fmt.Errorf("check session %s: %w", sess.ID, err)
		}

		if existing != nil {
			if replaceAll {
				_ = db.DeleteSession(sess.ID)
			} else {
				remoteParsed, err := time.Parse(time.RFC3339, sess.CreatedAt)
				if err != nil {
					continue
				}
				if !remoteParsed.After(existing.CreatedAt) {
					continue
				}
				_ = db.DeleteSession(sess.ID)
			}
		}

		if _, err := db.CreateSession(sess.ID); err != nil {
			return fmt.Errorf("create session %s: %w", sess.ID, err)
		}

		for _, msg := range sess.Messages {
			if _, err := db.SaveMessage(sess.ID, msg.Role, msg.Content); err != nil {
				return fmt.Errorf("save message in session %s: %w", sess.ID, err)
			}
		}
	}

	return nil
}
