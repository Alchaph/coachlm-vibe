package exportimport

import (
	"encoding/json"
	"testing"
	"time"

	"coachlm/internal/storage"
)

func TestExportChatData_EmptySessions(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	data, err := ExportChatData(db)
	if err != nil {
		t.Fatalf("ExportChatData: %v", err)
	}

	var envelope ChatExportEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if envelope.Version != "1" {
		t.Errorf("Version = %q, want %q", envelope.Version, "1")
	}
	if len(envelope.Sessions) != 0 {
		t.Errorf("Sessions = %d, want 0", len(envelope.Sessions))
	}
	if envelope.ExportedAt == "" {
		t.Error("ExportedAt should be set")
	}
}

func TestExportChatData_WithMessages(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	sess, err := db.CreateSession("sess-1")
	if err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	_ = sess

	if _, err := db.SaveMessage("sess-1", "user", "Hello"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	if _, err := db.SaveMessage("sess-1", "assistant", "Hi there"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	data, err := ExportChatData(db)
	if err != nil {
		t.Fatalf("ExportChatData: %v", err)
	}

	var envelope ChatExportEnvelope
	if err := json.Unmarshal(data, &envelope); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(envelope.Sessions) != 1 {
		t.Fatalf("Sessions = %d, want 1", len(envelope.Sessions))
	}
	if envelope.Sessions[0].ID != "sess-1" {
		t.Errorf("Session ID = %q, want %q", envelope.Sessions[0].ID, "sess-1")
	}
	if len(envelope.Sessions[0].Messages) != 2 {
		t.Errorf("Messages = %d, want 2", len(envelope.Sessions[0].Messages))
	}
	if envelope.Sessions[0].Messages[0].Role != "user" {
		t.Errorf("Message[0].Role = %q, want %q", envelope.Sessions[0].Messages[0].Role, "user")
	}
	if envelope.Sessions[0].Messages[0].Content != "Hello" {
		t.Errorf("Message[0].Content = %q, want %q", envelope.Sessions[0].Messages[0].Content, "Hello")
	}
}

func TestExportChatData_NilDB(t *testing.T) {
	_, err := ExportChatData(nil)
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

func TestImportChatData_NilDB(t *testing.T) {
	err := ImportChatData(nil, []byte(`{"version":"1","sessions":[]}`), false)
	if err == nil {
		t.Error("expected error for nil DB")
	}
}

func TestImportChatData_InvalidJSON(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	err := ImportChatData(db, []byte("not json"), false)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestImportChatData_MergeSessions(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	if _, err := db.CreateSession("sess-1"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := db.SaveMessage("sess-1", "user", "local message"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	importData := ChatExportEnvelope{
		Version:    "1",
		ExportedAt: time.Now().Format(time.RFC3339),
		Sessions: []ChatSessionExport{
			{
				ID:        "sess-2",
				CreatedAt: time.Now().Format(time.RFC3339),
				Messages: []ChatMessageExport{
					{Role: "user", Content: "remote message", Timestamp: time.Now().Format(time.RFC3339)},
				},
			},
		},
	}
	data, _ := json.Marshal(importData)

	if err := ImportChatData(db, data, false); err != nil {
		t.Fatalf("ImportChatData: %v", err)
	}

	sessions, err := db.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("Sessions = %d, want 2 (merge)", len(sessions))
	}
}

func TestImportChatData_ReplaceAll(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	if _, err := db.CreateSession("sess-old"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := db.SaveMessage("sess-old", "user", "old"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	importData := ChatExportEnvelope{
		Version:    "1",
		ExportedAt: time.Now().Format(time.RFC3339),
		Sessions: []ChatSessionExport{
			{
				ID:        "sess-new",
				CreatedAt: time.Now().Format(time.RFC3339),
				Messages: []ChatMessageExport{
					{Role: "user", Content: "new", Timestamp: time.Now().Format(time.RFC3339)},
				},
			},
		},
	}
	data, _ := json.Marshal(importData)

	if err := ImportChatData(db, data, true); err != nil {
		t.Fatalf("ImportChatData replace: %v", err)
	}

	sessions, err := db.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 1 {
		t.Fatalf("Sessions = %d, want 1 (replaced)", len(sessions))
	}
	if sessions[0].ID != "sess-new" {
		t.Errorf("Session ID = %q, want %q", sessions[0].ID, "sess-new")
	}
}

func TestImportChatData_DuplicateSessionSkipped(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	if _, err := db.CreateSession("sess-1"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := db.SaveMessage("sess-1", "user", "original"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	existingSession, _ := db.GetSession("sess-1")

	importData := ChatExportEnvelope{
		Version:    "1",
		ExportedAt: time.Now().Format(time.RFC3339),
		Sessions: []ChatSessionExport{
			{
				ID:        "sess-1",
				CreatedAt: existingSession.CreatedAt.Add(-1 * time.Hour).Format(time.RFC3339),
				Messages: []ChatMessageExport{
					{Role: "user", Content: "older version", Timestamp: time.Now().Format(time.RFC3339)},
				},
			},
		},
	}
	data, _ := json.Marshal(importData)

	if err := ImportChatData(db, data, false); err != nil {
		t.Fatalf("ImportChatData: %v", err)
	}

	messages, _ := db.GetMessages("sess-1")
	if len(messages) != 1 {
		t.Fatalf("Messages = %d, want 1", len(messages))
	}
	if messages[0].Content != "original" {
		t.Errorf("Content = %q, want %q (should keep local)", messages[0].Content, "original")
	}
}

func TestImportChatData_NewerRemoteReplacesLocal(t *testing.T) {
	db := newTestDB(t)
	defer db.Close()

	if _, err := db.CreateSession("sess-1"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := db.SaveMessage("sess-1", "user", "original"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	importData := ChatExportEnvelope{
		Version:    "1",
		ExportedAt: time.Now().Format(time.RFC3339),
		Sessions: []ChatSessionExport{
			{
				ID:        "sess-1",
				CreatedAt: time.Now().Add(1 * time.Hour).Format(time.RFC3339),
				Messages: []ChatMessageExport{
					{Role: "user", Content: "newer version", Timestamp: time.Now().Format(time.RFC3339)},
				},
			},
		},
	}
	data, _ := json.Marshal(importData)

	if err := ImportChatData(db, data, false); err != nil {
		t.Fatalf("ImportChatData: %v", err)
	}

	messages, _ := db.GetMessages("sess-1")
	if len(messages) != 1 {
		t.Fatalf("Messages = %d, want 1", len(messages))
	}
	if messages[0].Content != "newer version" {
		t.Errorf("Content = %q, want %q (should accept newer)", messages[0].Content, "newer version")
	}
}

func TestExportImportChatData_RoundTrip(t *testing.T) {
	db1 := newTestDB(t)
	defer db1.Close()

	if _, err := db1.CreateSession("rt-1"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := db1.SaveMessage("rt-1", "user", "question"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	if _, err := db1.SaveMessage("rt-1", "assistant", "answer"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}
	if _, err := db1.CreateSession("rt-2"); err != nil {
		t.Fatalf("CreateSession: %v", err)
	}
	if _, err := db1.SaveMessage("rt-2", "user", "another question"); err != nil {
		t.Fatalf("SaveMessage: %v", err)
	}

	data, err := ExportChatData(db1)
	if err != nil {
		t.Fatalf("ExportChatData: %v", err)
	}

	db2 := newTestDB(t)
	defer db2.Close()

	if err := ImportChatData(db2, data, false); err != nil {
		t.Fatalf("ImportChatData: %v", err)
	}

	sessions, err := db2.ListSessions()
	if err != nil {
		t.Fatalf("ListSessions: %v", err)
	}
	if len(sessions) != 2 {
		t.Errorf("Sessions = %d, want 2", len(sessions))
	}

	msgs1, _ := db2.GetMessages("rt-1")
	if len(msgs1) != 2 {
		t.Errorf("rt-1 messages = %d, want 2", len(msgs1))
	}
	msgs2, _ := db2.GetMessages("rt-2")
	if len(msgs2) != 1 {
		t.Errorf("rt-2 messages = %d, want 1", len(msgs2))
	}
}

func TestChatExportEnvelope_ValidJSON(t *testing.T) {
	envelope := ChatExportEnvelope{
		Version:    "1",
		ExportedAt: time.Now().Format(time.RFC3339),
		Sessions:   []ChatSessionExport{},
	}

	data, err := json.Marshal(envelope)
	if err != nil {
		t.Fatalf("Marshal: %v", err)
	}

	var decoded ChatExportEnvelope
	if err := json.Unmarshal(data, &decoded); err != nil {
		t.Fatalf("Unmarshal: %v", err)
	}
	if decoded.Version != "1" {
		t.Errorf("Version = %q, want %q", decoded.Version, "1")
	}

	m := make(map[string]interface{})
	_ = json.Unmarshal(data, &m)
	if _, ok := m["version"]; !ok {
		t.Error("expected 'version' key in JSON")
	}
	if _, ok := m["exported_at"]; !ok {
		t.Error("expected 'exported_at' key in JSON")
	}
	if _, ok := m["sessions"]; !ok {
		t.Error("expected 'sessions' key in JSON")
	}
}

func newChatTestDB(t *testing.T) *storage.DB {
	t.Helper()
	db, err := storage.New(":memory:")
	if err != nil {
		t.Fatalf("create test DB: %v", err)
	}
	return db
}
