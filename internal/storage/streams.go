package storage

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

type ActivityStream struct {
	ID         int64
	ActivityID int64
	StreamType string
	Data       json.RawMessage
	CreatedAt  time.Time
}

func (db *DB) SaveActivityStream(activityID int64, streamType string, data json.RawMessage) error {
	if streamType == "" {
		return errors.New("stream type must not be empty")
	}
	if len(data) == 0 {
		return errors.New("stream data must not be empty")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(
		`INSERT INTO activity_streams (activity_id, stream_type, data) VALUES (?, ?, ?)`,
		activityID, streamType, string(data),
	)
	if err != nil {
		return fmt.Errorf("save activity stream: %w", err)
	}
	return nil
}

func (db *DB) GetActivityStreams(activityID int64) ([]ActivityStream, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(
		`SELECT id, activity_id, stream_type, data, created_at
		 FROM activity_streams
		 WHERE activity_id = ?
		 ORDER BY stream_type ASC`, activityID,
	)
	if err != nil {
		return nil, fmt.Errorf("query activity streams: %w", err)
	}
	defer rows.Close()

	var streams []ActivityStream
	for rows.Next() {
		var s ActivityStream
		var dataStr string
		if err := rows.Scan(&s.ID, &s.ActivityID, &s.StreamType, &dataStr, &s.CreatedAt); err != nil {
			return nil, fmt.Errorf("scan activity stream: %w", err)
		}
		s.Data = json.RawMessage(dataStr)
		streams = append(streams, s)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate activity streams: %w", err)
	}
	return streams, nil
}

func (db *DB) GetActivityStreamByType(activityID int64, streamType string) (*ActivityStream, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var s ActivityStream
	var dataStr string
	err := db.conn.QueryRow(
		`SELECT id, activity_id, stream_type, data, created_at
		 FROM activity_streams
		 WHERE activity_id = ? AND stream_type = ?`,
		activityID, streamType,
	).Scan(&s.ID, &s.ActivityID, &s.StreamType, &dataStr, &s.CreatedAt)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get activity stream by type: %w", err)
	}
	s.Data = json.RawMessage(dataStr)
	return &s, nil
}

func (db *DB) DeleteActivityStreams(activityID int64) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`DELETE FROM activity_streams WHERE activity_id = ?`, activityID)
	if err != nil {
		return fmt.Errorf("delete activity streams: %w", err)
	}
	return nil
}
