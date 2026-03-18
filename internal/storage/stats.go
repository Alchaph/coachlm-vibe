package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

type AthleteStats struct {
	RecentRunCount      int
	RecentRunDistance   float64
	RecentRunMovingTime int
	RecentRunElevation  float64
	YTDRunCount         int
	YTDRunDistance      float64
	YTDRunMovingTime    int
	YTDRunElevation     float64
	AllRunCount         int
	AllRunDistance      float64
	AllRunMovingTime    int
	AllRunElevation     float64
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

func (db *DB) SaveAthleteStats(stats *AthleteStats) error {
	if stats == nil {
		return errors.New("stats is nil")
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO athlete_stats
			(id, recent_run_count, recent_run_distance, recent_run_moving_time, recent_run_elevation,
			 ytd_run_count, ytd_run_distance, ytd_run_moving_time, ytd_run_elevation,
			 all_run_count, all_run_distance, all_run_moving_time, all_run_elevation,
			 created_at, updated_at)
		VALUES
			(1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?,
			 COALESCE((SELECT created_at FROM athlete_stats WHERE id = 1), CURRENT_TIMESTAMP),
			 CURRENT_TIMESTAMP)`,
		stats.RecentRunCount,
		stats.RecentRunDistance,
		stats.RecentRunMovingTime,
		stats.RecentRunElevation,
		stats.YTDRunCount,
		stats.YTDRunDistance,
		stats.YTDRunMovingTime,
		stats.YTDRunElevation,
		stats.AllRunCount,
		stats.AllRunDistance,
		stats.AllRunMovingTime,
		stats.AllRunElevation,
	)
	if err != nil {
		return fmt.Errorf("save athlete stats: %w", err)
	}
	return nil
}

func (db *DB) GetAthleteStats() (*AthleteStats, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	s := &AthleteStats{}
	err := db.conn.QueryRow(`
		SELECT recent_run_count, recent_run_distance, recent_run_moving_time, recent_run_elevation,
		       ytd_run_count, ytd_run_distance, ytd_run_moving_time, ytd_run_elevation,
		       all_run_count, all_run_distance, all_run_moving_time, all_run_elevation,
		       created_at, updated_at
		FROM athlete_stats
		WHERE id = 1`).Scan(
		&s.RecentRunCount,
		&s.RecentRunDistance,
		&s.RecentRunMovingTime,
		&s.RecentRunElevation,
		&s.YTDRunCount,
		&s.YTDRunDistance,
		&s.YTDRunMovingTime,
		&s.YTDRunElevation,
		&s.AllRunCount,
		&s.AllRunDistance,
		&s.AllRunMovingTime,
		&s.AllRunElevation,
		&s.CreatedAt,
		&s.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get athlete stats: %w", err)
	}
	return s, nil
}
