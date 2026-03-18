package storage

import (
	"testing"
)

func validProfile() *AthleteProfile {
	return &AthleteProfile{
		Age:                 35,
		MaxHR:               185,
		ThresholdPaceSecs:   270,
		WeeklyMileageTarget: 50.0,
		RaceGoals:           "Sub-3 marathon",
		InjuryHistory:       "IT band 2024-06",
	}
}

func TestSaveAndGetProfile(t *testing.T) {
	db := newTestDB(t)
	p := validProfile()

	if err := db.SaveProfile(p); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	got, err := db.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if got == nil {
		t.Fatal("GetProfile returned nil after save")
	}

	if got.Age != p.Age {
		t.Errorf("Age = %d, want %d", got.Age, p.Age)
	}
	if got.MaxHR != p.MaxHR {
		t.Errorf("MaxHR = %d, want %d", got.MaxHR, p.MaxHR)
	}
	if got.ThresholdPaceSecs != p.ThresholdPaceSecs {
		t.Errorf("ThresholdPaceSecs = %d, want %d", got.ThresholdPaceSecs, p.ThresholdPaceSecs)
	}
	if got.WeeklyMileageTarget != p.WeeklyMileageTarget {
		t.Errorf("WeeklyMileageTarget = %f, want %f", got.WeeklyMileageTarget, p.WeeklyMileageTarget)
	}
	if got.RaceGoals != p.RaceGoals {
		t.Errorf("RaceGoals = %q, want %q", got.RaceGoals, p.RaceGoals)
	}
	if got.InjuryHistory != p.InjuryHistory {
		t.Errorf("InjuryHistory = %q, want %q", got.InjuryHistory, p.InjuryHistory)
	}
	if got.CreatedAt.IsZero() {
		t.Error("CreatedAt should not be zero")
	}
	if got.UpdatedAt.IsZero() {
		t.Error("UpdatedAt should not be zero")
	}
}

func TestGetProfileEmpty(t *testing.T) {
	db := newTestDB(t)

	got, err := db.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile on empty db: %v", err)
	}
	if got != nil {
		t.Errorf("expected nil profile on empty db, got %+v", got)
	}
}

func TestProfilePartialUpdate(t *testing.T) {
	db := newTestDB(t)
	p := validProfile()

	if err := db.SaveProfile(p); err != nil {
		t.Fatalf("first SaveProfile: %v", err)
	}

	first, err := db.GetProfile()
	if err != nil {
		t.Fatalf("first GetProfile: %v", err)
	}

	p.Age = 36
	p.RaceGoals = "Sub-2:55 marathon"
	if err := db.SaveProfile(p); err != nil {
		t.Fatalf("second SaveProfile: %v", err)
	}

	second, err := db.GetProfile()
	if err != nil {
		t.Fatalf("second GetProfile: %v", err)
	}

	if second.Age != 36 {
		t.Errorf("Age after update = %d, want 36", second.Age)
	}
	if second.RaceGoals != "Sub-2:55 marathon" {
		t.Errorf("RaceGoals after update = %q, want %q", second.RaceGoals, "Sub-2:55 marathon")
	}
	if second.MaxHR != first.MaxHR {
		t.Errorf("MaxHR changed unexpectedly: %d → %d", first.MaxHR, second.MaxHR)
	}
	if second.ThresholdPaceSecs != first.ThresholdPaceSecs {
		t.Errorf("ThresholdPaceSecs changed unexpectedly: %d → %d", first.ThresholdPaceSecs, second.ThresholdPaceSecs)
	}
}

func TestValidateProfileValid(t *testing.T) {
	db := newTestDB(t)
	if err := db.ValidateProfile(validProfile()); err != nil {
		t.Errorf("valid profile should pass validation: %v", err)
	}
}

func TestValidateProfileNil(t *testing.T) {
	db := newTestDB(t)
	if err := db.ValidateProfile(nil); err == nil {
		t.Error("nil profile should fail validation")
	}
}

func TestValidateProfileInvalidAge(t *testing.T) {
	db := newTestDB(t)

	tests := []struct {
		name string
		age  int
	}{
		{"negative", -5},
		{"zero", 0},
		{"too_high", 121},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validProfile()
			p.Age = tt.age
			if err := db.ValidateProfile(p); err == nil {
				t.Errorf("age %d should fail validation", tt.age)
			}
		})
	}
}

func TestValidateProfileInvalidHR(t *testing.T) {
	db := newTestDB(t)

	tests := []struct {
		name  string
		maxHR int
	}{
		{"too_low", 99},
		{"too_high", 221},
		{"zero", 0},
		{"negative", -10},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validProfile()
			p.MaxHR = tt.maxHR
			if err := db.ValidateProfile(p); err == nil {
				t.Errorf("maxHR %d should fail validation", tt.maxHR)
			}
		})
	}
}

