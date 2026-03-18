package cloudsync

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"
)

type mockProvider struct {
	mu        sync.Mutex
	uploads   map[string][]byte
	modTimes  map[string]time.Time
	uploadErr error
	dlErr     error
}

func newMockProvider() *mockProvider {
	return &mockProvider{
		uploads:  make(map[string][]byte),
		modTimes: make(map[string]time.Time),
	}
}

func (m *mockProvider) Name() string { return "mock" }

func (m *mockProvider) Upload(_ context.Context, key string, data []byte) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.uploadErr != nil {
		return m.uploadErr
	}
	m.uploads[key] = data
	m.modTimes[key] = time.Now()
	return nil
}

func (m *mockProvider) Download(_ context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.dlErr != nil {
		return nil, m.dlErr
	}
	data, ok := m.uploads[key]
	if !ok {
		return nil, ErrNotFound
	}
	return data, nil
}

func (m *mockProvider) LastModified(_ context.Context, key string) (time.Time, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	t, ok := m.modTimes[key]
	if !ok {
		return time.Time{}, ErrNotFound
	}
	return t, nil
}

type mockStateStore struct {
	mu             sync.Mutex
	lastSynced     time.Time
	lastChatSynced time.Time
}

func (s *mockStateStore) GetLastSyncedAt() (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastSynced, nil
}

func (s *mockStateStore) SetLastSyncedAt(t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastSynced = t
	return nil
}

func (s *mockStateStore) GetLastChatSyncAt() (time.Time, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.lastChatSynced, nil
}

func (s *mockStateStore) SetLastChatSyncAt(t time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.lastChatSynced = t
	return nil
}

