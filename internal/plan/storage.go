package plan

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"coachlm/internal/storage"
)

// Storage wraps the database handle and provides plan-specific CRUD.
type Storage struct {
	db *storage.DB
}

// NewStorage creates a plan storage layer backed by the given DB.
func NewStorage(db *storage.DB) *Storage {
	return &Storage{db: db}
}

// --- Race CRUD ---

// CreateRace inserts a race. The caller must set r.ID before calling.
func (s *Storage) CreateRace(r *Race) error {
	if err := ValidateRace(r); err != nil {
		return fmt.Errorf("validate race: %w", err)
	}
	if r.RaceDate.Before(time.Now().Truncate(24 * time.Hour)) {
		return errors.New("race date cannot be in the past")
	}

	conn := s.db.Conn()
	_, err := conn.Exec(`
		INSERT INTO races
			(id, name, distance_km, race_date, terrain, elevation_m, goal_time_s, priority, is_active, created_at)
		VALUES
			(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		r.ID, r.Name, r.DistanceKm, r.RaceDate, r.Terrain,
		r.ElevationM, r.GoalTimeSec, r.Priority, r.IsActive,
		time.Now(),
	)
	if err != nil {
		return fmt.Errorf("create race: %w", err)
	}
	return nil
}

// UpdateRace updates an existing race by ID.
func (s *Storage) UpdateRace(r *Race) error {
	if err := ValidateRace(r); err != nil {
		return fmt.Errorf("validate race: %w", err)
	}
	if r.ID == "" {
		return errors.New("race ID is required")
	}

	conn := s.db.Conn()
	result, err := conn.Exec(`
		UPDATE races SET
			name = ?, distance_km = ?, race_date = ?, terrain = ?,
			elevation_m = ?, goal_time_s = ?, priority = ?, is_active = ?
		WHERE id = ?`,
		r.Name, r.DistanceKm, r.RaceDate, r.Terrain,
		r.ElevationM, r.GoalTimeSec, r.Priority, r.IsActive,
		r.ID,
	)
	if err != nil {
		return fmt.Errorf("update race: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("race not found: %s", r.ID)
	}
	return nil
}

// DeleteRace removes a race and its plans (cascade).
func (s *Storage) DeleteRace(id string) error {
	if id == "" {
		return errors.New("race ID is required")
	}
	conn := s.db.Conn()
	result, err := conn.Exec(`DELETE FROM races WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete race: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("race not found: %s", id)
	}
	return nil
}

// ListRaces returns all races ordered by date ascending.
func (s *Storage) ListRaces() ([]Race, error) {
	conn := s.db.Conn()
	rows, err := conn.Query(`
		SELECT id, name, distance_km, race_date, terrain, elevation_m,
		       goal_time_s, priority, is_active, created_at
		FROM races
		ORDER BY race_date ASC`)
	if err != nil {
		return nil, fmt.Errorf("list races: %w", err)
	}
	defer rows.Close()

	var races []Race
	for rows.Next() {
		var r Race
		var elevM sql.NullFloat64
		var goalS sql.NullInt64
		if err := rows.Scan(
			&r.ID, &r.Name, &r.DistanceKm, &r.RaceDate, &r.Terrain,
			&elevM, &goalS, &r.Priority, &r.IsActive, &r.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan race: %w", err)
		}
		if elevM.Valid {
			v := elevM.Float64
			r.ElevationM = &v
		}
		if goalS.Valid {
			v := int(goalS.Int64)
			r.GoalTimeSec = &v
		}
		races = append(races, r)
	}
	return races, rows.Err()
}

// GetRace returns a single race by ID, or (nil, nil) if not found.
func (s *Storage) GetRace(id string) (*Race, error) {
	conn := s.db.Conn()
	var r Race
	var elevM sql.NullFloat64
	var goalS sql.NullInt64
	err := conn.QueryRow(`
		SELECT id, name, distance_km, race_date, terrain, elevation_m,
		       goal_time_s, priority, is_active, created_at
		FROM races WHERE id = ?`, id).Scan(
		&r.ID, &r.Name, &r.DistanceKm, &r.RaceDate, &r.Terrain,
		&elevM, &goalS, &r.Priority, &r.IsActive, &r.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get race: %w", err)
	}
	if elevM.Valid {
		v := elevM.Float64
		r.ElevationM = &v
	}
	if goalS.Valid {
		v := int(goalS.Int64)
		r.GoalTimeSec = &v
	}
	return &r, nil
}

