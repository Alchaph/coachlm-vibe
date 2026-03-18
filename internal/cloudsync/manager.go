package cloudsync

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"
)

const (
	ContextKey     = "coachlm/context.coachctx"
	ChatKey        = "coachlm/chat_sessions.coachctx"
	DebouncePeriod = 30 * time.Second
)

type SyncStatus struct {
	Enabled        bool   `json:"enabled"`
	Provider       string `json:"provider"`
	LastSyncedAt   string `json:"lastSyncedAt"`
	LastChatSyncAt string `json:"lastChatSyncAt"`
	Syncing        bool   `json:"syncing"`
	LastError      string `json:"lastError"`
}

type DataExporter func() ([]byte, error)

type DataImporter func(data []byte, replaceAll bool) error

type SyncStateStore interface {
	GetLastSyncedAt() (time.Time, error)
	SetLastSyncedAt(t time.Time) error
	GetLastChatSyncAt() (time.Time, error)
	SetLastChatSyncAt(t time.Time) error
}

type Manager struct {
	provider CloudProvider

	exportContext DataExporter
	importContext DataImporter
	exportChat    DataExporter
	importChat    DataImporter
	stateStore    SyncStateStore

	mu      sync.Mutex
	syncing bool
	lastErr error
	stopCh  chan struct{}
	stopped bool

	debounceMu    sync.Mutex
	debounceTimer *time.Timer

	chatDebounceMu    sync.Mutex
	chatDebounceTimer *time.Timer
}

type ManagerConfig struct {
	Provider      CloudProvider
	ExportContext DataExporter
	ImportContext DataImporter
	ExportChat    DataExporter
	ImportChat    DataImporter
	StateStore    SyncStateStore
}

func NewManager(cfg ManagerConfig) (*Manager, error) {
	if cfg.Provider == nil {
		return nil, errors.New("cloudsync: provider is required")
	}
	if cfg.ExportContext == nil {
		return nil, errors.New("cloudsync: export context function is required")
	}
	if cfg.ImportContext == nil {
		return nil, errors.New("cloudsync: import context function is required")
	}
	if cfg.StateStore == nil {
		return nil, errors.New("cloudsync: state store is required")
	}

	return &Manager{
		provider:      cfg.Provider,
		exportContext: cfg.ExportContext,
		importContext: cfg.ImportContext,
		exportChat:    cfg.ExportChat,
		importChat:    cfg.ImportChat,
		stateStore:    cfg.StateStore,
		stopCh:        make(chan struct{}),
	}, nil
}

func (m *Manager) TriggerSync() {
	m.debounceMu.Lock()
	defer m.debounceMu.Unlock()

	if m.debounceTimer != nil {
		m.debounceTimer.Stop()
	}
	m.debounceTimer = time.AfterFunc(DebouncePeriod, func() {
		_ = m.syncContext(context.Background())
	})
}

func (m *Manager) TriggerChatSync() {
	if m.exportChat == nil || m.importChat == nil {
		return
	}

	m.chatDebounceMu.Lock()
	defer m.chatDebounceMu.Unlock()

	if m.chatDebounceTimer != nil {
		m.chatDebounceTimer.Stop()
	}
	m.chatDebounceTimer = time.AfterFunc(DebouncePeriod, func() {
		_ = m.syncChat(context.Background())
	})
}

func (m *Manager) SyncNow() error {
	ctxErr := m.syncContext(context.Background())
	chatErr := m.syncChat(context.Background())
	if ctxErr != nil {
		return ctxErr
	}
	return chatErr
}

func (m *Manager) CheckRemote(ctx context.Context) (*RemoteStatus, error) {
	localCtxSync, _ := m.stateStore.GetLastSyncedAt()
	localChatSync, _ := m.stateStore.GetLastChatSyncAt()

	status := &RemoteStatus{}

	remoteCtxTime, err := m.provider.LastModified(ctx, ContextKey)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return nil, fmt.Errorf("check remote context: %w", err)
	}
	if err == nil && remoteCtxTime.After(localCtxSync) {
		status.ContextNewer = true
		status.ContextRemoteTime = remoteCtxTime
	}

	if m.exportChat != nil {
		remoteChatTime, err := m.provider.LastModified(ctx, ChatKey)
		if err != nil && !errors.Is(err, ErrNotFound) {
			return nil, fmt.Errorf("check remote chat: %w", err)
		}
		if err == nil && remoteChatTime.After(localChatSync) {
			status.ChatNewer = true
			status.ChatRemoteTime = remoteChatTime
		}
	}

	return status, nil
}