func TestValidateProfileInvalidPace(t *testing.T) {
	db := newTestDB(t)

	tests := []struct {
		name string
		pace int
	}{
		{"zero", 0},
		{"negative", -100},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := validProfile()
			p.ThresholdPaceSecs = tt.pace
			if err := db.ValidateProfile(p); err == nil {
				t.Errorf("pace %d should fail validation", tt.pace)
			}
		})
	}
}

func TestValidateProfileBoundaryValues(t *testing.T) {
	db := newTestDB(t)

	tests := []struct {
		name    string
		profile *AthleteProfile
		valid   bool
	}{
		{"age_min", &AthleteProfile{Age: 1, MaxHR: 100, ThresholdPaceSecs: 1}, true},
		{"age_max", &AthleteProfile{Age: 120, MaxHR: 220, ThresholdPaceSecs: 1}, true},
		{"hr_min", &AthleteProfile{Age: 30, MaxHR: 100, ThresholdPaceSecs: 300}, true},
		{"hr_max", &AthleteProfile{Age: 30, MaxHR: 220, ThresholdPaceSecs: 300}, true},
		{"pace_min", &AthleteProfile{Age: 30, MaxHR: 180, ThresholdPaceSecs: 1}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := db.ValidateProfile(tt.profile)
			if tt.valid && err != nil {
				t.Errorf("expected valid, got error: %v", err)
			}
			if !tt.valid && err == nil {
				t.Error("expected invalid, got nil error")
			}
		})
	}
}

func TestProfileEmptyOptionalFields(t *testing.T) {
	db := newTestDB(t)
	p := &AthleteProfile{
		Age:               30,
		MaxHR:             180,
		ThresholdPaceSecs: 300,
		RaceGoals:         "",
		InjuryHistory:     "",
	}

	if err := db.SaveProfile(p); err != nil {
		t.Fatalf("SaveProfile with empty optionals: %v", err)
	}

	got, err := db.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if got.RaceGoals != "" {
		t.Errorf("RaceGoals = %q, want empty", got.RaceGoals)
	}
	if got.InjuryHistory != "" {
		t.Errorf("InjuryHistory = %q, want empty", got.InjuryHistory)
	}
}

func TestSaveProfileValidationRejectsInvalid(t *testing.T) {
	db := newTestDB(t)
	p := validProfile()
	p.Age = 0

	err := db.SaveProfile(p)
	if err == nil {
		t.Fatal("SaveProfile should reject invalid profile")
	}

	got, err := db.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile after rejected save: %v", err)
	}
	if got != nil {
		t.Error("no profile should exist after rejected save")
	}
}

func TestProfileHeartRateZones(t *testing.T) {
	db := newTestDB(t)
	p := validProfile()
	p.HeartRateZones = `[{"min":0,"max":115},{"min":115,"max":152},{"min":152,"max":171},{"min":171,"max":190},{"min":190,"max":-1}]`

	if err := db.SaveProfile(p); err != nil {
		t.Fatalf("SaveProfile with HR zones: %v", err)
	}

	got, err := db.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if got.HeartRateZones != p.HeartRateZones {
		t.Errorf("HeartRateZones = %q, want %q", got.HeartRateZones, p.HeartRateZones)
	}
}

func TestProfileHeartRateZonesPreservedOnUpdate(t *testing.T) {
	db := newTestDB(t)
	p := validProfile()
	p.HeartRateZones = `[{"min":0,"max":120}]`

	if err := db.SaveProfile(p); err != nil {
		t.Fatalf("first SaveProfile: %v", err)
	}

	p.Age = 36
	p.HeartRateZones = `[{"min":0,"max":120}]`
	if err := db.SaveProfile(p); err != nil {
		t.Fatalf("second SaveProfile: %v", err)
	}

	got, err := db.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if got.HeartRateZones != `[{"min":0,"max":120}]` {
		t.Errorf("HeartRateZones lost after update: %q", got.HeartRateZones)
	}
	if got.Age != 36 {
		t.Errorf("Age = %d, want 36", got.Age)
	}
}

func TestProfileHeartRateZonesEmpty(t *testing.T) {
	db := newTestDB(t)
	p := validProfile()

	if err := db.SaveProfile(p); err != nil {
		t.Fatalf("SaveProfile: %v", err)
	}

	got, err := db.GetProfile()
	if err != nil {
		t.Fatalf("GetProfile: %v", err)
	}
	if got.HeartRateZones != "" {
		t.Errorf("HeartRateZones = %q, want empty", got.HeartRateZones)
	}
}
