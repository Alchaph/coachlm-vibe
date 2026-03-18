package storage

import (
	"testing"
	"time"
)

func TestCloudSyncState_CRUD(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()

	got, err := db.GetCloudSyncState()
	if err != nil {
		t.Fatalf("GetCloudSyncState (empty): %v", err)
	}
	if got != nil {
		t.Errorf("expected nil for empty table, got %+v", got)
	}

	now := time.Now().Truncate(time.Second)
	state := &CloudSyncState{
		Provider:       "s3",
		LastSyncedAt:   now,
		LastChatSyncAt: now.Add(-1 * time.Hour),
		RemoteEtag:     "etag-123",
	}
	if err := db.SaveCloudSyncState(state); err != nil {
		t.Fatalf("SaveCloudSyncState: %v", err)
	}

	loaded, err := db.GetCloudSyncState()
	if err != nil {
		t.Fatalf("GetCloudSyncState: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected non-nil state")
	}
	if loaded.Provider != "s3" {
		t.Errorf("Provider = %q, want %q", loaded.Provider, "s3")
	}
	if loaded.RemoteEtag != "etag-123" {
		t.Errorf("RemoteEtag = %q, want %q", loaded.RemoteEtag, "etag-123")
	}
	if loaded.LastSyncedAt.IsZero() {
		t.Error("LastSyncedAt should not be zero")
	}
	if loaded.LastChatSyncAt.IsZero() {
		t.Error("LastChatSyncAt should not be zero")
	}
}

func TestCloudSyncState_Update(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()

	state := &CloudSyncState{
		Provider:     "s3",
		LastSyncedAt: time.Now().Truncate(time.Second),
		RemoteEtag:   "v1",
	}
	if err := db.SaveCloudSyncState(state); err != nil {
		t.Fatalf("Save initial: %v", err)
	}

	state.RemoteEtag = "v2"
	state.LastChatSyncAt = time.Now().Truncate(time.Second)
	if err := db.SaveCloudSyncState(state); err != nil {
		t.Fatalf("Save update: %v", err)
	}

	loaded, err := db.GetCloudSyncState()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if loaded.RemoteEtag != "v2" {
		t.Errorf("RemoteEtag = %q, want %q", loaded.RemoteEtag, "v2")
	}
	if loaded.LastChatSyncAt.IsZero() {
		t.Error("LastChatSyncAt should be set after update")
	}
}

func TestCloudSyncState_Delete(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()

	state := &CloudSyncState{
		Provider:     "s3",
		LastSyncedAt: time.Now(),
		RemoteEtag:   "etag",
	}
	if err := db.SaveCloudSyncState(state); err != nil {
		t.Fatalf("Save: %v", err)
	}

	if err := db.DeleteCloudSyncState(); err != nil {
		t.Fatalf("Delete: %v", err)
	}

	got, err := db.GetCloudSyncState()
	if err != nil {
		t.Fatalf("Get after delete: %v", err)
	}
	if got != nil {
		t.Error("expected nil after delete")
	}
}

func TestCloudSyncState_SaveNilReturnsError(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()

	if err := db.SaveCloudSyncState(nil); err == nil {
		t.Error("expected error for nil state")
	}
}

func TestCloudSyncState_ZeroTimesStoredAsNull(t *testing.T) {
	db, err := New(":memory:")
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	defer db.Close()

	state := &CloudSyncState{
		Provider:   "gdrive",
		RemoteEtag: "e1",
	}
	if err := db.SaveCloudSyncState(state); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := db.GetCloudSyncState()
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if !loaded.LastSyncedAt.IsZero() {
		t.Errorf("LastSyncedAt should be zero, got %v", loaded.LastSyncedAt)
	}
	if !loaded.LastChatSyncAt.IsZero() {
		t.Errorf("LastChatSyncAt should be zero, got %v", loaded.LastChatSyncAt)
	}
}
