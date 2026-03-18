package cloudsync

import (
	"time"

	"coachlm/internal/storage"
)

type StateAdapter struct {
	db *storage.DB
}

func NewStateAdapter(db *storage.DB) *StateAdapter {
	return &StateAdapter{db: db}
}

func (a *StateAdapter) GetLastSyncedAt() (time.Time, error) {
	state, err := a.db.GetCloudSyncState()
	if err != nil {
		return time.Time{}, err
	}
	if state == nil {
		return time.Time{}, nil
	}
	return state.LastSyncedAt, nil
}

func (a *StateAdapter) SetLastSyncedAt(t time.Time) error {
	state, err := a.db.GetCloudSyncState()
	if err != nil {
		return err
	}
	if state == nil {
		state = &storage.CloudSyncState{}
	}
	state.LastSyncedAt = t
	return a.db.SaveCloudSyncState(state)
}

func (a *StateAdapter) GetLastChatSyncAt() (time.Time, error) {
	state, err := a.db.GetCloudSyncState()
	if err != nil {
		return time.Time{}, err
	}
	if state == nil {
		return time.Time{}, nil
	}
	return state.LastChatSyncAt, nil
}

func (a *StateAdapter) SetLastChatSyncAt(t time.Time) error {
	state, err := a.db.GetCloudSyncState()
	if err != nil {
		return err
	}
	if state == nil {
		state = &storage.CloudSyncState{}
	}
	state.LastChatSyncAt = t
	return a.db.SaveCloudSyncState(state)
}