func (m *Manager) PullRemote(ctx context.Context) error {
	status, err := m.CheckRemote(ctx)
	if err != nil {
		return err
	}

	if status.ContextNewer {
		data, err := m.provider.Download(ctx, ContextKey)
		if err != nil {
			return fmt.Errorf("pull context: %w", err)
		}
		if err := m.importContext(data, true); err != nil {
			return fmt.Errorf("import context: %w", err)
		}
		if err := m.stateStore.SetLastSyncedAt(status.ContextRemoteTime); err != nil {
			return fmt.Errorf("update context sync time: %w", err)
		}
	}

	if status.ChatNewer && m.importChat != nil {
		data, err := m.provider.Download(ctx, ChatKey)
		if err != nil {
			return fmt.Errorf("pull chat: %w", err)
		}
		if err := m.importChat(data, false); err != nil {
			return fmt.Errorf("import chat: %w", err)
		}
		if err := m.stateStore.SetLastChatSyncAt(status.ChatRemoteTime); err != nil {
			return fmt.Errorf("update chat sync time: %w", err)
		}
	}

	return nil
}

func (m *Manager) GetStatus() SyncStatus {
	m.mu.Lock()
	defer m.mu.Unlock()

	lastSynced, _ := m.stateStore.GetLastSyncedAt()
	lastChat, _ := m.stateStore.GetLastChatSyncAt()

	s := SyncStatus{
		Enabled:  true,
		Provider: m.provider.Name(),
		Syncing:  m.syncing,
	}
	if !lastSynced.IsZero() {
		s.LastSyncedAt = lastSynced.Format(time.RFC3339)
	}
	if !lastChat.IsZero() {
		s.LastChatSyncAt = lastChat.Format(time.RFC3339)
	}
	if m.lastErr != nil {
		s.LastError = m.lastErr.Error()
	}
	return s
}

func (m *Manager) Stop() {
	m.mu.Lock()
	if m.stopped {
		m.mu.Unlock()
		return
	}
	m.stopped = true
	close(m.stopCh)
	m.mu.Unlock()

	m.debounceMu.Lock()
	if m.debounceTimer != nil {
		m.debounceTimer.Stop()
	}
	m.debounceMu.Unlock()

	m.chatDebounceMu.Lock()
	if m.chatDebounceTimer != nil {
		m.chatDebounceTimer.Stop()
	}
	m.chatDebounceMu.Unlock()
}

func (m *Manager) syncContext(ctx context.Context) error {
	if !m.tryAcquire() {
		return nil
	}
	defer m.release()

	data, err := m.exportContext()
	if err != nil {
		m.setError(fmt.Errorf("export context: %w", err))
		return err
	}

	if err := m.provider.Upload(ctx, ContextKey, data); err != nil {
		m.setError(fmt.Errorf("upload context: %w", err))
		return err
	}

	now := time.Now()
	if err := m.stateStore.SetLastSyncedAt(now); err != nil {
		m.setError(fmt.Errorf("save sync state: %w", err))
		return err
	}

	m.setError(nil)
	return nil
}

func (m *Manager) syncChat(ctx context.Context) error {
	if m.exportChat == nil {
		return nil
	}

	if !m.tryAcquire() {
		return nil
	}
	defer m.release()

	data, err := m.exportChat()
	if err != nil {
		m.setError(fmt.Errorf("export chat: %w", err))
		return err
	}

	if err := m.provider.Upload(ctx, ChatKey, data); err != nil {
		m.setError(fmt.Errorf("upload chat: %w", err))
		return err
	}

	now := time.Now()
	if err := m.stateStore.SetLastChatSyncAt(now); err != nil {
		m.setError(fmt.Errorf("save chat sync state: %w", err))
		return err
	}

	m.setError(nil)
	return nil
}

func (m *Manager) tryAcquire() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.syncing || m.stopped {
		return false
	}
	m.syncing = true
	return true
}

func (m *Manager) release() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.syncing = false
}

func (m *Manager) setError(err error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.lastErr = err
}

type RemoteStatus struct {
	ContextNewer      bool
	ContextRemoteTime time.Time
	ChatNewer         bool
	ChatRemoteTime    time.Time
}
