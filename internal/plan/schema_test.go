package plan

import (
	"testing"
	"time"
)

func TestValidateRace_Valid(t *testing.T) {
	r := &Race{
		Name:       "Spring Marathon",
		DistanceKm: 42.195,
		RaceDate:   time.Now().AddDate(0, 3, 0),
		Terrain:    TerrainRoad,
		Priority:   PriorityA,
	}
	if err := ValidateRace(r); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateRace_Nil(t *testing.T) {
	if err := ValidateRace(nil); err == nil {
		t.Error("expected error for nil race")
	}
}

func TestValidateRace_MissingName(t *testing.T) {
	r := &Race{
		DistanceKm: 10,
		RaceDate:   time.Now().AddDate(0, 1, 0),
		Terrain:    TerrainRoad,
		Priority:   PriorityA,
	}
	err := ValidateRace(r)
	if err == nil {
		t.Error("expected error for missing name")
	}
}

func TestValidateRace_ZeroDistance(t *testing.T) {
	r := &Race{
		Name:     "Test",
		RaceDate: time.Now().AddDate(0, 1, 0),
		Terrain:  TerrainRoad,
		Priority: PriorityA,
	}
	err := ValidateRace(r)
	if err == nil {
		t.Error("expected error for zero distance")
	}
}

func TestValidateRace_NegativeDistance(t *testing.T) {
	r := &Race{
		Name:       "Test",
		DistanceKm: -5,
		RaceDate:   time.Now().AddDate(0, 1, 0),
		Terrain:    TerrainRoad,
		Priority:   PriorityA,
	}
	err := ValidateRace(r)
	if err == nil {
		t.Error("expected error for negative distance")
	}
}

func TestValidateRace_ZeroDate(t *testing.T) {
	r := &Race{
		Name:       "Test",
		DistanceKm: 10,
		Terrain:    TerrainRoad,
		Priority:   PriorityA,
	}
	err := ValidateRace(r)
	if err == nil {
		t.Error("expected error for zero date")
	}
}

func TestValidateRace_InvalidTerrain(t *testing.T) {
	r := &Race{
		Name:       "Test",
		DistanceKm: 10,
		RaceDate:   time.Now().AddDate(0, 1, 0),
		Terrain:    "water",
		Priority:   PriorityA,
	}
	err := ValidateRace(r)
	if err == nil {
		t.Error("expected error for invalid terrain")
	}
}

func TestValidateRace_InvalidPriority(t *testing.T) {
	r := &Race{
		Name:       "Test",
		DistanceKm: 10,
		RaceDate:   time.Now().AddDate(0, 1, 0),
		Terrain:    TerrainRoad,
		Priority:   "Z",
	}
	err := ValidateRace(r)
	if err == nil {
		t.Error("expected error for invalid priority")
	}
}

func TestValidateRace_AllTerrains(t *testing.T) {
	for terrain := range validTerrains {
		r := &Race{
			Name:       "Test",
			DistanceKm: 10,
			RaceDate:   time.Now().AddDate(0, 1, 0),
			Terrain:    terrain,
			Priority:   PriorityA,
		}
		if err := ValidateRace(r); err != nil {
			t.Errorf("terrain %q should be valid, got: %v", terrain, err)
		}
	}
}

func TestValidateRace_AllPriorities(t *testing.T) {
	for prio := range validPriorities {
		r := &Race{
			Name:       "Test",
			DistanceKm: 10,
			RaceDate:   time.Now().AddDate(0, 1, 0),
			Terrain:    TerrainRoad,
			Priority:   prio,
		}
		if err := ValidateRace(r); err != nil {
			t.Errorf("priority %q should be valid, got: %v", prio, err)
		}
	}
}

func TestValidateRace_OptionalFields(t *testing.T) {
	elev := 500.0
	goal := 12600
	r := &Race{
		Name:        "Mountain Race",
		DistanceKm:  21.1,
		RaceDate:    time.Now().AddDate(0, 2, 0),
		Terrain:     TerrainTrail,
		ElevationM:  &elev,
		GoalTimeSec: &goal,
		Priority:    PriorityB,
	}
	if err := ValidateRace(r); err != nil {
		t.Errorf("expected no error with optional fields, got: %v", err)
	}
}

// --- ValidateSession tests ---

func TestValidateSession_Valid(t *testing.T) {
	s := &Session{
		Type:        SessionEasy,
		DayOfWeek:   1,
		DurationMin: 45,
		HRZone:      2,
	}
	if err := ValidateSession(s); err != nil {
		t.Errorf("expected no error, got: %v", err)
	}
}

func TestValidateSession_Nil(t *testing.T) {
	if err := ValidateSession(nil); err == nil {
		t.Error("expected error for nil session")
	}
}

func TestValidateSession_InvalidType(t *testing.T) {
	s := &Session{
		Type:        "sprint",
		DayOfWeek:   1,
		DurationMin: 30,
	}
	err := ValidateSession(s)
	if err == nil {
		t.Error("expected error for invalid session type")
	}
}

func TestValidateSession_DayOfWeekTooLow(t *testing.T) {
	s := &Session{
		Type:        SessionEasy,
		DayOfWeek:   0,
		DurationMin: 30,
	}
	err := ValidateSession(s)
	if err == nil {
		t.Error("expected error for day_of_week 0")
	}
}

func TestValidateSession_DayOfWeekTooHigh(t *testing.T) {
	s := &Session{
		Type:        SessionEasy,
		DayOfWeek:   8,
		DurationMin: 30,
	}
	err := ValidateSession(s)
	if err == nil {
		t.Error("expected error for day_of_week 8")
	}
}

func TestValidateSession_NegativeDuration(t *testing.T) {
	s := &Session{
		Type:        SessionEasy,
		DayOfWeek:   1,
		DurationMin: -10,
	}
	err := ValidateSession(s)
	if err == nil {
		t.Error("expected error for negative duration")
	}
}

func TestValidateSession_HRZoneTooHigh(t *testing.T) {
	s := &Session{
		Type:        SessionEasy,
		DayOfWeek:   1,
		DurationMin: 30,
		HRZone:      6,
	}
	err := ValidateSession(s)
	if err == nil {
		t.Error("expected error for hr_zone 6")
	}
}

func TestValidateSession_HRZoneNegative(t *testing.T) {
	s := &Session{
		Type:        SessionEasy,
		DayOfWeek:   1,
		DurationMin: 30,
		HRZone:      -1,
	}
	err := ValidateSession(s)
	if err == nil {
		t.Error("expected error for negative hr_zone")
	}
}

func TestValidateSession_ZeroDurationAllowed(t *testing.T) {
	s := &Session{
		Type:        SessionRest,
		DayOfWeek:   3,
		DurationMin: 0,
	}
	if err := ValidateSession(s); err != nil {
		t.Errorf("rest with zero duration should be valid, got: %v", err)
	}
}

func TestValidateSession_HRZoneZeroAllowed(t *testing.T) {
	s := &Session{
		Type:        SessionEasy,
		DayOfWeek:   1,
		DurationMin: 30,
		HRZone:      0,
	}
	if err := ValidateSession(s); err != nil {
		t.Errorf("hr_zone 0 should be valid (unspecified), got: %v", err)
	}
}

func TestValidateSession_AllTypes(t *testing.T) {
	for st := range validSessionTypes {
		s := &Session{
			Type:        st,
			DayOfWeek:   4,
			DurationMin: 30,
		}
		if err := ValidateSession(s); err != nil {
			t.Errorf("session type %q should be valid, got: %v", st, err)
		}
	}
}

func TestValidateSession_AllDays(t *testing.T) {
	for d := 1; d <= 7; d++ {
		s := &Session{
			Type:        SessionEasy,
			DayOfWeek:   d,
			DurationMin: 30,
		}
		if err := ValidateSession(s); err != nil {
			t.Errorf("day_of_week %d should be valid, got: %v", d, err)
		}
	}
}

func TestValidateSession_AllHRZones(t *testing.T) {
	for z := 0; z <= 5; z++ {
		s := &Session{
			Type:        SessionEasy,
			DayOfWeek:   1,
			DurationMin: 30,
			HRZone:      z,
		}
		if err := ValidateSession(s); err != nil {
			t.Errorf("hr_zone %d should be valid, got: %v", z, err)
		}
	}
}

// --- Enum map coverage ---

func TestSessionTypeMap(t *testing.T) {
	expected := []SessionType{
		SessionEasy, SessionTempo, SessionInterval, SessionLongRun,
		SessionStrength, SessionRest, SessionRace,
	}
	if len(validSessionTypes) != len(expected) {
		t.Errorf("validSessionTypes has %d entries, expected %d", len(validSessionTypes), len(expected))
	}
	for _, st := range expected {
		if !validSessionTypes[st] {
			t.Errorf("session type %q missing from validSessionTypes", st)
		}
	}
}

func TestSessionStatusMap(t *testing.T) {
	expected := []SessionStatus{StatusPlanned, StatusCompleted, StatusSkipped, StatusModified}
	if len(validSessionStatuses) != len(expected) {
		t.Errorf("validSessionStatuses has %d entries, expected %d", len(validSessionStatuses), len(expected))
	}
	for _, ss := range expected {
		if !validSessionStatuses[ss] {
			t.Errorf("session status %q missing from validSessionStatuses", ss)
		}
	}
}

func TestTerrainMap(t *testing.T) {
	expected := []Terrain{TerrainRoad, TerrainTrail, TerrainTrack}
	if len(validTerrains) != len(expected) {
		t.Errorf("validTerrains has %d entries, expected %d", len(validTerrains), len(expected))
	}
	for _, tr := range expected {
		if !validTerrains[tr] {
			t.Errorf("terrain %q missing from validTerrains", tr)
		}
	}
}

func TestPriorityMap(t *testing.T) {
	expected := []Priority{PriorityA, PriorityB, PriorityC}
	if len(validPriorities) != len(expected) {
		t.Errorf("validPriorities has %d entries, expected %d", len(validPriorities), len(expected))
	}
	for _, p := range expected {
		if !validPriorities[p] {
			t.Errorf("priority %q missing from validPriorities", p)
		}
	}
}