// SetActiveRace makes exactly one race active, deactivating all others.
func (s *Storage) SetActiveRace(id string) error {
	if id == "" {
		return errors.New("race ID is required")
	}
	conn := s.db.Conn()
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	if _, err := tx.Exec(`UPDATE races SET is_active = 0`); err != nil {
		return fmt.Errorf("deactivate races: %w", err)
	}
	result, err := tx.Exec(`UPDATE races SET is_active = 1 WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("activate race: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("race not found: %s", id)
	}
	return tx.Commit()
}

// GetActiveRace returns the currently active race, or (nil, nil) if none.
func (s *Storage) GetActiveRace() (*Race, error) {
	conn := s.db.Conn()
	var r Race
	var elevM sql.NullFloat64
	var goalS sql.NullInt64
	err := conn.QueryRow(`
		SELECT id, name, distance_km, race_date, terrain, elevation_m,
		       goal_time_s, priority, is_active, created_at
		FROM races WHERE is_active = 1`).Scan(
		&r.ID, &r.Name, &r.DistanceKm, &r.RaceDate, &r.Terrain,
		&elevM, &goalS, &r.Priority, &r.IsActive, &r.CreatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active race: %w", err)
	}
	if elevM.Valid {
		v := elevM.Float64
		r.ElevationM = &v
	}
	if goalS.Valid {
		v := int(goalS.Int64)
		r.GoalTimeSec = &v
	}
	return &r, nil
}

// --- Plan CRUD ---

// SavePlan inserts a plan with its weeks and sessions inside a transaction.
func (s *Storage) SavePlan(p *TrainingPlan) error {
	if p == nil {
		return errors.New("plan is nil")
	}
	if p.RaceID == "" {
		return errors.New("plan race_id is required")
	}

	conn := s.db.Conn()
	tx, err := conn.Begin()
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback()

	// Archive existing plans for this race (set generated_at to mark them as old,
	// but keep in DB for history).
	if _, err := tx.Exec(`
		INSERT INTO training_plans (id, race_id, generated_at, llm_backend, prompt_hash)
		VALUES (?, ?, ?, ?, ?)`,
		p.ID, p.RaceID, p.GeneratedAt, p.LLMBackend, p.PromptHash,
	); err != nil {
		return fmt.Errorf("insert plan: %w", err)
	}

	for _, w := range p.Weeks {
		if _, err := tx.Exec(`
			INSERT INTO plan_weeks (id, plan_id, week_number, week_start)
			VALUES (?, ?, ?, ?)`,
			w.ID, p.ID, w.WeekNumber, w.WeekStart,
		); err != nil {
			return fmt.Errorf("insert week %d: %w", w.WeekNumber, err)
		}

		for _, sess := range w.Sessions {
			if _, err := tx.Exec(`
				INSERT INTO plan_sessions
					(id, week_id, day_of_week, session_type, duration_min, distance_km,
					 hr_zone, pace_min_low, pace_min_high, notes, status,
					 actual_duration_min, actual_distance_km, completed_at)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
				sess.ID, w.ID, sess.DayOfWeek, sess.Type, sess.DurationMin,
				sess.DistanceKm, sess.HRZone, sess.PaceMinLow, sess.PaceMinHigh,
				sess.Notes, StatusPlanned, nil, nil, nil,
			); err != nil {
				return fmt.Errorf("insert session day %d week %d: %w", sess.DayOfWeek, w.WeekNumber, err)
			}
		}
	}

	return tx.Commit()
}

// GetActivePlan returns the latest plan for the active race, or (nil, nil).
func (s *Storage) GetActivePlan() (*TrainingPlan, error) {
	race, err := s.GetActiveRace()
	if err != nil || race == nil {
		return nil, err
	}

	conn := s.db.Conn()
	var p TrainingPlan
	err = conn.QueryRow(`
		SELECT id, race_id, generated_at, llm_backend, prompt_hash
		FROM training_plans
		WHERE race_id = ?
		ORDER BY generated_at DESC
		LIMIT 1`, race.ID).Scan(
		&p.ID, &p.RaceID, &p.GeneratedAt, &p.LLMBackend, &p.PromptHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get active plan: %w", err)
	}
	return &p, nil
}