func newTestManager(t *testing.T) (*Manager, *mockProvider, *mockStateStore) {
	t.Helper()
	prov := newMockProvider()
	store := &mockStateStore{}
	mgr, err := NewManager(ManagerConfig{
		Provider:      prov,
		ExportContext: func() ([]byte, error) { return []byte(`{"context":"data"}`), nil },
		ImportContext: func(data []byte, replaceAll bool) error { return nil },
		ExportChat:    func() ([]byte, error) { return []byte(`{"chat":"data"}`), nil },
		ImportChat:    func(data []byte, replaceAll bool) error { return nil },
		StateStore:    store,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return mgr, prov, store
}

func TestNewManager_RequiresProvider(t *testing.T) {
	_, err := NewManager(ManagerConfig{
		ExportContext: func() ([]byte, error) { return nil, nil },
		ImportContext: func([]byte, bool) error { return nil },
		StateStore:    &mockStateStore{},
	})
	if err == nil {
		t.Error("expected error for nil provider")
	}
}

func TestNewManager_RequiresExportContext(t *testing.T) {
	_, err := NewManager(ManagerConfig{
		Provider:      newMockProvider(),
		ImportContext: func([]byte, bool) error { return nil },
		StateStore:    &mockStateStore{},
	})
	if err == nil {
		t.Error("expected error for nil export context")
	}
}

func TestNewManager_RequiresImportContext(t *testing.T) {
	_, err := NewManager(ManagerConfig{
		Provider:      newMockProvider(),
		ExportContext: func() ([]byte, error) { return nil, nil },
		StateStore:    &mockStateStore{},
	})
	if err == nil {
		t.Error("expected error for nil import context")
	}
}

func TestNewManager_RequiresStateStore(t *testing.T) {
	_, err := NewManager(ManagerConfig{
		Provider:      newMockProvider(),
		ExportContext: func() ([]byte, error) { return nil, nil },
		ImportContext: func([]byte, bool) error { return nil },
	})
	if err == nil {
		t.Error("expected error for nil state store")
	}
}

func TestSyncNow_UploadsContextAndChat(t *testing.T) {
	mgr, prov, store := newTestManager(t)
	defer mgr.Stop()

	if err := mgr.SyncNow(); err != nil {
		t.Fatalf("SyncNow: %v", err)
	}

	if _, ok := prov.uploads[ContextKey]; !ok {
		t.Error("context not uploaded")
	}
	if _, ok := prov.uploads[ChatKey]; !ok {
		t.Error("chat not uploaded")
	}

	lastSynced, _ := store.GetLastSyncedAt()
	if lastSynced.IsZero() {
		t.Error("last synced at not set")
	}

	lastChat, _ := store.GetLastChatSyncAt()
	if lastChat.IsZero() {
		t.Error("last chat sync at not set")
	}
}

func TestSyncNow_UploadError(t *testing.T) {
	mgr, prov, _ := newTestManager(t)
	defer mgr.Stop()

	prov.uploadErr = errors.New("network error")

	err := mgr.SyncNow()
	if err == nil {
		t.Error("expected error on upload failure")
	}

	status := mgr.GetStatus()
	if status.LastError == "" {
		t.Error("expected LastError to be set")
	}
}

func TestGetStatus_ReflectsManagerState(t *testing.T) {
	mgr, _, _ := newTestManager(t)
	defer mgr.Stop()

	status := mgr.GetStatus()
	if !status.Enabled {
		t.Error("expected enabled")
	}
	if status.Provider != "mock" {
		t.Errorf("Provider = %q, want %q", status.Provider, "mock")
	}
	if status.Syncing {
		t.Error("should not be syncing initially")
	}
	if status.LastSyncedAt != "" {
		t.Error("expected empty last synced")
	}
}

func TestGetStatus_AfterSync(t *testing.T) {
	mgr, _, _ := newTestManager(t)
	defer mgr.Stop()

	_ = mgr.SyncNow()

	status := mgr.GetStatus()
	if status.LastSyncedAt == "" {
		t.Error("expected non-empty last synced after sync")
	}
	if status.LastChatSyncAt == "" {
		t.Error("expected non-empty last chat sync after sync")
	}
	if status.LastError != "" {
		t.Errorf("unexpected error: %s", status.LastError)
	}
}

func TestStop_PreventsSubsequentSync(t *testing.T) {
	mgr, prov, _ := newTestManager(t)

	mgr.Stop()

	_ = mgr.SyncNow()

	if len(prov.uploads) != 0 {
		t.Error("sync should not execute after Stop()")
	}
}

func TestStop_Idempotent(t *testing.T) {
	mgr, _, _ := newTestManager(t)
	mgr.Stop()
	mgr.Stop()
}

func TestSingleFlight(t *testing.T) {
	prov := newMockProvider()
	store := &mockStateStore{}
	uploadCount := 0
	var mu sync.Mutex

	slowExport := func() ([]byte, error) {
		time.Sleep(50 * time.Millisecond)
		mu.Lock()
		uploadCount++
		mu.Unlock()
		return []byte("data"), nil
	}

	mgr, err := NewManager(ManagerConfig{
		Provider:      prov,
		ExportContext: slowExport,
		ImportContext: func([]byte, bool) error { return nil },
		StateStore:    store,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer mgr.Stop()

	var wg sync.WaitGroup
	for i := 0; i < 5; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = mgr.SyncNow()
		}()
	}
	wg.Wait()

	mu.Lock()
	count := uploadCount
	mu.Unlock()

	if count >= 5 {
		t.Errorf("expected fewer than 5 exports due to single-flight, got %d", count)
	}
}

func TestCheckRemote_DetectsNewerRemote(t *testing.T) {
	mgr, prov, store := newTestManager(t)
	defer mgr.Stop()

	past := time.Now().Add(-1 * time.Hour)
	future := time.Now().Add(1 * time.Hour)

	_ = store.SetLastSyncedAt(past)
	_ = store.SetLastChatSyncAt(past)

	prov.mu.Lock()
	prov.modTimes[ContextKey] = future
	prov.modTimes[ChatKey] = future
	prov.mu.Unlock()

	status, err := mgr.CheckRemote(context.Background())
	if err != nil {
		t.Fatalf("CheckRemote: %v", err)
	}
	if !status.ContextNewer {
		t.Error("expected ContextNewer = true")
	}
	if !status.ChatNewer {
		t.Error("expected ChatNewer = true")
	}
}

func TestCheckRemote_NoRemoteFile(t *testing.T) {
	mgr, _, _ := newTestManager(t)
	defer mgr.Stop()

	status, err := mgr.CheckRemote(context.Background())
	if err != nil {
		t.Fatalf("CheckRemote: %v", err)
	}
	if status.ContextNewer || status.ChatNewer {
		t.Error("expected no newer remote when files don't exist")
	}
}

func TestPullRemote_ImportsNewerData(t *testing.T) {
	var imported bool
	prov := newMockProvider()
	store := &mockStateStore{}

	mgr, err := NewManager(ManagerConfig{
		Provider:      prov,
		ExportContext: func() ([]byte, error) { return []byte("ctx"), nil },
		ImportContext: func(data []byte, replaceAll bool) error {
			imported = true
			if !replaceAll {
				t.Error("expected replaceAll=true for context pull")
			}
			return nil
		},
		ImportChat: func(data []byte, replaceAll bool) error { return nil },
		ExportChat: func() ([]byte, error) { return []byte("chat"), nil },
		StateStore: store,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer mgr.Stop()

	past := time.Now().Add(-1 * time.Hour)
	_ = store.SetLastSyncedAt(past)

	prov.mu.Lock()
	prov.uploads[ContextKey] = []byte("remote context data")
	prov.modTimes[ContextKey] = time.Now()
	prov.mu.Unlock()

	if err := mgr.PullRemote(context.Background()); err != nil {
		t.Fatalf("PullRemote: %v", err)
	}
	if !imported {
		t.Error("expected context to be imported")
	}
}

func TestTriggerChatSync_NilExportChat(t *testing.T) {
	prov := newMockProvider()
	store := &mockStateStore{}

	mgr, err := NewManager(ManagerConfig{
		Provider:      prov,
		ExportContext: func() ([]byte, error) { return []byte("ctx"), nil },
		ImportContext: func([]byte, bool) error { return nil },
		StateStore:    store,
	})
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	defer mgr.Stop()

	mgr.TriggerChatSync()
}
