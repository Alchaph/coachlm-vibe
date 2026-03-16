package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type Activity struct {
	ID           int64
	StravaID     int64
	Name         string
	ActivityType string
	StartDate    time.Time
	Distance     float64
	DurationSecs int
	AvgPaceSecs  int
	AvgHR        int
	MaxHR        int
	AvgCadence   float64
	Source       string
	CreatedAt    time.Time
}

// SaveActivity uses INSERT OR IGNORE to deduplicate on strava_id.
func (db *DB) SaveActivity(activity *Activity) error {
	if activity == nil {
		return errors.New("activity is nil")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`
		INSERT OR IGNORE INTO activities
			(strava_id, name, activity_type, start_date, distance, duration_secs, avg_pace_secs, avg_hr, max_hr, avg_cadence, source)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		activity.StravaID,
		activity.Name,
		activity.ActivityType,
		activity.StartDate,
		activity.Distance,
		activity.DurationSecs,
		activity.AvgPaceSecs,
		activity.AvgHR,
		activity.MaxHR,
		activity.AvgCadence,
		activity.Source,
	)
	if err != nil {
		return fmt.Errorf("save activity: %w", err)
	}
	return nil
}

// GetActivityByStravaID returns (nil, nil) if not found.
func (db *DB) GetActivityByStravaID(stravaID int64) (*Activity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	a := &Activity{}
	err := db.conn.QueryRow(`
		SELECT id, strava_id, name, activity_type, start_date, distance,
		       duration_secs, avg_pace_secs, avg_hr, max_hr, avg_cadence, source, created_at
		FROM activities
		WHERE strava_id = ?`, stravaID).Scan(
		&a.ID,
		&a.StravaID,
		&a.Name,
		&a.ActivityType,
		&a.StartDate,
		&a.Distance,
		&a.DurationSecs,
		&a.AvgPaceSecs,
		&a.AvgHR,
		&a.MaxHR,
		&a.AvgCadence,
		&a.Source,
		&a.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get activity by strava id: %w", err)
	}
	return a, nil
}

func (db *DB) ListActivities(limit, offset int) ([]Activity, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	rows, err := db.conn.Query(`
		SELECT id, strava_id, name, activity_type, start_date, distance,
		       duration_secs, avg_pace_secs, avg_hr, max_hr, avg_cadence, source, created_at
		FROM activities
		ORDER BY start_date DESC
		LIMIT ? OFFSET ?`, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("list activities: %w", err)
	}
	defer rows.Close()

	var activities []Activity
	for rows.Next() {
		var a Activity
		if err := rows.Scan(
			&a.ID,
			&a.StravaID,
			&a.Name,
			&a.ActivityType,
			&a.StartDate,
			&a.Distance,
			&a.DurationSecs,
			&a.AvgPaceSecs,
			&a.AvgHR,
			&a.MaxHR,
			&a.AvgCadence,
			&a.Source,
			&a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan activity: %w", err)
		}
		activities = append(activities, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate activities: %w", err)
	}
	return activities, nil
}
