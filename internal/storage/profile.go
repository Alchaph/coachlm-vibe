package storage

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// AthleteProfile holds the runner's profile data used for coaching context.
// Only one profile row exists (singleton, id=1).
type AthleteProfile struct {
	Age                 int
	MaxHR               int
	ThresholdPaceSecs   int
	WeeklyMileageTarget float64
	RaceGoals           string
	InjuryHistory       string
	ExperienceLevel     string
	TrainingDaysPerWeek int
	RestingHR           int
	PreferredTerrain    string
	HeartRateZones      string
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// ValidateProfile checks that required numeric fields are within acceptable ranges.
func (db *DB) ValidateProfile(profile *AthleteProfile) error {
	if profile == nil {
		return errors.New("profile is nil")
	}
	if profile.Age < 1 || profile.Age > 120 {
		return fmt.Errorf("age must be between 1 and 120, got %d", profile.Age)
	}
	if profile.MaxHR < 100 || profile.MaxHR > 220 {
		return fmt.Errorf("max HR must be between 100 and 220, got %d", profile.MaxHR)
	}
	if profile.ThresholdPaceSecs <= 0 {
		return fmt.Errorf("threshold pace must be greater than 0, got %d", profile.ThresholdPaceSecs)
	}
	return nil
}

// SaveProfile upserts the athlete profile (INSERT OR REPLACE with id=1).
// It validates the profile before saving.
func (db *DB) SaveProfile(profile *AthleteProfile) error {
	if err := db.ValidateProfile(profile); err != nil {
		return fmt.Errorf("validate profile: %w", err)
	}

	db.mu.Lock()
	defer db.mu.Unlock()

	_, err := db.conn.Exec(`
		INSERT OR REPLACE INTO athlete_profile
			(id, age, max_hr, threshold_pace_secs, weekly_mileage_target, race_goals, injury_history, experience_level, training_days_per_week, resting_hr, preferred_terrain, heart_rate_zones, created_at, updated_at)
		VALUES
			(1, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, COALESCE((SELECT created_at FROM athlete_profile WHERE id = 1), CURRENT_TIMESTAMP), CURRENT_TIMESTAMP)`,
		profile.Age,
		profile.MaxHR,
		profile.ThresholdPaceSecs,
		profile.WeeklyMileageTarget,
		profile.RaceGoals,
		profile.InjuryHistory,
		profile.ExperienceLevel,
		profile.TrainingDaysPerWeek,
		profile.RestingHR,
		profile.PreferredTerrain,
		profile.HeartRateZones,
	)
	if err != nil {
		return fmt.Errorf("save profile: %w", err)
	}
	return nil
}

// GetProfile retrieves the athlete profile.
// Returns (nil, nil) if no profile exists yet.
func (db *DB) GetProfile() (*AthleteProfile, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	p := &AthleteProfile{}
	err := db.conn.QueryRow(`
		SELECT age, max_hr, threshold_pace_secs, weekly_mileage_target,
		       race_goals, injury_history, experience_level, training_days_per_week,
		       resting_hr, preferred_terrain, heart_rate_zones, created_at, updated_at
		FROM athlete_profile
		WHERE id = 1`).Scan(
		&p.Age,
		&p.MaxHR,
		&p.ThresholdPaceSecs,
		&p.WeeklyMileageTarget,
		&p.RaceGoals,
		&p.InjuryHistory,
		&p.ExperienceLevel,
		&p.TrainingDaysPerWeek,
		&p.RestingHR,
		&p.PreferredTerrain,
		&p.HeartRateZones,
		&p.CreatedAt,
		&p.UpdatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get profile: %w", err)
	}
	return p, nil
}