// GetPlanWeeks loads all weeks (with sessions) for a plan.
func (s *Storage) GetPlanWeeks(planID string) ([]Week, error) {
	if planID == "" {
		return nil, errors.New("plan ID is required")
	}

	conn := s.db.Conn()
	weekRows, err := conn.Query(`
		SELECT id, plan_id, week_number, week_start
		FROM plan_weeks
		WHERE plan_id = ?
		ORDER BY week_number ASC`, planID)
	if err != nil {
		return nil, fmt.Errorf("list weeks: %w", err)
	}
	defer weekRows.Close()

	var weeks []Week
	for weekRows.Next() {
		var w Week
		if err := weekRows.Scan(&w.ID, &w.PlanID, &w.WeekNumber, &w.WeekStart); err != nil {
			return nil, fmt.Errorf("scan week: %w", err)
		}
		weeks = append(weeks, w)
	}
	if err := weekRows.Err(); err != nil {
		return nil, fmt.Errorf("iterate weeks: %w", err)
	}

	// Load sessions for each week.
	for i := range weeks {
		sessRows, err := conn.Query(`
			SELECT id, week_id, day_of_week, session_type, duration_min, distance_km,
			       hr_zone, pace_min_low, pace_min_high, notes, status,
			       actual_duration_min, actual_distance_km, completed_at
			FROM plan_sessions
			WHERE week_id = ?
			ORDER BY day_of_week ASC`, weeks[i].ID)
		if err != nil {
			return nil, fmt.Errorf("list sessions for week %d: %w", weeks[i].WeekNumber, err)
		}

		for sessRows.Next() {
			var sess Session
			var actualDur sql.NullInt64
			var actualDist sql.NullFloat64
			var completedAt sql.NullTime
			if err := sessRows.Scan(
				&sess.ID, &sess.WeekID, &sess.DayOfWeek, &sess.Type,
				&sess.DurationMin, &sess.DistanceKm, &sess.HRZone,
				&sess.PaceMinLow, &sess.PaceMinHigh, &sess.Notes, &sess.Status,
				&actualDur, &actualDist, &completedAt,
			); err != nil {
				sessRows.Close()
				return nil, fmt.Errorf("scan session: %w", err)
			}
			if actualDur.Valid {
				v := int(actualDur.Int64)
				sess.ActualDurationMin = &v
			}
			if actualDist.Valid {
				sess.ActualDistanceKm = &actualDist.Float64
			}
			if completedAt.Valid {
				sess.CompletedAt = &completedAt.Time
			}
			weeks[i].Sessions = append(weeks[i].Sessions, sess)
		}
		sessRows.Close()
	}

	return weeks, nil
}

// UpdateSessionStatus marks a session with the given status and optional actual metrics.
func (s *Storage) UpdateSessionStatus(sessionID string, status SessionStatus, actual ActualMetrics) error {
	if sessionID == "" {
		return errors.New("session ID is required")
	}
	if !validSessionStatuses[status] {
		return fmt.Errorf("invalid session status: %q", status)
	}

	conn := s.db.Conn()
	var completedAt interface{}
	if status == StatusCompleted || status == StatusModified {
		now := time.Now()
		completedAt = now
	}

	result, err := conn.Exec(`
		UPDATE plan_sessions SET
			status = ?, actual_duration_min = ?, actual_distance_km = ?, completed_at = ?
		WHERE id = ?`,
		status, actual.DurationMin, actual.DistanceKm, completedAt, sessionID,
	)
	if err != nil {
		return fmt.Errorf("update session status: %w", err)
	}
	rows, _ := result.RowsAffected()
	if rows == 0 {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	return nil
}

// GetLatestPlanForRace returns the most recent plan for a given race ID.
func (s *Storage) GetLatestPlanForRace(raceID string) (*TrainingPlan, error) {
	conn := s.db.Conn()
	var p TrainingPlan
	err := conn.QueryRow(`
		SELECT id, race_id, generated_at, llm_backend, prompt_hash
		FROM training_plans
		WHERE race_id = ?
		ORDER BY generated_at DESC
		LIMIT 1`, raceID).Scan(
		&p.ID, &p.RaceID, &p.GeneratedAt, &p.LLMBackend, &p.PromptHash,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get latest plan for race: %w", err)
	}
	return &p, nil
}
