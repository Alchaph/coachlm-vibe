// Package plan implements training plan generation, storage, and context assembly.
package plan

import (
	"errors"
	"fmt"
	"time"
)

// SessionType enumerates the kinds of training sessions in a plan.
type SessionType string

const (
	SessionEasy     SessionType = "easy"
	SessionTempo    SessionType = "tempo"
	SessionInterval SessionType = "intervals"
	SessionLongRun  SessionType = "long_run"
	SessionStrength SessionType = "strength"
	SessionRest     SessionType = "rest"
	SessionRace     SessionType = "race"
)

var validSessionTypes = map[SessionType]bool{
	SessionEasy:     true,
	SessionTempo:    true,
	SessionInterval: true,
	SessionLongRun:  true,
	SessionStrength: true,
	SessionRest:     true,
	SessionRace:     true,
}

// SessionStatus tracks whether a planned session was executed.
type SessionStatus string

const (
	StatusPlanned   SessionStatus = "planned"
	StatusCompleted SessionStatus = "completed"
	StatusSkipped   SessionStatus = "skipped"
	StatusModified  SessionStatus = "modified"
)

var validSessionStatuses = map[SessionStatus]bool{
	StatusPlanned:   true,
	StatusCompleted: true,
	StatusSkipped:   true,
	StatusModified:  true,
}

// Terrain describes the surface a race takes place on.
type Terrain string

const (
	TerrainRoad  Terrain = "road"
	TerrainTrail Terrain = "trail"
	TerrainTrack Terrain = "track"
)

var validTerrains = map[Terrain]bool{
	TerrainRoad:  true,
	TerrainTrail: true,
	TerrainTrack: true,
}

// Priority indicates how important a race is within a training cycle.
type Priority string

const (
	PriorityA Priority = "A"
	PriorityB Priority = "B"
	PriorityC Priority = "C"
)

var validPriorities = map[Priority]bool{
	PriorityA: true,
	PriorityB: true,
	PriorityC: true,
}

// Race represents a goal race the athlete is training for.
type Race struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	DistanceKm  float64   `json:"distanceKm"`
	RaceDate    time.Time `json:"raceDate"`
	Terrain     Terrain   `json:"terrain"`
	ElevationM  *float64  `json:"elevationM,omitempty"`
	GoalTimeSec *int      `json:"goalTimeSec,omitempty"`
	Priority    Priority  `json:"priority"`
	IsActive    bool      `json:"isActive"`
	CreatedAt   time.Time `json:"createdAt"`
}

// ValidateRace checks required fields and business rules.
func ValidateRace(r *Race) error {
	if r == nil {
		return errors.New("race is nil")
	}
	if r.Name == "" {
		return errors.New("race name is required")
	}
	if r.DistanceKm <= 0 {
		return errors.New("distance must be greater than zero")
	}
	if r.RaceDate.IsZero() {
		return errors.New("race date is required")
	}
	if !validTerrains[r.Terrain] {
		return fmt.Errorf("invalid terrain: %q", r.Terrain)
	}
	if !validPriorities[r.Priority] {
		return fmt.Errorf("invalid priority: %q", r.Priority)
	}
	return nil
}

// TrainingPlan holds a generated plan linked to a specific race.
type TrainingPlan struct {
	ID          string    `json:"id"`
	RaceID      string    `json:"raceId"`
	GeneratedAt time.Time `json:"generatedAt"`
	LLMBackend  string    `json:"llmBackend"`
	PromptHash  string    `json:"promptHash"`
	Weeks       []Week    `json:"weeks,omitempty"`
}

// Week represents one training week within a plan.
type Week struct {
	ID         string    `json:"id"`
	PlanID     string    `json:"planId"`
	WeekNumber int       `json:"weekNumber"`
	WeekStart  time.Time `json:"weekStart"`
	Sessions   []Session `json:"sessions"`
}

// Session is a single workout day in a training plan.
type Session struct {
	ID                string        `json:"id"`
	WeekID            string        `json:"weekId"`
	DayOfWeek         int           `json:"dayOfWeek"` // 1=Mon, 7=Sun
	Type              SessionType   `json:"type"`
	DurationMin       int           `json:"durationMin"`
	DistanceKm        float64       `json:"distanceKm,omitempty"`
	HRZone            int           `json:"hrZone,omitempty"` // 1-5
	PaceMinLow        float64       `json:"paceMinLow,omitempty"`
	PaceMinHigh       float64       `json:"paceMinHigh,omitempty"`
	Notes             string        `json:"notes"`
	Status            SessionStatus `json:"status"`
	ActualDurationMin *int          `json:"actualDurationMin,omitempty"`
	ActualDistanceKm  *float64      `json:"actualDistanceKm,omitempty"`
	CompletedAt       *time.Time    `json:"completedAt,omitempty"`
}

// ActualMetrics holds what the athlete actually did for a session.
type ActualMetrics struct {
	DurationMin *int     `json:"durationMin,omitempty"`
	DistanceKm  *float64 `json:"distanceKm,omitempty"`
}

// ValidateSession checks required fields on a session.
func ValidateSession(s *Session) error {
	if s == nil {
		return errors.New("session is nil")
	}
	if !validSessionTypes[s.Type] {
		return fmt.Errorf("invalid session type: %q", s.Type)
	}
	if s.DayOfWeek < 1 || s.DayOfWeek > 7 {
		return fmt.Errorf("day_of_week must be 1-7, got %d", s.DayOfWeek)
	}
	if s.DurationMin < 0 {
		return fmt.Errorf("duration must be non-negative, got %d", s.DurationMin)
	}
	if s.HRZone < 0 || s.HRZone > 5 {
		return fmt.Errorf("hr_zone must be 0-5, got %d", s.HRZone)
	}
	return nil
}

// LLMPlanResponse is the shape we instruct the LLM to return.
// Used for JSON unmarshalling of the generated plan.
type LLMPlanResponse struct {
	Weeks []LLMWeek `json:"weeks"`
}

// LLMWeek is a week in the LLM-generated JSON.
type LLMWeek struct {
	WeekNumber int          `json:"week_number"`
	Sessions   []LLMSession `json:"sessions"`
}

// LLMSession is a session in the LLM-generated JSON.
type LLMSession struct {
	DayOfWeek   int     `json:"day_of_week"`
	Type        string  `json:"type"`
	DurationMin int     `json:"duration_min"`
	DistanceKm  float64 `json:"distance_km,omitempty"`
	HRZone      int     `json:"hr_zone,omitempty"`
	PaceLow     float64 `json:"pace_low,omitempty"`
	PaceHigh    float64 `json:"pace_high,omitempty"`
	Notes       string  `json:"notes"`
}
